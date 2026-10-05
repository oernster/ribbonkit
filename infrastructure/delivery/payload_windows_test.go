package delivery

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/oernster/ribbonkit/domain/identity"
	"github.com/oernster/ribbonkit/infrastructure/setup"
)

// packs is the application the tests' payloads carry.
var packs = setup.Product{App: identity.App{Name: "TestRibbon", AppID: "uk.example.TestRibbon"}}

// put writes body at path.
func put(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), fileMode); err != nil {
		t.Fatal(err)
	}
}

func TestThePayloadHoldsTheApplicationAndItsLicence(t *testing.T) {
	t.Parallel()
	app, dir := t.TempDir(), t.TempDir()
	put(t, filepath.Join(app, packs.Exe()), "the program")
	licence := filepath.Join(dir, "LICENSE")
	put(t, licence, "terms")
	archive := filepath.Join(dir, "payload.zip")
	if err := Payload([]string{"-app", app, "-licence", licence, "-out", archive}, io.Discard, packs); err != nil {
		t.Fatal(err)
	}
	reader, err := zip.OpenReader(archive)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	if len(reader.File) != 2 || reader.File[0].Name != packs.Exe() || reader.File[1].Name != setup.LicenceFile {
		t.Errorf("the archive holds %v", reader.File)
	}
	if _, err := os.Stat(archive + packingSuffix); err == nil {
		t.Error("the packing file was left behind")
	}
}

// A packing that is refused leaves the archive that was there before.
func TestARefusedPackingLeavesTheOldArchive(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	archive := filepath.Join(dir, "payload.zip")
	put(t, archive, "the placeholder")
	err := Payload([]string{"-app", t.TempDir(), "-licence", "x", "-out", archive}, io.Discard, packs)
	if !errors.Is(err, setup.ErrNoApplication) {
		t.Errorf("got %v", err)
	}
	if got, _ := os.ReadFile(archive); string(got) != "the placeholder" {
		t.Errorf("the archive now holds %q", got)
	}
	if _, err := os.Stat(archive + packingSuffix); err == nil {
		t.Error("the packing file was left behind")
	}
}

func TestEveryPayloadFlagIsNeeded(t *testing.T) {
	t.Parallel()
	if err := Payload([]string{"-app", "a", "-out", "b"}, io.Discard, packs); !errors.Is(err, errPayloadFlags) {
		t.Errorf("got %v", err)
	}
	if err := Payload([]string{"-unknown"}, io.Discard, packs); err == nil {
		t.Error("an unknown flag was taken")
	}
	if err := Payload([]string{"-app", "a", "-licence", "l", "-out", filepath.Join(t.TempDir(), "absent", "p.zip")}, io.Discard, packs); err == nil {
		t.Error("an archive in a folder that is not there was written")
	}
}

// An archive that cannot be moved into place, because a folder stands at its name, is refused and
// the packing removed.
func TestAnArchiveThatCannotBeMovedIntoPlaceIsRefused(t *testing.T) {
	t.Parallel()
	app, dir := t.TempDir(), t.TempDir()
	put(t, filepath.Join(app, packs.Exe()), "the program")
	licence := filepath.Join(dir, "LICENSE")
	put(t, licence, "terms")
	archive := filepath.Join(dir, "payload.zip")
	if err := os.Mkdir(archive, folderMode); err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(archive, "kept"), "so a rename over the folder cannot succeed")
	if err := Payload([]string{"-app", app, "-licence", licence, "-out", archive}, io.Discard, packs); err == nil {
		t.Error("an archive was moved over a folder holding a file")
	}
}
