package structural

// The kit's web half is installed into an application's page as a package, so it may reach nothing
// outside the kit: every relative import and every url() in web stays inside the repository (the
// self-reading cycle it imports lives beside the setup page, which can import nothing). Every
// package it names is one package.json states as a peer dependency (else a test tool an
// application's own test set-up supplies).

import (
	"encoding/json"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/oernster/ribbonkit/structure"
)

// kitTestTools are the packages the kit's suites and its testing stand-in run under, with why each
// is not among the kit's peer dependencies: an application's own test set-up supplies them.
var kitTestTools = map[string]string{
	"vitest":                 "the test runner every suite is written for",
	"@testing-library/react": "renders the kit's components in its suites",
}

var (
	// pageSpecifier finds a script import's module: from 'x', import 'x' and import('x').
	pageSpecifier = regexp.MustCompile(`(?:\bfrom|\bimport)\s*\(?\s*['"]([^'"]+)['"]`)
	// cssReference finds what a stylesheet reaches: url(x) and @import 'x'.
	cssReference = regexp.MustCompile(`url\(\s*['"]?([^'")]+)|@import\s+['"]([^'"]+)['"]`)
	// pageComment is a block comment or a whole-line one, whose words name modules without importing
	// them (index.ts says how an application imports the kit).
	pageComment = regexp.MustCompile(`(?s)/\*.*?\*/|(?m)^\s*//.*$`)
)

// kitPeers answers the packages package.json names as its peer dependencies.
func kitPeers(t *testing.T) map[string]bool {
	t.Helper()
	var stated struct {
		PeerDependencies map[string]string `json:"peerDependencies"`
	}
	if err := json.Unmarshal([]byte(structure.Read(t, filepath.Join(structure.Root(t), "package.json"))), &stated); err != nil {
		t.Fatal(err)
	}
	peers := map[string]bool{}
	for name := range stated.PeerDependencies {
		peers[name] = true
	}
	if len(peers) == 0 {
		t.Fatal("package.json states no peer dependencies, the read is wrong")
	}
	return peers
}

// packageOf answers the package a bare specifier names: its first part; two for a scoped one.
func packageOf(specifier string) string {
	parts := strings.Split(specifier, "/")
	if strings.HasPrefix(specifier, "@") && len(parts) > 1 {
		return parts[0] + "/" + parts[1]
	}
	return parts[0]
}

// referencesOf answers every module or file path reaches, by the form its extension is written in.
func referencesOf(t *testing.T, path string) []string {
	t.Helper()
	form := pageSpecifier
	if filepath.Ext(path) == ".css" {
		form = cssReference
	}
	var found []string
	for _, match := range form.FindAllStringSubmatch(pageComment.ReplaceAllString(structure.Read(t, path), ""), -1) {
		for _, group := range match[1:] {
			if group != "" {
				found = append(found, strings.TrimSpace(group))
			}
		}
	}
	return found
}

func TestTheWebHalfReachesNothingOutsideTheKit(t *testing.T) {
	root := structure.Root(t)
	peers := kitPeers(t)
	for _, path := range webFiles(t) {
		inside := structure.Relative(root, path)
		for _, reference := range referencesOf(t, path) {
			if strings.HasPrefix(reference, ".") {
				if target := structure.Relative(root, filepath.Join(filepath.Dir(path), reference)); strings.HasPrefix(target, "..") {
					t.Errorf("%s reaches %s, outside the kit", inside, reference)
				}
				continue
			}
			name := packageOf(reference)
			if _, tool := kitTestTools[name]; !peers[name] && !tool {
				t.Errorf("%s imports %s, which package.json does not depend on", inside, reference)
			}
		}
	}
}
