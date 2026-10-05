package structure

import (
	"go/parser"
	"go/token"
	"math"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// The size rule is held to the number an editor shows beside the last line.
func TestLineCountCountsTheLinesAnEditorShows(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    int
	}{
		{"an empty file", "", 0},
		{"one line ending in a newline", "one\n", 1},
		{"one line with no newline at the end", "one", 1},
		{"two lines ending in a newline", "one\ntwo\n", 2},
		{"two lines ending in carriage returns and newlines", "one\r\ntwo\r\n", 2},
		{"a blank line before the end", "one\n\n", 2},
	}
	dir := t.TempDir()
	for index, each := range cases {
		path := filepath.Join(dir, "case"+string(rune('a'+index))+".txt")
		if err := os.WriteFile(path, []byte(each.content), 0o600); err != nil {
			t.Fatalf("writing %s: %v", each.name, err)
		}
		if got := LineCount(t, path); got != each.want {
			t.Errorf("%s: LineCount gave %d, want %d", each.name, got, each.want)
		}
	}
}

// A pair under the floor is reported, as is a token neither the scheme nor Classic states; a pair
// that meets it is not. #777 on #fff is 4.48:1, just under; #888 on #000 is 6.26:1.
func TestAContrastShortfallIsReported(t *testing.T) {
	dir := t.TempDir()
	classic := filepath.Join(dir, "theme.css")
	schemes := filepath.Join(dir, "colours.css")
	files := map[string]string{
		classic: ":root {\n  --text: #000;\n  --cell: #fff;\n}\n:root[data-theme='dark'] {\n  --text: #fff;\n  --cell: #000;\n}\n",
		schemes: ":root[data-colour='dim'] {\n  --text: light-dark(#777, #888);\n}\n",
	}
	for path, text := range files {
		if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	half := Half{Classic: classic, Schemes: schemes}
	got := contrastShortfalls(t, []Half{half}, []string{"classic", "dim"}, []string{"text", "absent"}, []string{"cell"})
	want := []string{
		"classic light --absent on --cell: --absent is stated by neither classic nor Classic",
		"classic dark --absent on --cell: --absent is stated by neither classic nor Classic",
		"dim light --text on --cell: #777 on #fff is 4.48:1, under 4.5:1",
		"dim light --absent on --cell: --absent is stated by neither dim nor Classic",
		"dim dark --absent on --cell: --absent is stated by neither dim nor Classic",
	}
	if !slices.Equal(got, want) {
		t.Errorf("shortfalls:\n%q\nwant:\n%q", got, want)
	}
}

// A file's layer and an import's are read from where they sit, in a repository whose layers are
// under a folder and in one whose layers are at its root.
func TestLayersAreReadFromWhereThingsSit(t *testing.T) {
	root := filepath.Join("repo")
	app := Layout{Root: root, Module: "example.test/app", Layered: []string{"internal"}, Dependencies: []string{"example.test/kit"}}
	kit := Layout{Root: root, Module: "example.test/kit", Layered: []string{"."}}
	files := map[string]struct {
		layout Layout
		path   string
		want   string
	}{
		"under a folder":         {app, filepath.Join(root, "internal", "domain", "a.go"), Domain},
		"at the root of a layer": {kit, filepath.Join(root, "ui", "window", "a.go"), UI},
		"outside the layers":     {app, filepath.Join(root, "main.go"), ""},
		"a file of the folder":   {app, filepath.Join(root, "internal", "a.go"), ""},
	}
	for name, each := range files {
		if got := each.layout.LayerOf(each.path); got != each.want {
			t.Errorf("%s: LayerOf(%s) = %q, want %q", name, each.path, got, each.want)
		}
	}
	imports := map[string]string{
		"example.test/app/internal/application/x": Application,
		"example.test/kit/infrastructure/desktop": Infrastructure,
		"example.test/app/internal/product":       "product",
		"example.test/app/tools/x":                "",
		"example.test/kitchen/domain/x":           "",
		"os":                                      "",
	}
	for imported, want := range imports {
		if got := app.ImportLayer(imported); got != want {
			t.Errorf("ImportLayer(%s) = %q, want %q", imported, got, want)
		}
	}
}

func TestNetworkPackageRecognitionIsExact(t *testing.T) {
	for _, network := range []string{"net", "net/http", "crypto/tls", "golang.org/x/net/html"} {
		if !IsNetworkPackage(network) {
			t.Errorf("%s was not recognised", network)
		}
	}
	for _, ordinary := range []string{"network-free", "netlify", "crypto/sha256"} {
		if IsNetworkPackage(ordinary) {
			t.Errorf("%s was taken for a network package", ordinary)
		}
	}
}

func TestRequestRecognitionIsExact(t *testing.T) {
	for _, plant := range []string{"fetch('x')", "window.fetch (url)", "new XMLHttpRequest()", "new WebSocket(u)",
		"navigator.sendBeacon(u)", `url("https://x.test/a.woff")`, "http://x.test"} {
		if len(PageAsks(plant)) == 0 {
			t.Errorf("%q was not recognised", plant)
		}
	}
	for _, ordinary := range []string{"prefetch(x)", "api.fetchClocks(x)", "fetched", `xmlns="http://www.w3.org/2000/svg"`} {
		if asked := PageAsks(ordinary); len(asked) > 0 {
			t.Errorf("%q was taken for a request: %v", ordinary, asked)
		}
	}
	source := `package p; import "os/exec"; var _ = exec.Command`
	call := `package p; import "golang.org/x/sys/windows"; func f() { windows.ShellExecute(0, nil, nil, nil, nil, 0) }`
	library := `package p; var _ = "WinHTTP.DLL"`
	for name, text := range map[string]string{"import": source, "call": call} {
		file, err := parser.ParseFile(token.NewFileSet(), name+".go", text, 0)
		if err != nil || !StartsProcess(file) {
			t.Errorf("the %s was not recognised as starting a program (%v)", name, err)
		}
	}
	file, err := parser.ParseFile(token.NewFileSet(), "library.go", library, 0)
	if err != nil || !slices.Equal(LibrariesNamed(file), []string{"WinHTTP.DLL"}) {
		t.Errorf("the library was not recognised (%v)", err)
	}
}

func TestContrastIsComputedAsTheStandardStatesIt(t *testing.T) {
	// Black on white is the scale's top, 21:1; a colour on itself is its bottom, 1:1. The two greys
	// are the ones usually quoted either side of the floor, which only the sRGB curve gets right.
	cases := []struct {
		first, second string
		want          float64
	}{
		{"#000000", "#ffffff", 21},
		{"#fff", "#000", 21},
		{"#777777", "#777777", 1},
		{"#767676", "#ffffff", 4.54},
		{"#777777", "#ffffff", 4.48},
	}
	// tolerance is half the last place of a ratio quoted to two decimals.
	const tolerance = 0.005
	for _, c := range cases {
		got, err := Contrast(c.first, c.second)
		if err != nil || math.Abs(got-c.want) > tolerance {
			t.Errorf("Contrast(%s, %s) = %v, %v; want %v", c.first, c.second, got, err, c.want)
		}
	}
	if _, err := Contrast("rgb(0 0 0)", "#fff"); err == nil {
		t.Error("a colour form the check cannot read was accepted")
	}
}
