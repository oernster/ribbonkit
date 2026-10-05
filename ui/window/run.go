package window

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

// webViewFolder names the web view's own data folder inside the settings folder. Left to Wails, it
// would be a folder named for the executable beside the settings folder, which uninstalling with
// "Also forget my settings" did not reach (FR-806); inside it, forgetting removes it with the rest.
const webViewFolder = "WebView2"

// run runs the window, binding bound and serving the page from assets, keeping the web view's data
// in the settings folder dir; where dir is empty, as in the run that generates bindings, Wails
// chooses. It starts hidden: startup places it and takes it off the taskbar before the page shows it
// (FR-101). One instance runs per user, under the application's id (FR-506).
func (a *Window) run(bound any, assets fs.FS, dir string) error {
	webViewData := ""
	if dir != "" {
		webViewData = filepath.Join(dir, webViewFolder)
	}
	err := wails.Run(&options.App{
		Title:         a.product.App.Name,
		Frameless:     true,
		DisableResize: true,
		StartHidden:   true,
		AssetServer:   &assetserver.Options{Assets: assets},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               a.product.App.AppID,
			OnSecondInstanceLaunch: func(options.SecondInstanceData) { a.secondInstance() },
		},
		// The web view is transparent, so wherever the page draws less than opaque the desktop shows
		// through; the window's own paint follows the chosen opacity (FR-622, opacity.go).
		Windows: &windows.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			BackdropType:         windows.None,
			WindowClassName:      a.product.WindowClass,
			WebviewUserDataPath:  webViewData,
			Theme:                windows.SystemDefault,
			DisablePinchZoom:     true,
			IsZoomControlEnabled: false,
		},
		Mac: &mac.Options{WebviewIsTransparent: true},
		// Wails turns WebKitGTK's hardware acceleration off unless Linux options are given, as the
		// workaround for its blank windows (wailsapp/wails#2977); giving them leaves the policy at
		// its zero value, Always. Never keeps Wails' choice for the ribbon too.
		Linux: &linux.Options{
			WindowIsTranslucent: true,
			WebviewGpuPolicy:    linux.WebviewGpuPolicyNever,
		},
		OnStartup:     a.startup,
		OnDomReady:    a.domReady,
		OnBeforeClose: a.beforeClose,
		OnShutdown:    a.shutdown,
		Bind:          []any{bound},
	})
	if err != nil {
		return fmt.Errorf("running the window: %w", err)
	}
	return nil
}
