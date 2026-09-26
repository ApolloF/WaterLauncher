// WaterLauncher is a Windows game launcher that finds every game on the PC
// on its own, store installs and unofficial copies alike.
package main

import (
	"embed"
	"log"
	"os"

	"github.com/ApolloF/WaterLauncher/internal/app"
	"github.com/ApolloF/WaterLauncher/internal/logx"
	"github.com/ApolloF/WaterLauncher/internal/meta"
	"github.com/ApolloF/WaterLauncher/internal/platform"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
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
	shell := app.NewShell(core)
	launcher := app.NewLaunchService(core)

	wa := application.New(application.Options{
		Name:        "WaterLauncher",
		Description: "Game launcher that finds every game on your PC",
		Services: []application.Service{
			application.NewService(app.NewLibraryService(core)),
			application.NewService(launcher),
			application.NewService(app.NewSavesService(core)),
			application.NewService(app.NewAddonsService(core)),
			application.NewService(app.NewSettingsService(core)),
			application.NewService(app.NewPadService(core)),
		},
		Assets: application.AssetOptions{
			Handler:    application.AssetFileServerFS(assets),
			Middleware: meta.ArtHandler(platform.CacheDir("art")),
		},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: "nl.apollof.waterlauncher",
			OnSecondInstanceLaunch: func(d application.SecondInstanceData) {
				if !launcher.PlayFromArgs(d.Args) {
					shell.OpenMain()
				}
			},
		},
		Windows: application.WindowsOptions{
			WebviewUserDataPath: platform.CacheDir("webview"),
			// The interface closes while a game runs; WaterLauncher keeps
			// going in the tray. Closing the window yourself still quits.
			DisableQuitOnLastWindowClosed: true,
		},
	})
	// "--play <id>" starts a game straight away, without the interface.
	if _, ok := app.PlayArg(os.Args[1:]); ok {
		shell.StartHidden()
		wa.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*application.ApplicationEvent) {
			go launcher.PlayFromArgs(os.Args[1:])
		})
	} else {
		shell.Start()
	}

	if err := wa.Run(); err != nil {
		logx.Printf("run: %v", err)
		log.Fatal(err)
	}
}
