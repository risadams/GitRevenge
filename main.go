// Command gitrevenge is the Git Revenge desktop git client.
package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"github.com/risadams/gitrevenge/internal/app"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	a := app.New()

	err := wails.Run(&options.App{
		Title:     "Git Revenge",
		Width:     1280,
		Height:    800,
		MinWidth:  800,
		MinHeight: 500,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 30, G: 30, B: 30, A: 1},
		OnStartup:        a.Startup,
		OnShutdown:       a.Shutdown,
		Bind: []any{
			a,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
