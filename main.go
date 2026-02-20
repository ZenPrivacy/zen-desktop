package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/ZenPrivacy/zen-desktop/internal/app"
	"github.com/ZenPrivacy/zen-desktop/internal/autostart"
	"github.com/ZenPrivacy/zen-desktop/internal/cfg"
	"github.com/ZenPrivacy/zen-desktop/internal/constants"
	"github.com/ZenPrivacy/zen-desktop/internal/logger"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	debug.SetGCPercent(5)
	startOnDomReady := flag.Bool("start", false, "Start the service when DOM is ready")
	startHidden := flag.Bool("hidden", false, "Start the application in hidden mode")
	uninstallCA := flag.Bool("uninstall-ca", false, "Uninstall the CA and exit")
	flag.Parse()

	err := logger.SetupLogger()
	if err != nil {
		log.Printf("failed to setup logger: %v", err)
	}
	log.Printf("initializing the app; version=%q", cfg.Version)

	config, err := cfg.NewConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	app, err := app.NewApp(constants.AppName, config, *startOnDomReady)
	if err != nil {
		log.Fatalf("failed to create app: %v", err)
	}

	if *uninstallCA {
		if err := app.UninstallCA(); err != nil {
			// UninstallCA logs the error internally
			os.Exit(1)
		}

		log.Println("CA uninstalled successfully")
		return
	}

	autostart := &autostart.Manager{}

	err = wails.Run(&options.App{
		Title:         constants.AppName,
		Width:         420,
		Height:        650,
		DisableResize: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:     app.Startup,
		OnBeforeClose: app.BeforeClose,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               constants.InstanceID,
			OnSecondInstanceLaunch: app.OnSecondInstanceLaunch,
		},
		Bind: []interface{}{
			app,
			config,
			autostart,
		},
		EnumBind: []interface{}{
			cfg.UpdatePolicyEnum,
		},
		Mac: &mac.Options{
			About: &mac.AboutInfo{
				Title:   constants.AppName,
				Message: fmt.Sprintf("Your Comprehensive Ad-Blocker and Privacy Guard\nVersion: %s\n© 2026 Zen contributors", cfg.Version),
			},
		},
		HideWindowOnClose: runtime.GOOS == "darwin" || runtime.GOOS == "windows",
		StartHidden:       *startHidden,
	})

	if err != nil {
		log.Fatal(err)
	}
}
