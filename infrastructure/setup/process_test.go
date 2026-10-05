//go:build windows

package setup

// Processes are found and closed by image name (FR-807). Every process started here is a copy of
// the test binary under a name nothing else carries, so no real application is ever found or ended.

import (
	"crypto/rand"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// deletionDeadline bounds the wait for a removal started cold by a fresh PowerShell.
const deletionDeadline = 30 * time.Second

func TestARunningCopyIsFoundAndClosedByItsImageName(t *testing.T) {
	program := standInCopy(t)
	startStandIn(t, program)
	copies := AppProcesses(sample)
	copies.image = strings.ToUpper(filepath.Base(program))
	if !copies.Running() {
		t.Fatal("the running stand-in was not found by its name in upper case")
	}
	if err := copies.Close(); err != nil {
		t.Fatal(err)
	}
	if copies.Running() {
		t.Error("the stand-in is still running after closing")
	}
	if (Processes{image: "no-process-is-named-" + rand.Text() + ".exe"}).Running() {
		t.Error("a name nothing carries was found running")
	}
	if AppProcesses(sample).image != sample.Exe() || AppProcesses(sample).wait != closeTimeout {
		t.Error("setup looks for another program or gives it another wait")
	}
	refusal := AppProcesses(sample).Refusal()
	if !errors.Is(refusal, ErrAppRunning) || refusal.Error() != sample.App.Name+" is running" {
		t.Errorf("the refusal is %q, want ErrAppRunning reading %q", refusal, sample.App.Name+" is running")
	}
}

// FR-807: a copy still running when the wait runs out is reported, asking for it to be closed by
// hand. The stand-in really runs; ending it is refused, as it is for a copy another account or an
// elevated one is running, which setup cannot open.
func TestACopyThatWillNotCloseIsReported(t *testing.T) {
	program := standInCopy(t)
	startStandIn(t, program)
	copies := AppProcesses(sample)
	copies.image = filepath.Base(program)
	copies.end = func(uint32) {}
	copies.wait = pollStep
	started := time.Now()
	err := copies.Close()
	if !errors.Is(err, ErrStillRunning) {
		t.Fatalf("Close answered %v, want ErrStillRunning", err)
	}
	// The page shows the words as they come, so they read as a sentence about the application.
	if want := sample.App.Name + " could not be closed; please close it by hand, then try again"; err.Error() != want {
		t.Errorf("Close said %q, want %q", err, want)
	}
	if waited := time.Since(started); waited < copies.wait {
		t.Errorf("reported after %v, before the wait of %v ran out", waited, copies.wait)
	}
	if !copies.Running() {
		t.Error("the stand-in went, so nothing refused to close")
	}
}

// Launch starts the program and returns once the wait for its window runs out; a program that
// cannot be started says so.
func TestLaunchStartsTheProgramAndBoundsTheWait(t *testing.T) {
	t.Setenv(standInVariable, "1")
	program := standInCopy(t)
	started := time.Now()
	if err := Launch(program, "NoSuchClass"+rand.Text(), pollStep); err != nil {
		t.Fatal(err)
	}
	if waited := time.Since(started); waited > deletionDeadline {
		t.Errorf("the wait ran for %v", waited)
	}
	// The stand-in reads no input and exits; its folder can go only once it has.
	launched := Processes{image: filepath.Base(program)}
	for deadline := time.Now().Add(deletionDeadline); launched.Running() && time.Now().Before(deadline); {
		time.Sleep(pollStep)
	}
	if err := Launch(filepath.Join(t.TempDir(), "absent.exe"), "x", pollStep); err == nil {
		t.Error("a program that is not there was started")
	}
	if TakeFocus("NoSuchClass" + rand.Text()) {
		t.Error("focus was given to a window that does not exist")
	}
}

// FR-806: the install folder goes once setup has closed; not before.
func TestTheInstallFolderGoesOnceSetupHasClosed(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "Awkward $name `n O'Brien", sample.App.Name)
	write(t, filepath.Join(dir, sample.Exe()), "x")
	pid, closeSetup := startStandIn(t, standInCopy(t))
	if err := deleteAfter(pid, dir); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)
	if !exists(dir) {
		t.Fatal("the folder went while setup was still running")
	}
	closeSetup()
	deadline := time.Now().Add(deletionDeadline)
	for exists(dir) && time.Now().Before(deadline) {
		time.Sleep(pollStep)
	}
	if exists(dir) {
		t.Error("the folder is still there after setup closed")
	}
	if !exists(filepath.Dir(dir)) {
		t.Error("the folder around it went too")
	}
}

// The folder reaches the removal as an environment value, never typed into the script.
func TestTheRemovalIsToldTheFolderAsAValue(t *testing.T) {
	t.Parallel()
	args, env := dirDeletion(42, `C:\a $b\SampleRibbon`)
	script := args[len(args)-1]
	if !strings.Contains(script, "Wait-Process -Id 42") || strings.Contains(script, `C:\a $b`) {
		t.Errorf("script %q", script)
	}
	if len(env) != 1 || env[0] != deletionDirVariable+`=C:\a $b\SampleRibbon` {
		t.Errorf("env %v", env)
	}
}

// The theme value reads dark only where it says so; anything missing reads light.
func TestTheWindowsThemeIsRead(t *testing.T) {
	key := scratchKey(t)
	if prefersDarkAt(key) {
		t.Error("a missing key read as dark")
	}
	opened, _, err := registry.CreateKey(registry.CURRENT_USER, key, registry.SET_VALUE)
	if err != nil {
		t.Fatal(err)
	}
	defer opened.Close()
	for value, dark := range map[uint32]bool{darkThemeValue: true, 1: false} {
		if err := opened.SetDWordValue(themeValueName, value); err != nil {
			t.Fatal(err)
		}
		if prefersDarkAt(key) != dark {
			t.Errorf("%d read as dark %v", value, !dark)
		}
	}
	_ = SystemPrefersDark()
}

// The window's ground is read from the sheet, light then dark; a sheet without both is refused.
func TestTheGroundsAreReadFromTheSheet(t *testing.T) {
	t.Parallel()
	light, dark, err := Surfaces(":root { --surface: #f3F6fb; } :root[x] { --surface: #05132a; }")
	if err != nil || light != (Ground{0xf3, 0xf6, 0xfb}) || dark != (Ground{0x05, 0x13, 0x2a}) {
		t.Errorf("got %v %v (%v)", light, dark, err)
	}
	if _, _, err := Surfaces(":root { --surface: #ffffff; }"); !errors.Is(err, ErrNoSurfaces) {
		t.Errorf("one surface answered %v", err)
	}
}

// The places come from the environment and the shell's known folders; any missing is refused.
func TestThePlacesComeFromTheEnvironmentAndTheShell(t *testing.T) {
	t.Parallel()
	env := map[string]string{localAppData: `C:\L`, "APPDATA": `C:\R`}
	lookup := func(name string) (string, bool) { value, ok := env[name]; return value, ok }
	known := func(id *windows.KNOWNFOLDERID, _ uint32) (string, error) {
		if id == windows.FOLDERID_Desktop {
			return `C:\D`, nil
		}
		return `C:\R\Programs`, nil
	}
	places, err := ResolvePlaces(sample, lookup, known)
	want := Places{InstallDir: `C:\L\Programs\SampleRibbon`, StartMenu: `C:\R\Programs`, Desktop: `C:\D`, Settings: `C:\R\SampleRibbon`, Exe: sample.Exe()}
	if err != nil || places != want {
		t.Errorf("got %+v (%v), want %+v", places, err, want)
	}
	refused := errors.New("refused")
	for name, broken := range map[string]func() (Places, error){
		"no LOCALAPPDATA": func() (Places, error) {
			return ResolvePlaces(sample, func(string) (string, bool) { return "", false }, known)
		},
		"no APPDATA": func() (Places, error) {
			return ResolvePlaces(sample, func(n string) (string, bool) { return `C:\L`, n == localAppData }, known)
		},
		"no Start Menu": func() (Places, error) {
			return ResolvePlaces(sample, lookup, func(*windows.KNOWNFOLDERID, uint32) (string, error) { return "", refused })
		},
		"no Desktop": func() (Places, error) {
			return ResolvePlaces(sample, lookup, func(id *windows.KNOWNFOLDERID, f uint32) (string, error) {
				if id == windows.FOLDERID_Desktop {
					return "", refused
				}
				return known(id, f)
			})
		},
	} {
		if _, err := broken(); err == nil {
			t.Errorf("%s was not refused", name)
		}
	}
	real, err := ResolvePlaces(sample, os.LookupEnv, windows.KnownFolderPath)
	if err != nil || !strings.HasSuffix(real.InstallDir, filepath.Join(programsSubdir, sample.App.Name)) {
		t.Errorf("this machine resolved %+v (%v)", real, err)
	}
}
