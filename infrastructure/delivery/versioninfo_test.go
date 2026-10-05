package delivery

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// testAbout is a product as an application's product package describes it.
var testAbout = About{Name: "TestRibbon", Author: "A. Maker", Copyright: "© A. Maker"}

func TestTheResourceCarriesTheVersionAndTheIdentity(t *testing.T) {
	t.Parallel()
	for _, setup := range []bool{false, true} {
		body, err := versionResource("1.2.3", testAbout, setup)
		if err != nil {
			t.Fatal(err)
		}
		var r resource
		if err := json.Unmarshal(body, &r); err != nil {
			t.Fatal(err)
		}
		strings := r.Info[neutralLanguage]
		wantDescription := testAbout.Name
		if setup {
			wantDescription += setupSuffix
		}
		if r.Fixed.FileVersion != "1.2.3" || r.Fixed.ProductVersion != "1.2.3" || strings["ProductVersion"] != "1.2.3" ||
			strings["FileVersion"] != "1.2.3" || strings["ProductName"] != testAbout.Name ||
			strings["FileDescription"] != wantDescription || strings["CompanyName"] != testAbout.Author ||
			strings["LegalCopyright"] != testAbout.Copyright {
			t.Errorf("setup %v: %s", setup, body)
		}
	}
}

func TestAVersionThatIsNotMajorMinorPatchIsRefused(t *testing.T) {
	t.Parallel()
	for _, bad := range []string{"", "1.0", "1.0.0-dev", "v1.0.0"} {
		if _, err := versionResource(bad, testAbout, false); !errors.Is(err, ErrBadVersion) {
			t.Errorf("%q: got %v", bad, err)
		}
	}
}

func TestVersionInfoWritesTheFileMakingItsFolder(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "windows", "info.json")
	if err := VersionInfo([]string{"-version", "1.0.0", "-setup", "-out", out}, io.Discard, testAbout); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var r resource
	if err := json.Unmarshal(raw, &r); err != nil || r.Info[neutralLanguage]["FileDescription"] != testAbout.Name+setupSuffix {
		t.Errorf("%v: %s", err, raw)
	}
}

func TestVersionInfoRefusesWhatItCannotUse(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	blocked := filepath.Join(dir, "a file")
	if err := os.WriteFile(blocked, nil, fileMode); err != nil {
		t.Fatal(err)
	}
	cases := map[string][]string{
		"no -out":                 {"-version", "1.0.0"},
		"a bad version":           {"-version", "one", "-out", filepath.Join(dir, "a.json")},
		"an unknown flag":         {"-colour", "blue"},
		"an unwritable out":       {"-version", "1.0.0", "-out", dir},
		"a folder that is a file": {"-version", "1.0.0", "-out", filepath.Join(blocked, "info.json")},
	}
	for name, args := range cases {
		if err := VersionInfo(args, io.Discard, testAbout); err == nil {
			t.Errorf("%s was accepted", name)
		}
	}
}
