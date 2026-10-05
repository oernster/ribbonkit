package delivery

import (
	"fmt"
	"io"
	"strings"

	"github.com/oernster/ribbonkit/domain/identity"
)

// Wails' single-instance lock on Linux owns a bus name made from the instance id, which the kit's
// window sets to the app id: "org.wails_app_" then the id with dashes and dots as underscores, then
// ".SingleInstance" (read in Wails v2.12.0, internal/frontend/desktop/linux/single_instance.go).
const (
	singleInstancePrefix = "org.wails_app_"
	singleInstanceSuffix = ".SingleInstance"
)

// ShellNames writes the names the Linux and macOS build scripts need as shell assignments,
// NAME='value', each read from its one home in the application, so no script keeps a second copy:
//
//	eval "$(go run ./tools/identity)"
func ShellNames(out io.Writer, app identity.App, copyright string) {
	for _, pair := range [][2]string{
		{"APP_NAME", app.Name},
		{"APP_ID", app.AppID},
		{"BIN_NAME", strings.ToLower(app.Name)},
		{"SINGLE_INSTANCE_NAME", singleInstanceName(app.AppID)},
		{"COPYRIGHT", copyright},
	} {
		fmt.Fprintf(out, "%s='%s'\n", pair[0], pair[1])
	}
}

// singleInstanceName answers the bus name Wails' lock owns for id.
func singleInstanceName(id string) string {
	return singleInstancePrefix + strings.NewReplacer("-", "_", ".", "_").Replace(id) + singleInstanceSuffix
}
