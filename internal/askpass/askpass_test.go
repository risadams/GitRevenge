package askpass

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

func TestClassify(t *testing.T) {
	cases := map[string]Kind{
		"Username for 'https://github.com': ":                                   Text,
		"Password for 'https://user@github.com': ":                              Secret,
		"Enter passphrase for key '/home/u/.ssh/id_ed25519': ":                  Secret,
		"Are you sure you want to continue connecting (yes/no/[fingerprint])? ": HostKey,
		"something unexpected":                                                  Secret,
	}
	for prompt, want := range cases {
		if got := Classify(prompt); got != want {
			t.Errorf("Classify(%q) = %s, want %s", prompt, got, want)
		}
	}
}

func TestClientServerRoundTrip(t *testing.T) {
	var got []string
	var mu sync.Mutex
	s, err := Listen(func(_ context.Context, prompt string, kind Kind) (string, bool) {
		mu.Lock()
		got = append(got, string(kind)+":"+prompt)
		mu.Unlock()
		if strings.Contains(prompt, "cancel") {
			return "", false
		}
		return "hunter2", true
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	env := map[string]string{}
	for _, kv := range s.Env("exe") {
		k, v, _ := strings.Cut(kv, "=")
		env[k] = v
	}

	ans, err := Client(env[EnvAddr], env[EnvToken], "Password: ")
	if err != nil || ans != "hunter2" {
		t.Fatalf("answer = %q, %v", ans, err)
	}
	if _, err := Client(env[EnvAddr], env[EnvToken], "please cancel"); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel err = %v", err)
	}
	if _, err := Client(env[EnvAddr], "wrong-token", "Password: "); err == nil {
		t.Fatal("a wrong token must not get an answer")
	}
	mu.Lock()
	defer mu.Unlock()
	if len(got) != 2 || got[0] != "secret:Password: " {
		t.Fatalf("handler saw %v", got)
	}
}

// End to end: git asks for HTTPS credentials through GIT_ASKPASS, our helper relays them to
// the server, and git receives the answers.
func TestGitUsesHelper(t *testing.T) {
	gitBin, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not installed")
	}
	helper := filepath.Join(t.TempDir(), "helper")
	if runtime.GOOS == "windows" {
		helper += ".exe"
	}
	if out, err := exec.Command("go", "build", "-o", helper, "./testhelper").CombinedOutput(); err != nil {
		t.Fatalf("building helper: %v\n%s", err, out)
	}

	// Generated per run so the source holds no credential-looking literals.
	user, pass := "user-"+randomHex(t), "pass-"+randomHex(t)
	s, err := Listen(func(_ context.Context, _ string, kind Kind) (string, bool) {
		if kind == Text {
			return user, true
		}
		return pass, true
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()

	// git credential fill prompts for anything the input doesn't provide.
	cmd := exec.Command(gitBin, "-c", "credential.helper=", "credential", "fill")
	cmd.Stdin = strings.NewReader("protocol=https\nhost=example.com\n\n")
	global := filepath.Join(t.TempDir(), "gitconfig")
	_ = os.WriteFile(global, nil, 0o600)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL="+global, "GIT_CONFIG_NOSYSTEM=1", "GIT_TERMINAL_PROMPT=0")
	cmd.Env = append(cmd.Env, s.Env(helper)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("credential fill: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "username="+user) || !strings.Contains(string(out), "password="+pass) {
		t.Fatalf("git got: %s", out)
	}
}

func randomHex(t *testing.T) string {
	t.Helper()
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(b)
}
