package git

import (
	"regexp"
	"strings"
)

// Trailer is a "Key: value" line at the end of a commit message, as git interpret-trailers
// reads it (Co-authored-by, Signed-off-by and the like), or a closing reference such as
// "Closes #12", kept as Key "Closes" and Value "#12".
type Trailer struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

var (
	trailerLine = regexp.MustCompile(`^([A-Za-z0-9][A-Za-z0-9-]*)[ \t]*:[ \t]*(\S.*?)[ \t]*$`)
	// GitHub's and GitLab's closing keywords, which are written without a colon.
	closingLine = regexp.MustCompile(`(?i)^(close[sd]?|fix(?:e[sd])?|resolve[sd]?)[ \t]+(#\d+(?:[ \t]*,[ \t]*#\d+)*)[ \t]*$`)
	paragraphs  = regexp.MustCompile(`\n[ \t]*\n`)
)

// SplitTrailers separates a message body's trailers from the rest of it. As in git, the
// trailers are the last paragraph when every line in it is a trailer or an indented
// continuation of one. Closing references count as trailers too, and so does a paragraph
// of them just before the trailers, as in "Closes #12", blank line, "Co-authored-by: …".
func SplitTrailers(body string) (rest string, trailers []Trailer) {
	body = strings.TrimRight(strings.ReplaceAll(body, "\r\n", "\n"), " \t\n")
	paras := paragraphs.Split(body, -1)
	keep := len(paras)
	for keep > 0 {
		t := parseTrailerParagraph(paras[keep-1])
		if t == nil {
			break
		}
		trailers = append(t, trailers...)
		keep--
	}
	if len(trailers) == 0 {
		return body, nil
	}
	return strings.TrimSpace(strings.Join(paras[:keep], "\n\n")), trailers
}

// parseTrailerParagraph returns a paragraph's trailers, or nil if any line isn't one.
func parseTrailerParagraph(para string) []Trailer {
	var out []Trailer
	for line := range strings.SplitSeq(para, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if (line[0] == ' ' || line[0] == '\t') && len(out) > 0 {
			out[len(out)-1].Value += " " + strings.TrimSpace(line)
			continue
		}
		if m := closingLine.FindStringSubmatch(line); m != nil {
			out = append(out, Trailer{Key: m[1], Value: m[2]})
			continue
		}
		m := trailerLine.FindStringSubmatch(line)
		// "https://…" on a line of its own looks like a trailer keyed "https"; it isn't.
		if m == nil || strings.HasPrefix(m[2], "//") {
			return nil
		}
		out = append(out, Trailer{Key: m[1], Value: m[2]})
	}
	return out
}
