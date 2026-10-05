// Package delivery is the work behind every ribbon's build tools: the Windows version resource, the
// names the macOS and Linux scripts read, the Linux icon theme's sizes plus the setup program's
// payload on Windows. Each is a whole command, flags included, so an application's tool is a main
// that hands in what its own product package says and reports what comes back.
package delivery

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

// neutralLanguage keys the string table as language neutral with Unicode text, as Wails' own
// template does; Windows reads it back under the translation 000004b0.
const neutralLanguage = "0000"

// setupSuffix names the setup program after the product it installs.
const setupSuffix = " Setup"

// folderMode is how a missing output folder is made; fileMode how a file is written.
const (
	folderMode = 0o755
	fileMode   = 0o644
)

// semanticVersion is the form VERSION holds and every build script insists on.
var semanticVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)

// ErrBadVersion is answered for a version that is not major.minor.patch.
var ErrBadVersion = errors.New("the version is not major.minor.patch")

// errNoOut is answered when VersionInfo is not told where to write.
var errNoOut = errors.New("-out is required")

// About is what an executable's version resource says of the product: its name, who made it and
// the copyright line, each from the one home the application's product package gives it.
type About struct {
	Name      string
	Author    string
	Copyright string
}

// resource is the JSON winres reads: the fixed block and one string table per language.
type resource struct {
	Fixed struct {
		FileVersion    string `json:"file_version"`
		ProductVersion string `json:"product_version"`
	} `json:"fixed"`
	Info map[string]map[string]string `json:"info"`
}

// VersionInfo writes the Windows version resource wails build puts into an executable, from the
// version a build script read out of VERSION and about, so neither the version nor the author has
// a second home. Left to its own template, Wails wrote its fallback version, a placeholder copyright
// and its own advertisement, measured on 2026-09-27. Its flags:
//
//	-version 1.0.0 -out build/windows/info.json
//	-version 1.0.0 -setup -out installer/build/windows/info.json
func VersionInfo(args []string, errOut io.Writer, about About) error {
	flags := flag.NewFlagSet("versioninfo", flag.ContinueOnError)
	flags.SetOutput(errOut)
	version := flags.String("version", "", "the version, major.minor.patch")
	setup := flags.Bool("setup", false, "describe the setup program rather than the application")
	out := flags.String("out", "", "the info.json to write")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *out == "" {
		return errNoOut
	}
	body, err := versionResource(*version, about, *setup)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(*out), folderMode); err != nil {
		return fmt.Errorf("making %s: %w", filepath.Dir(*out), err)
	}
	if err := os.WriteFile(*out, body, fileMode); err != nil {
		return fmt.Errorf("writing %s: %w", *out, err)
	}
	return nil
}

// versionResource answers the resource for the application; for its setup program when setup is
// set.
func versionResource(version string, about About, setup bool) ([]byte, error) {
	if !semanticVersion.MatchString(version) {
		return nil, fmt.Errorf("%w: %q", ErrBadVersion, version)
	}
	description := about.Name
	if setup {
		description += setupSuffix
	}
	var r resource
	r.Fixed.FileVersion, r.Fixed.ProductVersion = version, version
	r.Info = map[string]map[string]string{neutralLanguage: {
		"ProductName":     about.Name,
		"ProductVersion":  version,
		"FileVersion":     version,
		"FileDescription": description,
		"CompanyName":     about.Author,
		"LegalCopyright":  about.Copyright,
	}}
	// A struct of strings and a map of strings always encodes, so there is no error to answer.
	body, _ := json.MarshalIndent(r, "", "\t")
	return append(body, '\n'), nil
}
