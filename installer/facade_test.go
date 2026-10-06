//go:build windows

package installer

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/oernster/ribbonkit/infrastructure/setup"
)

// carriedVersion is the version every test's setup carries.
const carriedVersion = "2.0.0"

// fakeMachine answers what it is told to and records what it is asked for.
type fakeMachine struct {
	existing  setup.Existing
	readErr   error
	repairErr error
	places    setup.Places
	steps     []setup.Step
	installed []setup.Choices
	applied   []setup.Choices
	forgot    []bool
	asked     int
}

func (m *fakeMachine) Read() (setup.Existing, error) {
	m.asked++
	return m.existing, m.readErr
}

func (m *fakeMachine) Places() setup.Places { return m.places }

func (m *fakeMachine) InstallSteps(_ setup.Carried, choices setup.Choices) []setup.Step {
	m.installed = append(m.installed, choices)
	return m.steps
}

func (m *fakeMachine) RepairSteps(setup.Carried) ([]setup.Step, error) { return m.steps, m.repairErr }

func (m *fakeMachine) UninstallSteps(forget bool) []setup.Step {
	m.forgot = append(m.forgot, forget)
	return m.steps
}

func (m *fakeMachine) ChoiceSteps(choices setup.Choices) []setup.Step {
	m.applied = append(m.applied, choices)
	return m.steps
}

// fakeProcesses stands for the application's running copies.
type fakeProcesses struct {
	running  bool
	closeErr error
	closed   int
}

// errRunning is the refusal fakeProcesses answers.
var errRunning = errors.New("the application is running")

func (p *fakeProcesses) Running() bool  { return p.running }
func (p *fakeProcesses) Refusal() error { return errRunning }
func (p *fakeProcesses) Close() error {
	p.closed++
	return p.closeErr
}

// recordingLog keeps every line the facade writes.
type recordingLog struct{ lines []string }

func (l *recordingLog) Record(line string) { l.lines = append(l.lines, line) }
func (l *recordingLog) Path() string       { return "setup.log" }

// holds reports whether any line contains part.
func (l *recordingLog) holds(part string) bool {
	return slices.ContainsFunc(l.lines, func(line string) bool { return strings.Contains(line, part) })
}

// recordingShell keeps what the facade asked of its window.
type recordingShell struct {
	shown, quits int
	reported     []ProgressDTO
}

func (w *recordingShell) show()                  { w.shown++ }
func (w *recordingShell) progress(p ProgressDTO) { w.reported = append(w.reported, p) }
func (w *recordingShell) quit()                  { w.quits++ }

// rig is a facade over fakes, its window already started.
type rig struct {
	setup     *Setup
	machine   *fakeMachine
	processes *fakeProcesses
	log       *recordingLog
	window    *recordingShell
}

func newRig(t *testing.T, args ...string) rig {
	t.Helper()
	r := rig{machine: &fakeMachine{}, processes: &fakeProcesses{}, log: &recordingLog{}, window: &recordingShell{}}
	// The name is set through the field, since the setup program reaches the install policy alone.
	var product setup.Product
	product.App.Name = "Ribbon"
	r.setup = New(Config{
		Product:   product,
		SetupID:   "RibbonSetupTest",
		Machine:   r.machine,
		Processes: r.processes,
		Log:       r.log,
		Carried:   setup.Carried{Version: carriedVersion},
		Args:      args,
	})
	r.setup.useWindow(r.window)
	return r
}

// step answers a step named name that records it ran in ran, failing with err.
func step(name string, ran *[]string, err error) setup.Step {
	return setup.Step{Name: name, Weight: 1, Do: func() error {
		*ran = append(*ran, name)
		return err
	}}
}

func TestAMachineThatCouldNotBeReadIsTheVerdict(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	r.setup.problem = errors.New("no LOCALAPPDATA")
	state := r.setup.DetectState()
	if state.Problem != "no LOCALAPPDATA" || state.Route != "" || r.machine.asked != 0 {
		t.Errorf("state %+v after %d reads; want the problem, no route and no read", state, r.machine.asked)
	}
	r = newRig(t)
	r.machine.readErr = errors.New("registry refused")
	state = r.setup.DetectState()
	if state.Problem != "registry refused" || state.Route != "" || !r.log.holds("reading the machine: registry refused") {
		t.Errorf("state %+v, log %v; want the read's failure as the verdict, logged", state, r.log.lines)
	}
}

func TestTheRouteComesFromOneReadingOfTheMachine(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	r.machine.existing = setup.Existing{Installed: true, Version: "1.0.0", Choices: setup.Choices{Desktop: true}}
	state := r.setup.DetectState()
	if state.Route != string(setup.RouteUpdate) || state.InstalledVersion != "1.0.0" || state.ThisVersion != carriedVersion {
		t.Errorf("state %+v; want the update route from 1.0.0", state)
	}
	if state.StartMenu || !state.Desktop || state.StartWithWindows {
		t.Errorf("boxes %+v; want the machine's own", state)
	}
	if state.AppName != "Ribbon" || state.LogPath != "setup.log" || state.Uninstall {
		t.Errorf("state %+v; want the product's name, the log's path and no uninstall", state)
	}
	if !r.log.holds("route update, installed 1.0.0, carried " + carriedVersion) {
		t.Errorf("log %v; want the route recorded", r.log.lines)
	}
}

func TestUninstallOpensOnTheUninstallScreen(t *testing.T) {
	t.Parallel()
	if !newRig(t, setup.UninstallFlag).setup.DetectState().Uninstall {
		t.Error("started with the uninstall flag, setup did not open on Uninstall")
	}
}

func TestNothingIsTouchedWhileTheApplicationRuns(t *testing.T) {
	t.Parallel()
	acts := map[string]func(*Setup) error{
		"install":   func(s *Setup) error { return s.Install(ChoicesDTO{}) },
		"repair":    (*Setup).Repair,
		"uninstall": func(s *Setup) error { return s.Uninstall(false) },
	}
	for name, act := range acts {
		r := newRig(t)
		r.processes.running = true
		var ran []string
		r.machine.steps = []setup.Step{step("write", &ran, nil)}
		if err := act(r.setup); !errors.Is(err, errRunning) || len(ran) != 0 {
			t.Errorf("%s answered %v having run %v; want the refusal and nothing run", name, err, ran)
		}
		if !r.log.holds(name+" asked for") || !r.log.holds(errRunning.Error()) {
			t.Errorf("%s logged %v; want the ask and the refusal", name, r.log.lines)
		}
	}
}

func TestInstallRunsItsStepsWithTheBar(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	var ran []string
	r.machine.steps = []setup.Step{step("copy", &ran, nil), step("register", &ran, nil)}
	if err := r.setup.Install(ChoicesDTO{StartMenu: true, StartWithWindows: true}); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(ran, []string{"copy", "register"}) {
		t.Errorf("ran %v", ran)
	}
	want := setup.Choices{StartMenu: true, StartWithWindows: true}
	if !slices.Equal(r.machine.installed, []setup.Choices{want}) {
		t.Errorf("install asked with %v; want %v", r.machine.installed, want)
	}
	last := r.window.reported[len(r.window.reported)-1]
	if len(r.window.reported) != len(ran)+1 || r.window.reported[0].Msg != "copy" || last.Pct != 100 {
		t.Errorf("the bar reported %v; want each step then the whole", r.window.reported)
	}
}

func TestAFailedStepStopsTheWork(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	var ran []string
	broken := errors.New("disk full")
	r.machine.steps = []setup.Step{step("copy", &ran, broken), step("register", &ran, nil)}
	if err := r.setup.Uninstall(true); !errors.Is(err, broken) || !slices.Equal(ran, []string{"copy"}) {
		t.Errorf("answered %v having run %v; want the failure after the first step", err, ran)
	}
	if !slices.Equal(r.machine.forgot, []bool{true}) {
		t.Errorf("uninstall asked with forget %v; want true", r.machine.forgot)
	}
}

func TestARepairThatCannotBePlannedIsLogged(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	r.machine.repairErr = errors.New("payload unreadable")
	if err := r.setup.Repair(); !errors.Is(err, r.machine.repairErr) || !r.log.holds("payload unreadable") {
		t.Errorf("answered %v, logged %v; want the planning failure", err, r.log.lines)
	}
}

func TestTheBoxesApplyWithoutTheBar(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	var ran []string
	r.machine.steps = []setup.Step{step("shortcut", &ran, nil)}
	if err := r.setup.Apply(ChoicesDTO{Desktop: true}); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(r.machine.applied, []setup.Choices{{Desktop: true}}) || len(ran) != 1 || len(r.window.reported) != 0 {
		t.Errorf("applied %v, ran %v, bar %v; want the boxes applied with no bar", r.machine.applied, ran, r.window.reported)
	}
	if !r.log.holds("applying the boxes") {
		t.Errorf("log %v", r.log.lines)
	}
}

func TestClosingTheRunningCopyIsLogged(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	r.processes.running = true
	if !r.setup.AppRunning() {
		t.Error("a running copy was not reported")
	}
	r.processes.closeErr = errors.New("would not close")
	if err := r.setup.CloseRunningApp(); !errors.Is(err, r.processes.closeErr) || r.processes.closed != 1 {
		t.Errorf("answered %v after %d closes", err, r.processes.closed)
	}
	if !r.log.holds("closing the running copy") || !r.log.holds("would not close") {
		t.Errorf("log %v", r.log.lines)
	}
}

func TestAProgramThatWillNotStartIsLogged(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	r.machine.places = setup.Places{InstallDir: t.TempDir(), Exe: "absent.exe"}
	if err := r.setup.LaunchApp(); err == nil || !r.log.holds("starting Ribbon") || !r.log.holds(filepath.Join(r.machine.places.InstallDir, "absent.exe")) {
		t.Errorf("answered %v, logged %v; want the failure to start, logged", err, r.log.lines)
	}
}

func TestQuitClosesTheWindowAndShowsWhenFocusIsRefused(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	r.setup.Quit()
	// No window of the test's class exists, so the keyboard cannot be taken and the window is shown.
	r.setup.TakeKeyboard()
	if r.window.quits != 1 || r.window.shown != 1 || !r.log.holds("closed") {
		t.Errorf("quit %d, shown %d, log %v", r.window.quits, r.window.shown, r.log.lines)
	}
}

func TestBeforeStartupTheWindowIsLeftAlone(t *testing.T) {
	t.Parallel()
	s := New(Config{Log: &recordingLog{}})
	s.Quit()
	s.TakeKeyboard()
	s.progress(setup.Progress{Percent: 100})
	s.startup(context.Background())
	if _, started := s.windowNow().(wailsShell); !started {
		t.Errorf("after startup the window is %T; want Wails'", s.windowNow())
	}
}

func TestALicenceThatCannotBeReadIsAnError(t *testing.T) {
	t.Parallel()
	if _, err := newRig(t).setup.Licence(); err == nil {
		t.Error("an empty payload answered a licence")
	}
}
