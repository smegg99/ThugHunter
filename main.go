// main.go
package main

import (
	"embed"
	"log"
	"os"
	"reflect"
	"sync/atomic"
	"time"
	"unsafe"

	"smegg.me/thughunter/common/config"
	"smegg.me/thughunter/common/i18n"
	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/datastore"
	"smegg.me/thughunter/core/scanner"
	"smegg.me/thughunter/core/scraper"
	configservice "smegg.me/thughunter/services/config"
	loggerservice "smegg.me/thughunter/services/logger"
	monitorservice "smegg.me/thughunter/services/monitor"
	browserservice "smegg.me/thughunter/services/program"
	scannerservice "smegg.me/thughunter/services/scanner"
	scraperservice "smegg.me/thughunter/services/scraper"
	themeservice "smegg.me/thughunter/services/theme"
	trayservice "smegg.me/thughunter/services/tray"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var appIcon []byte

func Cleanup() error {
	logger.Info().Msg("cleaning up...")
	return nil
}

func setAppImageEnv() {
	if config.LaunchedViaAppImage() {
		logger.Info().Int("WEBKIT_DISABLE_DMABUF_RENDERER", 1).Msg("running in AppImage environment, setting")
		os.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "1")
	}
}

// mainWindow holds the primary window so the single-instance callback
// can show and focus it when a second instance is launched.
var mainWindow *application.WebviewWindow

func createApp() *application.App {
	return application.New(application.Options{
		Name:        i18n.T("app.name"),
		Description: i18n.T("app.description"),
		Icon:        appIcon,
		Services: []application.Service{
			application.NewService(&browserservice.Service{}),
			application.NewService(&configservice.Service{}),
			application.NewService(&loggerservice.Service{}),
			application.NewService(&scraperservice.Service{}),
			application.NewService(&scannerservice.Service{}),
			application.NewService(&themeservice.Service{}),
			application.NewService(&monitorservice.Service{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: "me.smegg.thughunter",
			OnSecondInstanceLaunch: func(data application.SecondInstanceData) {
				logger.Info().Strs("args", data.Args).Msg("second instance launched, focusing existing window")
				if mainWindow != nil {
					mainWindow.Show()
					mainWindow.Focus()
				}
			},
		},
		Linux: application.LinuxOptions{
			ProgramName: "thughunter",
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})
}

func createWindow(app *application.App) *application.WebviewWindow {
	return app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     i18n.T("app.name"),
		Width:     1200,
		Height:    750,
		MinWidth:  1200,
		MinHeight: 750,
		// This fixes a weird issue on wayland where the window doesn't maximize properly, issue #4429 in the Wails repo.
		MaxWidth:  99999,
		MaxHeight: 99999,
		URL:       "/",
		Mac: application.MacWindow{
			CollectionBehavior: application.MacWindowCollectionBehaviorFullScreenPrimary,
		},
	})
}

func setupSystemTray(app *application.App, window *application.WebviewWindow, quitting *bool) {
	tray := app.SystemTray.New()
	tray.SetTooltip(i18n.T("tray.tooltip"))

	// Use the tray manager for theme- and state-aware icons.
	trayMgr := trayservice.NewManager(tray)
	trayMgr.RegisterEvents(app)

	buildTrayMenu := func() {
		trayMenu := app.NewMenu()
		trayMenu.Add(i18n.T("tray.show")).OnClick(func(ctx *application.Context) {
			window.Show()
			window.Focus()
		})
		trayMenu.AddSeparator()
		trayMenu.Add(i18n.T("tray.quit")).OnClick(func(ctx *application.Context) {
			*quitting = true
			app.Quit()
		})
		tray.SetMenu(trayMenu)
		tray.SetTooltip(i18n.T("tray.tooltip"))
	}
	buildTrayMenu()

	// Rebuild the tray menu when the locale changes.
	app.Event.On("config:preferences:changed", func(_ *application.CustomEvent) {
		buildTrayMenu()
	})

	// Workaround for Linux/KDE: the dbusmenu "opened" event
	// incorrectly calls clickHandler when the context menu opens on right-click.
	// Hook the unexported onMenuOpen field to detect and suppress this.
	var menuOpened atomic.Bool
	field := reflect.ValueOf(tray).Elem().FieldByName("onMenuOpen")
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(
		reflect.ValueOf(func() { menuOpened.Store(true) }),
	)

	tray.OnClick(func() {
		go func() {
			time.Sleep(5 * time.Millisecond)
			if menuOpened.CompareAndSwap(true, false) {
				return
			}
			if window.IsVisible() {
				window.Hide()
			} else {
				window.Show()
				window.Focus()
			}
		}()
	})
}

func setupCloseToTray(window *application.WebviewWindow, quitting *bool) {
	window.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		if *quitting {
			return
		}
		if config.Get().Preferences.CloseToTray {
			window.Hide()
			e.Cancel()
		}
	})
}

// setupLocaleSync listens for preference changes and updates the backend locale.
func setupLocaleSync(app *application.App) {
	app.Event.On("config:preferences:changed", func(_ *application.CustomEvent) {
		if lang := config.Get().Preferences.Language; lang != "" {
			i18n.SetLocale(lang)
		}
	})
}

func RunApplication() {
	setAppImageEnv()

	app := createApp()
	window := createWindow(app)
	mainWindow = window

	quitting := false
	setupLocaleSync(app)
	setupSystemTray(app, window, &quitting)
	setupCloseToTray(window, &quitting)

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

	// Set the backend locale from the user's saved language preference.
	if cfg.Preferences.Language != "" {
		i18n.SetLocale(cfg.Preferences.Language)
	}

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

	scanner.Initialize()

	RunApplication()
}
