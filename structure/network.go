package structure

// A ribbon's one network request is its update check (TimeRibbon NFR-S-1, FR-509). A request can
// leave through a network package, a program started for it, a system library loaded by name or
// the page itself; these checks hold each of those doors. What they cannot see: code reached
// through cgo's own C, a library Wails or the web view loads on its own account and a name built at
// run time, which is why the rules are held over the source rather than the binary.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// networkPackages are the packages a request is made through; a path beneath one is one of them.
var networkPackages = []string{"net", "crypto/tls", "golang.org/x/net"}

// processImport and processCalls are how a Go file starts another program.
const processImport = "os/exec"

var processCalls = []string{
	"os.StartProcess", "syscall.StartProcess", "syscall.ForkExec", "syscall.CreateProcess",
	"windows.CreateProcess", "windows.ShellExecute",
}

// SystemLibraries are the libraries a Windows build may load by name; a request library such as
// winhttp.dll or ws2_32.dll is not among them.
var SystemLibraries = []string{"user32.dll", "shell32.dll", "kernel32.dll", "gdi32.dll", "shcore.dll"}

// libraryName matches a string naming a Windows library.
var libraryName = regexp.MustCompile(`(?i)\.dll$`)

// pageRequest matches what makes a page ask a network: a request API or an address to fetch. The
// SVG namespace names a vocabulary, which nothing fetches.
var (
	pageRequest  = regexp.MustCompile(`\bfetch\s*\(|XMLHttpRequest|WebSocket|EventSource|sendBeacon|importScripts|WebTransport|RTCPeerConnection|https?://`)
	svgNamespace = "http://www.w3.org/"
)

// IsNetworkPackage answers whether imported is a package a request is made through.
func IsNetworkPackage(imported string) bool {
	return slices.ContainsFunc(networkPackages, func(network string) bool {
		return imported == network || strings.HasPrefix(imported, network+"/")
	})
}

// StartsProcess answers whether file imports os/exec or calls one of processCalls.
func StartsProcess(file *ast.File) bool {
	for _, item := range file.Imports {
		if strings.Trim(item.Path.Value, `"`) == processImport {
			return true
		}
	}
	found := false
	ast.Inspect(file, func(node ast.Node) bool {
		if selector, ok := node.(*ast.SelectorExpr); ok {
			if name, ok := selector.X.(*ast.Ident); ok && slices.Contains(processCalls, name.Name+"."+selector.Sel.Name) {
				found = true
			}
		}
		return !found
	})
	return found
}

// LibrariesNamed answers every string in file naming a Windows library.
func LibrariesNamed(file *ast.File) []string {
	var named []string
	ast.Inspect(file, func(node ast.Node) bool {
		if literal, ok := node.(*ast.BasicLit); ok && literal.Kind == token.STRING {
			if text, err := strconv.Unquote(literal.Value); err == nil && libraryName.MatchString(text) {
				named = append(named, text)
			}
		}
		return true
	})
	return named
}

// PageAsks answers each request the page text makes.
func PageAsks(text string) []string {
	return pageRequest.FindAllString(strings.ReplaceAll(text, svgNamespace, ""), -1)
}

// CheckOnlyTheExemptImportANetworkPackage fails for each Go file outside the exempt directories,
// from root, that imports a network package; also for each exempt directory that is not there, so
// a move cannot leave an exemption pointing at nothing while the package's new home goes unchecked.
func CheckOnlyTheExemptImportANetworkPackage(t testing.TB, root string, files []string, exempt ...string) {
	t.Helper()
	for _, dir := range exempt {
		if info, err := os.Stat(filepath.Join(root, filepath.FromSlash(dir))); err != nil || !info.IsDir() {
			t.Errorf("%s is exempt but is not a directory: %v", dir, err)
		}
	}
	for _, path := range files {
		if slices.Contains(exempt, Relative(root, filepath.Dir(path))) {
			continue
		}
		for _, imported := range ImportsOf(t, path) {
			if IsNetworkPackage(imported) {
				t.Errorf("%s imports %s; only the update check reaches the network", path, imported)
			}
		}
	}
}

// CheckOnlyNamedFilesStartAProcess fails for each shipped Go file not among starters (paths from
// root with forward slashes) that starts a program, for each library named outside
// SystemLibraries and for each starter that is not there, so a moved file cannot leave its
// permission behind for another.
func CheckOnlyNamedFilesStartAProcess(t testing.TB, root string, files []string, starters []string) {
	t.Helper()
	for _, starter := range starters {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(starter))); err != nil {
			t.Errorf("%s may start a program but is not there: %v", starter, err)
		}
	}
	for _, path := range Shipped(files) {
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatalf("parsing %s: %v", path, err)
		}
		if StartsProcess(file) && !slices.Contains(starters, Relative(root, path)) {
			t.Errorf("%s starts a program; only the named starters may", path)
		}
		for _, library := range LibrariesNamed(file) {
			if !slices.Contains(SystemLibraries, strings.ToLower(library)) {
				t.Errorf("%s names %s, a library not in SystemLibraries", path, library)
			}
		}
	}
}

// CheckThePageMakesNoRequest fails for each shipped page file that asks a network.
func CheckThePageMakesNoRequest(t testing.TB, files []string) {
	t.Helper()
	for _, path := range Shipped(files) {
		if asked := PageAsks(Read(t, path)); len(asked) > 0 {
			t.Errorf("%s makes a request (%v); only Go's update check asks a network", path, asked)
		}
	}
}
