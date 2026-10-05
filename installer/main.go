//go:build windows

package installer

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"

	"github.com/oernster/ribbonkit/infrastructure/setup"
)

// setupTitleSuffix follows the product's name in the setup window's title.
const setupTitleSuffix = " Setup"

// logExtension ends the step log's file name.
const logExtension = ".log"

// Program is what an application's setup command hands Main: its names (each from the one home its
// product package gives it) with what the command carries.
type Program struct {
	// Product is the application installed, with its publisher.
	Product setup.Product
	// SetupID names what is setup's own: its window class, its web view cache and its step log.
	SetupID string
	// RibbonClass is the class the application's window is created with, so setup knows it is up.
	RibbonClass string
	// Version is the version setup carries.
	Version string
	// Payload is the built application as a zip archive.
	Payload string
	// Pictures holds the page's pictures (Window.Pictures) under PicturesRoot.
	Pictures     fs.FS
	PicturesRoot string
	// Bind answers what Wails binds for the facade: a type named App in the command's own main
	// package embedding it, since Wails names a bound object by its package and type and the page
	// reaches the facade as main.App.
	Bind func(*Setup) any
}

// Main runs the setup program p describes (FR-801 to FR-810) and answers the exit code: zero once
// the window has closed, one where it could not be shown. Every step goes to a log in the temporary
// folder; where even that log could not be opened, the failure goes to standard error.
func Main(p Program) int {
	log, logErr := setup.OpenStepLog(filepath.Join(os.TempDir(), p.SetupID+logExtension))
	log.Record("setup " + p.Version + " started")
	places, placesErr := setup.ResolvePlaces(p.Product, os.LookupEnv, windows.KnownFolderPath)
	self, selfErr := os.Executable()
	problem := errors.Join(placesErr, selfErr)
	if problem != nil {
		log.Record("reading the machine: " + problem.Error())
	}
	facade := New(Config{
		Product:     p.Product,
		SetupID:     p.SetupID,
		RibbonClass: p.RibbonClass,
		Machine:     setup.NewMachine(p.Product, places, setup.AppsList(p.Product), setup.StartWithWindows(p.Product.App), setup.DeleteAfterExit),
		Processes:   setup.AppProcesses(p.Product),
		Log:         log,
		Carried:     setup.Carried{Payload: p.Payload, Self: self, Version: p.Version},
		Args:        os.Args[1:],
		PrefersDark: setup.SystemPrefersDark(),
		Problem:     problem,
	})
	if err := show(facade, p); err != nil {
		log.Record(err.Error())
		if logErr != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return 1
	}
	return 0
}

// show shows the setup window over facade with the product's title and p's pictures.
func show(facade *Setup, p Program) error {
	shown, err := fs.Sub(p.Pictures, p.PicturesRoot)
	if err != nil {
		return fmt.Errorf("reading the setup pictures: %w", err)
	}
	return Run(p.Bind(facade), facade, Window{Title: p.Product.App.Name + setupTitleSuffix, Pictures: shown})
}
