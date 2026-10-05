package structural

// The setup program's page has no build step, so nothing compiles it: a script the page never
// loads defines nothing that runs, which no test of the script alone can notice.

import (
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/oernster/ribbonkit/structure"
)

// setupPage is the page itself; setupScriptTag matches one script it loads.
var (
	setupPage      = "index.html"
	setupScriptTag = regexp.MustCompile(`<script src="([^"]+)"></script>`)
)

func TestTheSetupPageLoadsEveryScript(t *testing.T) {
	page := structure.Read(t, filepath.Join(structure.Root(t), setupPageDir, setupPage))
	var loaded []string
	for _, match := range setupScriptTag.FindAllStringSubmatch(page, -1) {
		loaded = append(loaded, match[1])
	}
	for _, path := range setupPageFiles(t) {
		name := filepath.Base(path)
		if strings.EqualFold(filepath.Ext(name), ".js") && !slices.Contains(loaded, name) {
			t.Errorf("the setup page never loads %s, so nothing it defines runs", name)
		}
	}
}
