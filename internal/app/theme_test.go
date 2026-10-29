package app

import "testing"

func TestThemeRoundTrip(t *testing.T) {
	a := newTestApp(t)

	if got, err := a.Theme(); err != nil || got != "" {
		t.Fatalf("default theme = %q, %v", got, err)
	}
	if err := a.SetTheme("dark"); err != nil {
		t.Fatal(err)
	}
	if got, _ := a.Theme(); got != "dark" {
		t.Fatalf("theme = %q", got)
	}
	if err := a.SetTheme("neon"); err == nil {
		t.Fatal("unknown theme accepted")
	}
	if got, _ := a.Theme(); got != "dark" {
		t.Fatalf("theme after bad set = %q", got)
	}
}
