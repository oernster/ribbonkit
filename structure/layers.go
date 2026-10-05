package structure

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// The four layers, named as the folders that hold them.
const (
	Domain         = "domain"
	Application    = "application"
	Infrastructure = "infrastructure"
	UI             = "ui"
)

// lineLimit is the module-size cap. dangerBand is five per cent below it: a file that lands between
// them is refactored down to safeLanding rather than left one edit away from breaching.
const (
	lineLimit   = 400
	dangerBand  = lineLimit - lineLimit/20
	safeLanding = 350
)

// forbiddenInDomain names the packages that would give the domain IO, randomness or a tz database
// of its own.
var forbiddenInDomain = []string{
	"net", "net/http", "os", "os/exec", "path/filepath", "io/ioutil", "math/rand", "math/rand/v2",
	"time/tzdata", "syscall", "golang.org/x/sys/windows",
}

// forbiddenCallsInDomain read the wall clock or load a zone, which would make the domain's answers
// depend on the machine.
var forbiddenCallsInDomain = []string{"time.Now", "time.Since", "time.Until", "time.LoadLocation"}

// Layout is how one repository is laid out in layers: what the rules need to know a file's layer
// and an import's.
type Layout struct {
	// Root is the repository's directory on disk.
	Root string
	// Module is the module path go.mod states.
	Module string
	// Layered are the folders under Root whose subfolders are the layers; "." when the layers sit
	// at Root itself.
	Layered []string
	// Dependencies are the import paths of other modules laid out the same way, layers directly
	// beneath, so an import of one of their layers is held to the same rules.
	Dependencies []string
	// CompositionRoot names the files, from Root with forward slashes, allowed to wire the
	// application to the infrastructure.
	CompositionRoot []string
}

// LayerOf answers the layer the file at path belongs to; empty outside the layered folders.
func (l Layout) LayerOf(path string) string {
	parts := strings.Split(Relative(l.Root, path), "/")
	for _, folder := range l.Layered {
		if folder == "." && len(parts) >= 2 {
			return parts[0]
		}
		if len(parts) >= 3 && parts[0] == folder {
			return parts[1]
		}
	}
	return ""
}

// ownedPrefixes are the import paths whose next part names a layer.
func (l Layout) ownedPrefixes() []string {
	prefixes := slices.Clone(l.Dependencies)
	for _, folder := range l.Layered {
		if folder == "." {
			prefixes = append(prefixes, l.Module)
		} else {
			prefixes = append(prefixes, l.Module+"/"+folder)
		}
	}
	return prefixes
}

// owned answers whether imported is a package of this module or of a dependency laid out alike.
func (l Layout) owned(imported string) bool {
	if strings.HasPrefix(imported, l.Module+"/") {
		return true
	}
	return slices.ContainsFunc(l.Dependencies, func(dependency string) bool {
		return strings.HasPrefix(imported, dependency+"/")
	})
}

// ImportLayer answers the layer imported is a package of; empty for any other package.
func (l Layout) ImportLayer(imported string) string {
	for _, prefix := range l.ownedPrefixes() {
		if inner, ok := strings.CutPrefix(imported, prefix+"/"); ok {
			layer, _, _ := strings.Cut(inner, "/")
			return layer
		}
	}
	return ""
}

// CheckDomainHasNoOutwardImports fails for each domain file importing anything of this module or
// its laid-out dependencies outside a domain.
func (l Layout) CheckDomainHasNoOutwardImports(t testing.TB, files []string) {
	t.Helper()
	l.eachImport(t, files, Domain, func(path, imported string) {
		t.Helper()
		if l.owned(imported) && l.ImportLayer(imported) != Domain {
			t.Errorf("%s imports %s: the domain depends on nothing", path, imported)
		}
	})
}

// CheckDomainIsPure fails for each domain file, tests aside, that imports IO, randomness or a tz
// database; also for each that reads the wall clock.
func (l Layout) CheckDomainIsPure(t testing.TB, files []string) {
	t.Helper()
	for _, path := range Shipped(files) {
		if l.LayerOf(path) != Domain {
			continue
		}
		for _, imported := range ImportsOf(t, path) {
			if slices.Contains(forbiddenInDomain, imported) {
				t.Errorf("%s imports %q: the domain performs no IO", path, imported)
			}
		}
		text := Read(t, path)
		for _, call := range forbiddenCallsInDomain {
			if strings.Contains(text, call+"(") {
				t.Errorf("%s calls %s: take the instant or the zone as an argument", path, call)
			}
		}
	}
}

// CheckApplicationDoesNotImportInfrastructure fails for each application file reaching an
// infrastructure package or Wails.
func (l Layout) CheckApplicationDoesNotImportInfrastructure(t testing.TB, files []string) {
	t.Helper()
	l.eachImport(t, files, Application, func(path, imported string) {
		t.Helper()
		if l.ImportLayer(imported) == Infrastructure || strings.Contains(imported, "wails") {
			t.Errorf("%s imports %s: the application depends on ports only", path, imported)
		}
	})
}

// CheckUIDependsOnTheApplicationOnly fails for each UI file, tests included, reaching an
// infrastructure package: a test that borrows an adapter runs against whichever platform builds it.
func (l Layout) CheckUIDependsOnTheApplicationOnly(t testing.TB, files []string) {
	t.Helper()
	l.eachImport(t, files, UI, func(path, imported string) {
		t.Helper()
		if l.ImportLayer(imported) == Infrastructure {
			t.Errorf("%s imports %s: the UI depends on the application's ports only", path, imported)
		}
	})
}

// CheckNothingBelowTheUIImportsIt fails for each application or infrastructure file reaching up
// into the UI.
func (l Layout) CheckNothingBelowTheUIImportsIt(t testing.TB, files []string) {
	t.Helper()
	for _, layer := range []string{Application, Infrastructure} {
		l.eachImport(t, files, layer, func(path, imported string) {
			t.Helper()
			if l.ImportLayer(imported) == UI {
				t.Errorf("%s imports %s: nothing below the UI depends on it", path, imported)
			}
		})
	}
}

// CheckWailsStaysOutOfInfrastructure fails for each infrastructure file importing Wails.
func (l Layout) CheckWailsStaysOutOfInfrastructure(t testing.TB, files []string) {
	t.Helper()
	l.eachImport(t, files, Infrastructure, func(path, imported string) {
		t.Helper()
		if strings.Contains(imported, "wails") {
			t.Errorf("%s imports %s: Wails belongs to the composition root", path, imported)
		}
	})
}

// CheckCompositionRootIsWhitelisted fails for each shipped file outside CompositionRoot importing
// both an application and an infrastructure package.
func (l Layout) CheckCompositionRootIsWhitelisted(t testing.TB, files []string) {
	t.Helper()
	for _, path := range Shipped(files) {
		var application, infrastructure bool
		for _, imported := range ImportsOf(t, path) {
			application = application || l.ImportLayer(imported) == Application
			infrastructure = infrastructure || l.ImportLayer(imported) == Infrastructure
		}
		if relative := Relative(l.Root, path); application && infrastructure && !slices.Contains(l.CompositionRoot, relative) {
			t.Errorf("%s wires application to infrastructure: only the composition root may", relative)
		}
	}
}

// eachImport calls found with every import of every file in layer.
func (l Layout) eachImport(t testing.TB, files []string, layer string, found func(path, imported string)) {
	t.Helper()
	for _, path := range files {
		if l.LayerOf(path) != layer {
			continue
		}
		for _, imported := range ImportsOf(t, path) {
			found(path, imported)
		}
	}
}

// CheckNoFileExceedsLineLimit fails for each file over the size cap.
func CheckNoFileExceedsLineLimit(t testing.TB, files []string) {
	t.Helper()
	for _, path := range files {
		if count := LineCount(t, path); count > lineLimit {
			t.Errorf("%s has %d lines, over the %d limit", path, count, lineLimit)
		}
	}
}

// CheckNoFileInDangerBand fails for each file inside the five per cent below the cap.
func CheckNoFileInDangerBand(t testing.TB, files []string) {
	t.Helper()
	for _, path := range files {
		if count := LineCount(t, path); count > dangerBand && count <= lineLimit {
			t.Errorf("%s has %d lines, inside the danger band %d to %d: reduce it to %d or fewer",
				path, count, dangerBand+1, lineLimit, safeLanding)
		}
	}
}

// CheckEveryExportedTypeIsDocumented fails for each exported type in a shipped Go file that has no
// doc comment.
func CheckEveryExportedTypeIsDocumented(t testing.TB, files []string) {
	t.Helper()
	for _, path := range Shipped(files) {
		if filepath.Ext(path) != ".go" {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parsing %s: %v", path, err)
		}
		for _, declaration := range parsed.Decls {
			general, ok := declaration.(*ast.GenDecl)
			if !ok || general.Tok != token.TYPE || general.Doc != nil {
				continue
			}
			for _, spec := range general.Specs {
				if typed, ok := spec.(*ast.TypeSpec); ok && typed.Name.IsExported() && typed.Doc == nil {
					t.Errorf("%s: exported type %s has no doc comment", path, typed.Name.Name)
				}
			}
		}
	}
}
