// main.go
package main

import (
	"embed"
	"log"
	"os"

	"smegg.me/thughunter/common/config"
	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/datastore"
	"smegg.me/thughunter/core/scraper"
	configservice "smegg.me/thughunter/services/config"
	loggerservice "smegg.me/thughunter/services/logger"
	themeservice "smegg.me/thughunter/services/theme"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func Cleanup() error {
	logger.Info().Msg("cleaning up...")
	return nil
}

func RunApplication() {
	if config.LaunchedViaAppImage() {
		logger.Info().Int("WEBKIT_DISABLE_DMABUF_RENDERER", 1).Msg("running in AppImage environment, setting")

		// Disable WebKit's DMA-BUF renderer to avoid rendering issues on some Linux systems. On my EndeavourOS, KDE Plasma with Wayland and AMD GPU it causes the entire window to go white and then crash.
		os.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "1")
	}

	app := application.New(application.Options{
		Name:        "Thug Hunter",
		Description: "Program for hunting thugs on the internet, powered by Censys Search in the background.",
		Services: []application.Service{
			application.NewService(&configservice.Service{}),
			application.NewService(&loggerservice.Service{}),
			application.NewService(&themeservice.Service{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "Thug Hunter",
		Width:     1200,
		Height:    800,
		MinWidth:  1000,
		MinHeight: 650,
		// This fixes a weird issue on wayland where the window doesn't maximize properly, issue #4429 in the Wails repo.
		MaxWidth:  99999,
		MaxHeight: 99999,
		URL:       "/",
		Mac: application.MacWindow{
			CollectionBehavior: application.MacWindowCollectionBehaviorFullScreenPrimary,
		},
	})

	themeservice.RegisterThemeWatcher(app)

	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	defer func() {
		if err := Cleanup(); err != nil {
			logger.Fatal().Err(err).Msg("cleanup")
		}
	}()

	logger.Initialize()
	logger.Info().Msg("starting thughunter...")

	logger.Debug().Msg("loading configuration")
	if path, err := config.Initialize(); err != nil {
		if e, ok := config.IsDefaultGenerated(err); ok {
			logger.Info().Str("config_path", e.Path).Msg("default config generated")
			os.Exit(0)
		}
		logger.Fatal().Msg(err.Error())
	} else {
		logger.Debug().Str("config_path", path).Msg("configuration loaded")
	}

	cfg := config.Get()
	logger.Debug().Msg("reinitializing logger with config")

	if err := logger.ReInitialize(cfg.Logger); err != nil {
		logger.Fatal().Err(err).Msg("logger")
	}

	logger.Debug().Str("db_path", cfg.Db.Path).Msg("initializing datastore")
	if err := datastore.Initialize(cfg.Db.Path); err != nil {
		logger.Fatal().Err(err).Msg("datastore")
	}

	if err := scraper.Initialize(); err != nil {
		logger.Fatal().Err(err).Msg("scraper")
	}

	scraper := scraper.Get()
	scraper.CreateAgent()

	// RunApplication()
}
