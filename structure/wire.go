package structure

// The wire is stated twice: Go structs with json tags and TypeScript interfaces. The type checker
// sees only the TypeScript and the marshaller sees only the Go, so these checks compare them, field
// for field in both directions. They also hold the event words Go emits to a page, which the page
// must name exactly: a word one side alone renames reaches nothing and nothing fails.

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

var (
	tsInterface = regexp.MustCompile(`(?s)export interface (\w+) \{(.*?)\n\}`)
	tsField     = regexp.MustCompile(`(?m)^\s+(\w+)\??:`)
)

// goWire answers each struct in files with its json names, sorted.
func goWire(t testing.TB, files []string) map[string][]string {
	t.Helper()
	out := map[string][]string{}
	for _, file := range files {
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(parsed, func(node ast.Node) bool {
			spec, ok := node.(*ast.TypeSpec)
			if !ok {
				return true
			}
			fields, ok := spec.Type.(*ast.StructType)
			if !ok {
				return true
			}
			var names []string
			for _, field := range fields.Fields.List {
				var tag string
				if field.Tag != nil {
					tag, _ = strconv.Unquote(field.Tag.Value)
				}
				names = append(names, strings.Split(reflect.StructTag(tag).Get("json"), ",")[0])
			}
			slices.Sort(names)
			out[spec.Name.Name] = names
			return true
		})
	}
	return out
}

// tsWire answers each interface in files with its field names, sorted.
func tsWire(t testing.TB, files []string) map[string][]string {
	t.Helper()
	out := map[string][]string{}
	for _, file := range files {
		for _, match := range tsInterface.FindAllStringSubmatch(Read(t, file), -1) {
			var names []string
			for _, field := range tsField.FindAllStringSubmatch(match[2], -1) {
				names = append(names, field[1])
			}
			slices.Sort(names)
			out[match[1]] = names
		}
	}
	return out
}

// CheckTheWireIsStatedAlikeOnBothSides fails unless every struct in goFiles and every interface in
// tsFiles is one of pairs (Go name to TypeScript name), each pair stating the same fields.
func CheckTheWireIsStatedAlikeOnBothSides(t testing.TB, pairs map[string]string, goFiles, tsFiles []string) {
	t.Helper()
	goSide, tsSide := goWire(t, goFiles), tsWire(t, tsFiles)
	if len(goSide) != len(pairs) || len(tsSide) != len(pairs) {
		t.Errorf("the Go wire states %d types and the TypeScript %d; the pairs name %d", len(goSide), len(tsSide), len(pairs))
	}
	for goName, tsName := range pairs {
		if !slices.Equal(goSide[goName], tsSide[tsName]) {
			t.Errorf("%s has %v in Go but %s has %v in TypeScript", goName, goSide[goName], tsName, tsSide[tsName])
		}
	}
}

// CheckThePageNamesEveryWord fails for each string constant in goFile named in words whose value no
// file of pageFiles states in the shape given for it (a format with one %s). The shape rather than
// the bare quoted word, since a view, a panel or a test may share the word and would hide a
// listener or a key that no longer matches.
func CheckThePageNamesEveryWord(t testing.TB, goFile string, words map[string]string, pageFiles []string) {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), goFile, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var page strings.Builder
	for _, file := range pageFiles {
		page.WriteString(Read(t, file))
	}
	for word, shape := range words {
		object := parsed.Scope.Lookup(word)
		if object == nil {
			t.Fatalf("%s has no constant %s", goFile, word)
		}
		literal := object.Decl.(*ast.ValueSpec).Values[0].(*ast.BasicLit).Value
		value, _ := strconv.Unquote(literal)
		if !strings.Contains(page.String(), fmt.Sprintf(shape, value)) {
			t.Errorf("%s emits %q (%s) but the page never names it", goFile, value, word)
		}
	}
}
