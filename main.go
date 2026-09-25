// WaterLauncher is a Windows game launcher that finds every game on the PC
// on its own, store installs and unofficial copies alike.
package main

import (
	"embed"
	"log"

	"github.com/ApolloF/WaterLauncher/internal/app"
	"github.com/ApolloF/WaterLauncher/internal/logx"
	"github.com/ApolloF/WaterLauncher/internal/meta"
	"github.com/ApolloF/WaterLauncher/internal/platform"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

// version is set at build time (-ldflags "-X main.version=v0.1.0").
var version = "dev"

func main() {
	core, err := app.NewCore(version)
	if err != nil {
		logx.Printf("start: %v", err)
		log.Fatal(err)
	}
	logx.Printf("WaterLauncher %s starting", version)

	var main *application.WebviewWindow
	wa := application.New(application.Options{
		Name:        "WaterLauncher",
		Description: "Game launcher that finds every game on your PC",
		Services: []application.Service{
			application.NewService(app.NewLibraryService(core)),
			application.NewService(app.NewSettingsService(core)),
		},
		Assets: application.AssetOptions{
			Handler:    application.AssetFileServerFS(assets),
			Middleware: meta.ArtHandler(platform.CacheDir("art")),
		},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: "nl.apollof.waterlauncher",
			OnSecondInstanceLaunch: func(application.SecondInstanceData) {
				if main != nil {
					main.Restore()
					main.Show()
					main.Focus()
				}
			},
		},
		Windows: application.WindowsOptions{WebviewUserDataPath: platform.CacheDir("webview")},
	})

	main = wa.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "main",
		Title:            "WaterLauncher",
		Width:            1440,
		Height:           900,
		MinWidth:         980,
		MinHeight:        620,
		Frameless:        true,
		BackgroundColour: application.NewRGB(10, 14, 19),
		URL:              "/",
		Windows: application.WindowsWindow{
			Theme: application.SystemDefault,
		},
	})

	if err := wa.Run(); err != nil {
		logx.Printf("run: %v", err)
		log.Fatal(err)
	}
}
