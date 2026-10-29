package git

import (
	"reflect"
	"testing"
)

func TestSplitTrailers(t *testing.T) {
	cases := []struct {
		name     string
		body     string
		rest     string
		trailers []Trailer
	}{
		{
			name:     "after a paragraph",
			body:     "Explains the change.\n\nCloses #12\nCo-Authored-By: Ada <ada@example.com>",
			rest:     "Explains the change.",
			trailers: []Trailer{{"Closes", "#12"}, {"Co-Authored-By", "Ada <ada@example.com>"}},
		},
		{
			name:     "the whole body",
			body:     "Signed-off-by: Linus <torvalds@example.org>\n",
			rest:     "",
			trailers: []Trailer{{"Signed-off-by", "Linus <torvalds@example.org>"}},
		},
		{
			name:     "continuation lines and CRLF",
			body:     "Body.\r\n\r\nNote: a long value\r\n  that wraps\r\nAcked-by: Bo <bo@example.com>",
			rest:     "Body.",
			trailers: []Trailer{{"Note", "a long value that wraps"}, {"Acked-by", "Bo <bo@example.com>"}},
		},
		{
			name:     "a closing paragraph before the trailers, as this repo writes them",
			body:     "Explains it.\n\nCloses #171\n\nCo-Authored-By: Claude <noreply@anthropic.com>",
			rest:     "Explains it.",
			trailers: []Trailer{{"Closes", "#171"}, {"Co-Authored-By", "Claude <noreply@anthropic.com>"}},
		},
		{
			name:     "several closing references",
			body:     "fixes #1, #2",
			rest:     "",
			trailers: []Trailer{{"fixes", "#1, #2"}},
		},
		{
			name: "a last paragraph that isn't all trailers",
			body: "Intro.\n\nSee: the docs\nand then some prose",
			rest: "Intro.\n\nSee: the docs\nand then some prose",
		},
		{
			name: "a closing keyword inside prose",
			body: "This closes #4 by accident.",
			rest: "This closes #4 by accident.",
		},
		{
			name: "a URL on its own line",
			body: "Intro.\n\nhttps://example.com/issue/1",
			rest: "Intro.\n\nhttps://example.com/issue/1",
		},
		{
			name: "trailer-like text before the last paragraph",
			body: "Fixes: something\n\nThen prose at the end.",
			rest: "Fixes: something\n\nThen prose at the end.",
		},
		{name: "empty", body: "", rest: ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rest, trailers := SplitTrailers(c.body)
			if rest != c.rest || !reflect.DeepEqual(trailers, c.trailers) {
				t.Errorf("SplitTrailers(%q) = %q, %v; want %q, %v", c.body, rest, trailers, c.rest, c.trailers)
			}
		})
	}
}
