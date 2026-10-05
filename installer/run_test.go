//go:build windows

package installer

import (
	"io/fs"
	"regexp"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

// pictureName finds a picture a page file names outright.
var pictureName = regexp.MustCompile(`[A-Za-z0-9_-]+\.png`)

// themeFaces are the appearances setup-shell.js switches between; it names each face's picture as
// the appearance followed by themeSuffix.
var themeFaces = []string{"light", "dark"}

// themeSuffix is how setup-shell.js ends a theme face's picture name.
const themeSuffix = "-mode.png"

// Every picture the page shows must be one the application is asked for; any other shows broken.
func TestPicturesNamesEveryPictureThePageShows(t *testing.T) {
	assets, err := fs.Sub(page, pageRoot)
	if err != nil {
		t.Fatal(err)
	}
	var shown []string
	err = fs.WalkDir(assets, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		raw, err := fs.ReadFile(assets, path)
		if err != nil {
			return err
		}
		text := string(raw)
		shown = append(shown, pictureName.FindAllString(strings.ReplaceAll(text, "'"+themeSuffix+"'", ""), -1)...)
		if strings.Contains(text, "'"+themeSuffix+"'") {
			for _, face := range themeFaces {
				shown = append(shown, face+themeSuffix)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(shown) == 0 {
		t.Fatal("the page names no picture, the scan is wrong")
	}
	for _, name := range shown {
		if !slices.Contains(Pictures, name) {
			t.Errorf("the page shows %s, which Pictures does not ask the application for", name)
		}
	}
	for _, name := range Pictures {
		if !slices.Contains(shown, name) {
			t.Errorf("Pictures asks for %s, which the page never shows", name)
		}
	}
}

func TestMissingNamesEachPictureNotSupplied(t *testing.T) {
	t.Parallel()
	all := fstest.MapFS{}
	for _, name := range Pictures {
		all[name] = &fstest.MapFile{Data: []byte(name)}
	}
	if missing := Missing(all); len(missing) != 0 {
		t.Errorf("every picture supplied, yet %v reported missing", missing)
	}
	delete(all, Pictures[0])
	if missing := Missing(all); !slices.Equal(missing, Pictures[:1]) {
		t.Errorf("missing %v, want %v", missing, Pictures[:1])
	}
}

func TestThePicturesAreLaidOverThePage(t *testing.T) {
	t.Parallel()
	pictures := fstest.MapFS{"icon.png": {Data: []byte("the application's mark")}}
	shown := fstest.MapFS{"icon.png": {Data: []byte("the page's own")}, "index.html": {Data: []byte("the page")}}
	served := layered{pictures: pictures, page: shown}
	for name, want := range map[string]string{"icon.png": "the application's mark", "index.html": "the page"} {
		got, err := fs.ReadFile(served, name)
		if err != nil || string(got) != want {
			t.Errorf("%s served %q (%v), want %q", name, got, err, want)
		}
	}
	if _, err := served.Open("absent.png"); err == nil {
		t.Error("a file neither holds was served")
	}
}
