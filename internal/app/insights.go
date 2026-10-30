package app

import (
	"encoding/base64"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// SaveFile asks where to save a file (an export) and writes the base64 content there. It
// returns the chosen path, or "" if the user cancelled.
func (a *App) SaveFile(title, defaultName, contentBase64 string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(contentBase64)
	if err != nil {
		return "", err
	}
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{Title: title, DefaultFilename: defaultName})
	if err != nil || path == "" {
		return "", err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", userError("Couldn't save " + path + ": " + err.Error())
	}
	return path, nil
}
