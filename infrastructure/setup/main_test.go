//go:build windows

package setup

// Shared fixtures. Nothing here reaches the real install folder, the real Start Menu or Desktop,
// the real Apps list entry or the real Run value: folders are t.TempDir() and every registry key is
// a scratch key under HKCU\Software\RibbonkitSetupTest, deleted when the test ends.

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows/registry"

	"github.com/oernster/ribbonkit/domain/identity/identitytest"
	"github.com/oernster/ribbonkit/infrastructure/startup"
)

// sample is the product every test here installs: the kit's sample application.
var sample = Product{App: identitytest.Sample, Publisher: "The Author"}

// standInVariable turns the test binary into a stand-in process: one that runs until its input
// closes, then exits. A copy of it plays the application or setup where a test needs a process to
// find, close or wait on.
const standInVariable = "RIBBONKIT_SETUP_STAND_IN"

// scratchParent holds every scratch key a test makes.
const scratchParent = `Software\RibbonkitSetupTest`

func TestMain(m *testing.M) {
	if os.Getenv(standInVariable) != "" {
		_, _ = io.Copy(io.Discard, os.Stdin)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// scratchKey answers a key of its own beneath HKCU\Software\RibbonkitSetupTest, deleted at the end.
func scratchKey(t *testing.T) string {
	t.Helper()
	key := scratchParent + `\` + rand.Text()
	t.Cleanup(func() {
		_ = registry.DeleteKey(registry.CURRENT_USER, key)
		_ = registry.DeleteKey(registry.CURRENT_USER, scratchParent)
	})
	return key
}

// fixture is a machine over temporary folders and scratch keys, with the removal of the install
// folder recorded rather than started.
type fixture struct {
	machine    Machine
	places     Places
	startupKey string
	record     Record
	scheduled  []string
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	root := t.TempDir()
	places := Places{
		InstallDir: filepath.Join(root, "Local", programsSubdir, sample.App.Name),
		StartMenu:  filepath.Join(root, "Roaming", "Programs"),
		Desktop:    filepath.Join(root, "Desktop"),
		Settings:   filepath.Join(root, "Roaming", sample.App.Name),
		Exe:        sample.Exe(),
	}
	record := AppsList(sample)
	record.key = scratchKey(t)
	f := &fixture{places: places, startupKey: scratchKey(t), record: record}
	entryFor := func(program string) StartupEntry { return startup.At(sample.App, f.startupKey, program) }
	f.machine = NewMachine(sample, places, f.record, entryFor, func(dir string) error {
		f.scheduled = append(f.scheduled, dir)
		return nil
	})
	return f
}

// zipOf builds an archive from a name-to-content map, as the string setup embeds its payload as.
func zipOf(t *testing.T, entries map[string]string) string {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, content := range entries {
		file, err := writer.Create(name)
		if err != nil {
			t.Fatalf("creating entry %q: %v", name, err)
		}
		if _, err := file.Write([]byte(content)); err != nil {
			t.Fatalf("writing entry %q: %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("closing the archive: %v", err)
	}
	return buffer.String()
}

// write puts content at path, making its folder.
func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), filePerm); err != nil {
		t.Fatal(err)
	}
}

// exists reports whether path is there.
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// standInCopy copies the test binary into a folder of its own under a name no other process
// carries, so a test can find and close it by image name without touching anything else.
func standInCopy(t *testing.T) string {
	t.Helper()
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	copied := filepath.Join(t.TempDir(), "tsprobe-"+rand.Text()+".exe")
	if err := CopyFile(self, copied); err != nil {
		t.Fatal(err)
	}
	return copied
}

// startStandIn starts program as a stand-in that runs until the answered function closes it.
func startStandIn(t *testing.T, program string) (pid int, closeIt func()) {
	t.Helper()
	cmd := exec.Command(program)
	cmd.Env = append(os.Environ(), standInVariable+"=1")
	input, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	closeIt = func() {
		_ = input.Close()
		_ = cmd.Wait()
	}
	t.Cleanup(closeIt)
	return cmd.Process.Pid, closeIt
}
