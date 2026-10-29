// Package askpass relays git and ssh credential prompts to the running app (ADR-0004).
//
// Network commands run with GIT_ASKPASS and SSH_ASKPASS pointing at the gitrevenge
// executable. Started that way, the executable runs Client: it sends the prompt to the
// app's Server over a loopback connection, prints the answer for git, and exits. The
// server is protected by a random token passed only in the child's environment.
package askpass

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// Environment variables that put the executable into helper mode.
const (
	EnvAddr  = "GITREVENGE_ASKPASS_ADDR"
	EnvToken = "GITREVENGE_ASKPASS_TOKEN"
)

// Kind classifies a prompt so the UI can show the right dialog.
type Kind string

// Prompt kinds.
const (
	// Text is a visible answer, such as a username.
	Text Kind = "text"
	// Secret is a hidden answer: a password, token or key passphrase.
	Secret Kind = "secret"
	// HostKey asks to trust an SSH server's key; the answer is "yes".
	HostKey Kind = "hostkey"
)

// Classify decides how to present a prompt from git or ssh.
func Classify(prompt string) Kind {
	p := strings.ToLower(prompt)
	switch {
	case strings.Contains(p, "continue connecting"):
		return HostKey
	case strings.HasPrefix(p, "username"):
		return Text
	default:
		// Passwords, passphrases, PINs and anything unrecognised are hidden.
		return Secret
	}
}

// Handler answers a prompt. ok=false means the user cancelled.
type Handler func(ctx context.Context, prompt string, kind Kind) (answer string, ok bool)

// Server accepts prompts from helper processes.
type Server struct {
	ln      net.Listener
	token   string
	handler Handler
	wg      sync.WaitGroup
}

// Listen starts a server on a random loopback port.
func Listen(handler Handler) (*Server, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	tok := make([]byte, 32)
	if _, err := rand.Read(tok); err != nil {
		_ = ln.Close()
		return nil, err
	}
	s := &Server{ln: ln, token: hex.EncodeToString(tok), handler: handler}
	s.wg.Go(s.serve)
	return s, nil
}

// Env returns the environment for a git command whose prompts should reach this server.
// exe is the path of the gitrevenge executable.
func (s *Server) Env(exe string) []string {
	return []string{
		"GIT_ASKPASS=" + exe,
		"SSH_ASKPASS=" + exe,
		// OpenSSH only uses SSH_ASKPASS without a terminal when told to, and older
		// versions also want DISPLAY set.
		"SSH_ASKPASS_REQUIRE=force",
		"DISPLAY=:0",
		EnvAddr + "=" + s.ln.Addr().String(),
		EnvToken + "=" + s.token,
	}
}

// Close stops accepting prompts and waits for in-flight ones to finish.
func (s *Server) Close() error {
	err := s.ln.Close()
	s.wg.Wait()
	return err
}

type request struct {
	Token  string `json:"token"`
	Prompt string `json:"prompt"`
}

type response struct {
	Answer string `json:"answer"`
	OK     bool   `json:"ok"`
}

func (s *Server) serve() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return
		}
		s.wg.Go(func() { s.handle(conn) })
	}
}

// promptTimeout bounds how long a prompt waits for the user before git is told no.
const promptTimeout = 5 * time.Minute

func (s *Server) handle(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	var req request
	if err := json.NewDecoder(bufio.NewReader(io.LimitReader(conn, 64<<10))).Decode(&req); err != nil {
		return
	}
	if subtle.ConstantTimeCompare([]byte(req.Token), []byte(s.token)) != 1 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), promptTimeout)
	defer cancel()
	answer, ok := s.handler(ctx, req.Prompt, Classify(req.Prompt))
	_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_ = json.NewEncoder(conn).Encode(response{Answer: answer, OK: ok})
}

// ErrCancelled means the user declined to answer.
var ErrCancelled = errors.New("prompt cancelled")

// IsHelper reports whether this process was started by git or ssh as the askpass helper.
func IsHelper() bool { return os.Getenv(EnvAddr) != "" && os.Getenv(EnvToken) != "" }

// Client asks the app to answer prompt and returns the answer.
func Client(addr, token, prompt string) (string, error) {
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return "", err
	}
	defer func() { _ = conn.Close() }()
	if err := json.NewEncoder(conn).Encode(request{Token: token, Prompt: prompt}); err != nil {
		return "", err
	}
	_ = conn.SetReadDeadline(time.Now().Add(promptTimeout + 30*time.Second))
	var res response
	if err := json.NewDecoder(conn).Decode(&res); err != nil {
		return "", fmt.Errorf("no answer from Git Revenge: %w", err)
	}
	if !res.OK {
		return "", ErrCancelled
	}
	return res.Answer, nil
}

// RunHelper is the askpass helper's main: it relays os.Args[1] and prints the answer for
// git or ssh. It returns the process exit code; non-zero tells git the prompt failed.
func RunHelper(stdout, stderr io.Writer) int {
	prompt := ""
	if len(os.Args) > 1 {
		prompt = os.Args[1]
	}
	answer, err := Client(os.Getenv(EnvAddr), os.Getenv(EnvToken), prompt)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "gitrevenge askpass:", err)
		return 1
	}
	_, _ = fmt.Fprintln(stdout, answer)
	return 0
}
