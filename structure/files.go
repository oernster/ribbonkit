// Package structure is the mechanics of the structural tests every repository built on the kit
// runs: finding its files, reading their imports, counting their lines and holding them to the
// layer, size, network and wire rules. Each repository keeps its own rules' data (its module, its
// layered folders, its exemptions) in its own tests and hands it here, so the checks have one home.
//
// Every check takes the test it reports to and fails it rather than answering a list, so a caller
// cannot forget to look. Every check that walks also fails when it finds nothing to walk, since an
// empty walk passes every rule.
package structure

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// alwaysSkipped are walked past in every repository: git's own folder and other people's code.
var alwaysSkipped = []string{".git", "node_modules"}

// Root answers the directory holding go.mod, walking up from the test's directory.
func Root(t testing.TB) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("no go.mod above the test directory")
		}
		dir = parent
	}
}

// GoFiles answers every Go source file under root, walking past alwaysSkipped and the folders
// named in skipped.
func GoFiles(t testing.TB, root string, skipped ...string) []string {
	t.Helper()
	return walk(t, root, func(path string) bool { return strings.HasSuffix(path, ".go") }, skipped)
}

// FilesWith answers every file under each of dirs whose extension is one of extensions, failing
// for a folder that holds none.
func FilesWith(t testing.TB, extensions []string, dirs ...string) []string {
	t.Helper()
	wanted := map[string]bool{}
	for _, extension := range extensions {
		wanted[extension] = true
	}
	var found []string
	for _, dir := range dirs {
		found = append(found, walk(t, dir, func(path string) bool { return wanted[filepath.Ext(path)] }, nil)...)
	}
	return found
}

// walk answers every file under root that keeps answers true for, failing when there is none.
func walk(t testing.TB, root string, keeps func(path string) bool, skipped []string) []string {
	t.Helper()
	skip := map[string]bool{}
	for _, name := range append(append([]string{}, alwaysSkipped...), skipped...) {
		skip[name] = true
	}
	var found []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && path != root && skip[entry.Name()] {
			return filepath.SkipDir
		}
		if !entry.IsDir() && keeps(path) {
			found = append(found, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}
	if len(found) == 0 {
		t.Fatalf("no file found under %s, the walk is wrong", root)
	}
	return found
}

// Shipped answers the files of paths that are not tests, which never ship.
func Shipped(paths []string) []string {
	var shipped []string
	for _, path := range paths {
		if !strings.HasSuffix(path, "_test.go") && !strings.Contains(filepath.Base(path), ".test.") {
			shipped = append(shipped, path)
		}
	}
	return shipped
}

// ImportsOf answers the import paths of the Go file at path.
func ImportsOf(t testing.TB, path string) []string {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}
	out := make([]string, 0, len(parsed.Imports))
	for _, item := range parsed.Imports {
		out = append(out, strings.Trim(item.Path.Value, `"`))
	}
	return out
}

// Read answers the text of the file at path.
func Read(t testing.TB, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(raw)
}

// LineCount counts the lines in a file as an editor numbers them: a newline ends a line rather than
// starting one.
func LineCount(t testing.TB, path string) int {
	t.Helper()
	text := Read(t, path)
	lines := strings.Count(text, "\n")
	if text != "" && !strings.HasSuffix(text, "\n") {
		lines++
	}
	return lines
}

// Relative answers path from root with forward slashes, as the rules name files.
func Relative(root, path string) string {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(relative)
}
