//go:build windows

package setup

import (
	"archive/zip"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// FR-802.
func TestExtractZipWritesEveryEntry(t *testing.T) {
	t.Parallel()
	dest := filepath.Join(t.TempDir(), sample.App.Name)
	entries := map[string]string{sample.Exe(): "the program", "folder/": "", "folder/nested/a.txt": "a note"}
	if err := ExtractZip(zipOf(t, entries), dest); err != nil {
		t.Fatalf("extracting: %v", err)
	}
	for name, want := range map[string]string{sample.Exe(): "the program", "folder/nested/a.txt": "a note"} {
		got, err := os.ReadFile(filepath.Join(dest, filepath.FromSlash(name)))
		if err != nil || string(got) != want {
			t.Errorf("%s holds %q (%v), want %q", name, got, err, want)
		}
	}
}

// FR-803: every entry is checked before any is written, so a payload holding one bad entry
// writes nothing at all. The refusal names the entry.
func TestExtractZipRejectsAPathThatEscapes(t *testing.T) {
	t.Parallel()
	for _, bad := range []string{"../escaped.txt", "folder/../../escaped.txt", "C:/Windows/escaped.txt", "/rooted.txt", "NUL"} {
		dest := filepath.Join(t.TempDir(), sample.App.Name)
		payload := zipOf(t, map[string]string{"a-good-one.txt": "fine", bad: "no"})
		err := ExtractZip(payload, dest)
		if !errors.Is(err, ErrUnsafePath) || !strings.Contains(err.Error(), "unsafe path in payload") ||
			!strings.Contains(err.Error(), bad) {
			t.Errorf("%s: got %v, want the refusal naming it", bad, err)
		}
		if exists(dest) {
			t.Errorf("%s: something was written before the refusal", bad)
		}
	}
}

func TestAPayloadThatIsNotAnArchiveIsRefused(t *testing.T) {
	t.Parallel()
	if err := ExtractZip("not a zip", t.TempDir()); err == nil {
		t.Fatal("bytes that are not an archive extracted")
	}
	if _, err := Licence("not a zip"); err == nil {
		t.Fatal("a licence was read from bytes that are not an archive")
	}
}

// An entry that cannot be written stops the install rather than leaving a half-written program.
func TestAnEntryThatCannotBeWrittenStopsTheExtraction(t *testing.T) {
	t.Parallel()
	// A folder standing where the program must be written.
	dest := t.TempDir()
	write(t, filepath.Join(dest, sample.Exe(), "in the way"), "x")
	if err := ExtractZip(zipOf(t, map[string]string{sample.Exe(): "x"}), dest); err == nil {
		t.Errorf("%s was written over a folder", sample.Exe())
	}
	// A file standing where an entry's folder must be made, then where a folder entry must be.
	for _, name := range []string{"assets/a.txt", "assets/"} {
		dest := t.TempDir()
		write(t, filepath.Join(dest, "assets"), "a file where a folder must go")
		if err := ExtractZip(zipOf(t, map[string]string{name: ""}), dest); err == nil {
			t.Errorf("%s was made beneath a file", name)
		}
	}
	blocked := filepath.Join(t.TempDir(), "a file")
	write(t, blocked, "x")
	if err := ExtractZip(zipOf(t, map[string]string{"a.txt": "x"}), filepath.Join(blocked, sample.App.Name)); err == nil {
		t.Error("an install folder was made beneath a file")
	}
}

// The payload's licence is what the Licence screen shows; a payload with none says so.
func TestTheLicenceIsReadFromThePayload(t *testing.T) {
	t.Parallel()
	text, err := Licence(zipOf(t, map[string]string{LicenceFile: "the terms", sample.Exe(): "x"}))
	if err != nil || text != "the terms" {
		t.Errorf("got %q (%v)", text, err)
	}
	if _, err := Licence(zipOf(t, map[string]string{sample.Exe(): "x"})); !errors.Is(err, ErrNoLicence) {
		t.Errorf("a payload with no licence answered %v", err)
	}
}

func TestCopyFileReproducesTheContentAndSkipsItself(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	source, target := filepath.Join(dir, "setup.exe"), filepath.Join(dir, UninstallExeName)
	write(t, source, "the setup program")
	if err := CopyFile(source, target); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(target); string(got) != "the setup program" {
		t.Errorf("the copy holds %q", got)
	}
	// Run from the Apps list, setup is the uninstaller: copying it over itself does nothing.
	if err := CopyFile(target, target); err != nil {
		t.Errorf("copying the uninstaller over itself: %v", err)
	}
	if got, _ := os.ReadFile(target); string(got) != "the setup program" {
		t.Errorf("copying over itself left %q", got)
	}
	if err := CopyFile(filepath.Join(dir, "absent.exe"), filepath.Join(dir, "out.exe")); err == nil {
		t.Error("a source that is not there was copied")
	}
	if err := os.Mkdir(filepath.Join(dir, "blocked.exe"), dirPerm); err != nil {
		t.Fatal(err)
	}
	if err := CopyFile(source, filepath.Join(dir, "blocked.exe")); err == nil {
		t.Error("a copy was written over a folder")
	}
}

func TestDirSizeKBTotalsTheTree(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Two kilobytes and a byte, so the division to whole kilobytes is exercised.
	write(t, filepath.Join(dir, "nested", "big"), strings.Repeat("x", 2*bytesPerKB+1))
	if size, err := DirSizeKB(dir); err != nil || size != 2 {
		t.Errorf("size = %d KB (%v), want 2", size, err)
	}
	if _, err := DirSizeKB(filepath.Join(dir, "absent")); err == nil {
		t.Error("a folder that is not there was sized")
	}
}

func TestRemoveTreeDeletesTheWholeTreeAndToleratesItsAbsence(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "tree")
	write(t, filepath.Join(dir, "nested", "a.txt"), "x")
	if err := RemoveTree(dir); err != nil || exists(dir) {
		t.Fatalf("removing: %v, still there %v", err, exists(dir))
	}
	if err := RemoveTree(dir); err != nil {
		t.Errorf("removing what was already gone: %v", err)
	}
}

// The payload carries the application folder at the root, then the licence beside it.
func TestPackCarriesTheApplicationAndTheLicence(t *testing.T) {
	t.Parallel()
	app, dir := t.TempDir(), t.TempDir()
	write(t, filepath.Join(app, sample.Exe()), "the program")
	write(t, filepath.Join(app, "sub", "extra.txt"), "extra")
	licence := filepath.Join(dir, "LICENSE")
	write(t, licence, "the terms")
	var out bytes.Buffer
	if err := Pack(&out, Payload{Exe: sample.Exe(), App: app, Licence: licence}); err != nil {
		t.Fatal(err)
	}
	reader, err := zip.NewReader(bytes.NewReader(out.Bytes()), int64(out.Len()))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, file := range reader.File {
		names = append(names, file.Name)
	}
	if strings.Join(names, ",") != sample.Exe()+",sub/extra.txt,"+LicenceFile {
		t.Errorf("the payload holds %v", names)
	}
}

func TestPackRefusesWhatItCannotCarry(t *testing.T) {
	t.Parallel()
	empty := t.TempDir()
	if err := Pack(&bytes.Buffer{}, Payload{Exe: sample.Exe(), App: empty}); !errors.Is(err, ErrNoApplication) {
		t.Errorf("a folder with no application answered %v", err)
	}
	app := t.TempDir()
	write(t, filepath.Join(app, sample.Exe()), "x")
	if err := Pack(&bytes.Buffer{}, Payload{Exe: sample.Exe(), App: app, Licence: filepath.Join(app, "absent")}); err == nil {
		t.Error("a missing licence was packed")
	}
}
