package desktop

import (
	"os"
	"testing"

	"github.com/oernster/ribbonkit/infrastructure/gtkmain"
)

// The tests run on X11 with GTK's loop running, as the application does.
func TestMain(m *testing.M) { os.Exit(gtkmain.ServeTests(m.Run)) }
