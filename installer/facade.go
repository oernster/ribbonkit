//go:build windows

// Package installer is a ribbon's setup program (FR-801 to FR-810): the window, its page and the
// facade the page calls, over the install policy in ribbonkit/infrastructure/setup. It names no
// product. The application's own setup command is its composition root: it carries the payload and
// the pictures the page shows, binds Setup by embedding it in a type of its own and calls Run.
package installer

import (
	"context"
	"slices"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/oernster/ribbonkit/infrastructure/setup"
)

const (
	// progressEvent carries each step's progress to the page.
	progressEvent = "progress"
	// launchWait bounds how long setup waits for the ribbon to show before it closes.
	launchWait = 5 * time.Second
)

// Machine is what the facade asks of the machine: setup.Machine in the setup program, a fake in the
// facade's tests.
type Machine interface {
	Read() (setup.Existing, error)
	Places() setup.Places
	InstallSteps(carried setup.Carried, choices setup.Choices) []setup.Step
	RepairSteps(carried setup.Carried) ([]setup.Step, error)
	UninstallSteps(forget bool) []setup.Step
	ChoiceSteps(choices setup.Choices) []setup.Step
}

// Processes is what the facade asks of the application's running copies: setup.Processes in the
// setup program.
type Processes interface {
	Running() bool
	Refusal() error
	Close() error
}

// Log is the step log the facade writes to: setup.StepLog in the setup program.
type Log interface {
	setup.Recorder
	Path() string
}

// shell is what the facade asks of the window it runs in.
type shell interface {
	show()
	progress(ProgressDTO)
	quit()
}

// unstarted is the shell before Wails has started the window: there is nothing yet to show, report
// to or close, so each call does nothing rather than hand Wails no context.
type unstarted struct{}

func (unstarted) show()                {}
func (unstarted) progress(ProgressDTO) {}
func (unstarted) quit()                {}

// wailsShell is the window Wails started, reached through the context it handed startup.
type wailsShell struct{ ctx context.Context }

func (w wailsShell) show() { wailsruntime.WindowShow(w.ctx) }
func (w wailsShell) progress(p ProgressDTO) {
	wailsruntime.EventsEmit(w.ctx, progressEvent, p)
}
func (w wailsShell) quit() { wailsruntime.Quit(w.ctx) }

// Config is what the setup window is built over.
type Config struct {
	// Product is the application this setup program installs.
	Product setup.Product
	// SetupID names what is setup's own: its window class, its web view cache and its step log.
	SetupID string
	// RibbonClass is the class the application's ribbon window is created with, which setup waits
	// for after starting it.
	RibbonClass string
	Machine     Machine
	Processes   Processes
	Log         Log
	Carried     setup.Carried
	Args        []string
	PrefersDark bool
	// Problem is why the machine could not be read at all; nil when it could.
	Problem error
}

// Setup is the facade the page calls: everything it can do goes through a method here. Each one
// hands straight to the setup package, which owns the install policy.
type Setup struct {
	window      shell
	product     setup.Product
	setupID     string
	ribbonClass string
	machine     Machine
	processes   Processes
	log         Log
	carried     setup.Carried
	uninstall   bool
	prefersDark bool
	problem     error
}

// New builds the facade. Started with -uninstall, setup opens on the Uninstall screen (FR-801).
func New(config Config) *Setup {
	return &Setup{
		window:      unstarted{},
		product:     config.Product,
		setupID:     config.SetupID,
		ribbonClass: config.RibbonClass,
		machine:     config.Machine,
		processes:   config.Processes,
		log:         config.Log,
		carried:     config.Carried,
		uninstall:   slices.Contains(config.Args, setup.UninstallFlag),
		prefersDark: config.PrefersDark,
		problem:     config.Problem,
	}
}

func (s *Setup) startup(ctx context.Context) { s.window = wailsShell{ctx: ctx} }

// domReady gives the page the keyboard once it exists, since a cold launch can lose the race that
// would otherwise hand it over.
func (s *Setup) domReady(context.Context) { s.TakeKeyboard() }

// TakeKeyboard gives the web view the keyboard; the page calls it when it finds it has none.
func (s *Setup) TakeKeyboard() {
	if !setup.TakeFocus(s.setupID) {
		s.window.show()
	}
}

// StateDTO is the one reading of the machine the page routes on. AppName travels with it because
// the page must not write the product's name down.
type StateDTO struct {
	AppName          string `json:"appName"`
	Route            string `json:"route"`
	Uninstall        bool   `json:"uninstall"`
	InstalledVersion string `json:"installedVersion"`
	ThisVersion      string `json:"thisVersion"`
	StartMenu        bool   `json:"startMenu"`
	Desktop          bool   `json:"desktop"`
	StartWithWindows bool   `json:"startWithWindows"`
	PrefersDark      bool   `json:"prefersDark"`
	LogPath          string `json:"logPath"`
	// Problem is why the machine could not be read; the page shows it as the failure verdict.
	Problem string `json:"problem"`
}

// ChoicesDTO carries the three boxes of FR-805 from the page.
type ChoicesDTO struct {
	StartMenu        bool `json:"startMenu"`
	Desktop          bool `json:"desktop"`
	StartWithWindows bool `json:"startWithWindows"`
}

func (c ChoicesDTO) choices() setup.Choices {
	return setup.Choices{StartMenu: c.StartMenu, Desktop: c.Desktop, StartWithWindows: c.StartWithWindows}
}

// ProgressDTO is emitted on the progress event while work runs.
type ProgressDTO struct {
	Pct int    `json:"pct"`
	Msg string `json:"msg"`
}

// DetectState reads the machine once and decides the route (FR-801).
func (s *Setup) DetectState() StateDTO {
	state := StateDTO{
		AppName:     s.product.App.Name,
		Uninstall:   s.uninstall,
		ThisVersion: s.carried.Version,
		PrefersDark: s.prefersDark,
		LogPath:     s.log.Path(),
	}
	if s.problem != nil {
		state.Problem = s.problem.Error()
		return state
	}
	existing, err := s.machine.Read()
	if err != nil {
		s.log.Record("reading the machine: " + err.Error())
		state.Problem = err.Error()
		return state
	}
	offered := setup.Offered(existing)
	state.Route = string(setup.RouteFor(existing, s.carried.Version))
	state.InstalledVersion = existing.Version
	state.StartMenu, state.Desktop, state.StartWithWindows = offered.StartMenu, offered.Desktop, offered.StartWithWindows
	s.log.Record("route " + state.Route + ", installed " + existing.Version + ", carried " + s.carried.Version)
	return state
}

// AppRunning reports whether the application is open, asked before any file is touched (FR-807).
func (s *Setup) AppRunning() bool { return s.processes.Running() }

// CloseRunningApp ends every running copy by image name and waits for them to go (FR-807).
func (s *Setup) CloseRunningApp() error {
	s.log.Record("closing the running copy")
	err := s.processes.Close()
	if err != nil {
		s.log.Record(err.Error())
	}
	return err
}

// Install performs an install, an update, a going back or a reinstall: all the same act (FR-802).
func (s *Setup) Install(choices ChoicesDTO) error {
	return s.perform("install", func() ([]setup.Step, error) {
		return s.machine.InstallSteps(s.carried, choices.choices()), nil
	})
}

// Repair writes the files again, keeping every box as it stands on the machine (FR-804).
func (s *Setup) Repair() error {
	return s.perform("repair", func() ([]setup.Step, error) { return s.machine.RepairSteps(s.carried) })
}

// Uninstall removes the application, forgetting the settings only when asked (FR-806).
func (s *Setup) Uninstall(forget bool) error {
	return s.perform("uninstall", func() ([]setup.Step, error) { return s.machine.UninstallSteps(forget), nil })
}

// Apply applies the boxes at once, as each changes on the Installed screen.
func (s *Setup) Apply(choices ChoicesDTO) error {
	s.log.Record("applying the boxes")
	return setup.Run(s.machine.ChoiceSteps(choices.choices()), s.log, func(setup.Progress) {})
}

// perform checks nothing is running, then runs the steps with the bar and the step log.
func (s *Setup) perform(name string, steps func() ([]setup.Step, error)) error {
	s.log.Record(name + " asked for")
	if s.processes.Running() {
		refusal := s.processes.Refusal()
		s.log.Record(refusal.Error())
		return refusal
	}
	list, err := steps()
	if err != nil {
		s.log.Record(err.Error())
		return err
	}
	return setup.Run(list, s.log, s.progress)
}

// LaunchApp starts the application and waits for the ribbon to come forward, so setup closes
// behind it.
func (s *Setup) LaunchApp() error {
	s.log.Record("starting " + s.product.App.Name)
	err := setup.Launch(s.machine.Places().Program(), s.ribbonClass, launchWait)
	if err != nil {
		s.log.Record(err.Error())
	}
	return err
}

// Licence answers the licence setup carries, for the Licence screen.
func (s *Setup) Licence() (string, error) { return setup.Licence(s.carried.Payload) }

// Quit closes the setup program.
func (s *Setup) Quit() {
	s.log.Record("closed")
	s.window.quit()
}

// progress reports how far the work has got.
func (s *Setup) progress(p setup.Progress) {
	s.window.progress(ProgressDTO{Pct: p.Percent, Msg: p.Step})
}
