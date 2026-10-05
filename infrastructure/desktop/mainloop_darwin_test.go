package desktop

import (
	"os"
	"testing"

	"github.com/oernster/ribbonkit/infrastructure/cocoamain"
)

// The tests run with AppKit's loop running on the main thread, as the application does.
func TestMain(m *testing.M) { os.Exit(cocoamain.Serve(m.Run)) }
