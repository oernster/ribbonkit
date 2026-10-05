//go:build windows

// Package setup holds the install policy behind a ribbon's setup program (FR-801 to FR-810):
// what the machine already holds, which conversation setup has with it and what installing,
// repairing and removing actually do. The setup program's window is a thin shell over it.
//
// Everything is per user. Files go under %LOCALAPPDATA%, shortcuts under %APPDATA% and the Desktop,
// records under HKCU, so Windows never asks for administrator rights (FR-810, CON-8).
package setup

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	// UninstallExeName is the copy of setup left in the install folder, which the Apps list runs
	// for Modify, Repair and Uninstall once the downloaded setup file is gone.
	UninstallExeName = "uninstall.exe"
	// LicenceFile is the licence the payload carries beside the application.
	LicenceFile = "LICENSE"
	// UninstallFlag opens setup on the Uninstall screen (FR-801). The Apps list passes it back to
	// the uninstaller copy, so the entry and the setup program both read it from here.
	UninstallFlag = "-uninstall"

	// dirPerm is how folders are made: the owner writes, everyone reads and enters.
	dirPerm = 0o755
	// filePerm is how files are written: the owner writes, everyone reads and runs.
	filePerm = 0o755
	// bytesPerKB turns a byte count into the kilobytes the Apps list's EstimatedSize holds.
	bytesPerKB = 1024
)

// ErrUnsafePath is answered when a payload entry names a path outside the install folder (FR-803).
var ErrUnsafePath = errors.New("unsafe path in payload")

// ErrNoLicence is answered when the payload carries no licence, as a build without one does.
var ErrNoLicence = errors.New("this setup program carries no licence")

// openPayload reads the embedded archive. It arrives as a string because the setup program embeds
// it as one, which Go keeps in the read-only image rather than charging to the process.
func openPayload(data string) (*zip.Reader, error) {
	reader, err := zip.NewReader(strings.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("opening the payload: %w", err)
	}
	return reader, nil
}

// ExtractZip writes every entry of the payload into dest (FR-802). Every entry is checked against
// dest before any is written, so a payload holding one that escapes writes nothing at all (FR-803).
func ExtractZip(data, dest string) error {
	reader, err := openPayload(data)
	if err != nil {
		return err
	}
	for _, file := range reader.File {
		if err := fenced(file.Name, dest); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(dest, dirPerm); err != nil {
		return fmt.Errorf("creating %s: %w", dest, err)
	}
	for _, file := range reader.File {
		if err := extractEntry(file, dest); err != nil {
			return err
		}
	}
	return nil
}

// fenced refuses an entry name that is not a plain path beneath dest: one that climbs out, one that
// is absolute or names a drive and one Windows reserves for a device.
func fenced(name, dest string) error {
	local := filepath.FromSlash(name)
	target := filepath.Join(dest, local)
	fence := filepath.Clean(dest) + string(os.PathSeparator)
	if !filepath.IsLocal(local) || !strings.HasPrefix(target+string(os.PathSeparator), fence) {
		return fmt.Errorf("%w: %s", ErrUnsafePath, name)
	}
	return nil
}

// extractEntry writes one entry already checked by fenced.
func extractEntry(file *zip.File, dest string) error {
	target := filepath.Join(dest, filepath.FromSlash(file.Name))
	if file.FileInfo().IsDir() {
		if err := os.MkdirAll(target, dirPerm); err != nil {
			return fmt.Errorf("creating %s: %w", target, err)
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(target), dirPerm); err != nil {
		return fmt.Errorf("creating %s: %w", filepath.Dir(target), err)
	}
	source, err := file.Open()
	if err != nil {
		return fmt.Errorf("opening entry %s: %w", file.Name, err)
	}
	defer source.Close()
	return writeFrom(source, target)
}

// writeFrom writes everything source holds into a file at target, replacing what was there.
func writeFrom(source io.Reader, target string) error {
	out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, filePerm)
	if err != nil {
		return fmt.Errorf("creating %s: %w", target, err)
	}
	if _, err := io.Copy(out, source); err != nil {
		_ = out.Close()
		return fmt.Errorf("writing %s: %w", target, err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("writing %s: %w", target, err)
	}
	return nil
}

// Licence answers the licence the payload carries, for the setup window's Licence screen.
func Licence(data string) (string, error) {
	reader, err := openPayload(data)
	if err != nil {
		return "", err
	}
	source, err := reader.Open(LicenceFile)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrNoLicence, err)
	}
	defer source.Close()
	text, err := io.ReadAll(source)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", LicenceFile, err)
	}
	return string(text), nil
}

// CopyFile copies src to dst, used to leave setup in the install folder as the uninstaller. Where
// src already is dst, as it is when the Apps list runs the uninstaller to repair, there is nothing
// to copy; a running program cannot be written over in any case.
func CopyFile(src, dst string) error {
	from, fromErr := os.Stat(src)
	to, toErr := os.Stat(dst)
	if fromErr == nil && toErr == nil && os.SameFile(from, to) {
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening %s: %w", src, err)
	}
	defer in.Close()
	return writeFrom(in, dst)
}

// DirSizeKB totals a folder tree in kilobytes, which is what the Apps list's EstimatedSize holds.
func DirSizeKB(dir string) (uint32, error) {
	var total int64
	err := filepath.WalkDir(dir, func(_ string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		total += info.Size()
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("measuring %s: %w", dir, err)
	}
	return uint32(total / bytesPerKB), nil
}

// RemoveTree deletes a folder with everything in it; one already gone is not an error.
func RemoveTree(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("removing %s: %w", dir, err)
	}
	return nil
}
