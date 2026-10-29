package app

import "github.com/risadams/gitrevenge/internal/settings"

// Theme returns the saved theme: "light", "dark", or "" to follow the operating system.
func (a *App) Theme() (string, error) {
	m, err := a.sessions()
	if err != nil {
		return "", err
	}
	return m.Settings().Get().Theme, nil
}

// SetTheme saves the theme: "light", "dark", or "" to follow the operating system.
func (a *App) SetTheme(theme string) error {
	switch theme {
	case "", "light", "dark":
	default:
		return userError("Unknown theme " + theme + ".")
	}
	m, err := a.sessions()
	if err != nil {
		return err
	}
	return m.Settings().Update(func(s *settings.Settings) { s.Theme = theme })
}
