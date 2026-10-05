//go:build windows

package setup

// Each step's weight is the time it was measured to take, in milliseconds: the median of seven runs
// on 2026-09-27 with the real 0.1.0 payload (4,973,101 bytes) and the real setup program copied as
// the uninstaller, into temporary folders and scratch registry keys. A step measured below the
// timer's resolution weighs 1, so it still moves the bar.
//
// Weighted by step count an install's six steps would show a sixth of the bar once the files were
// written, which took nearly half the time; weighted by time it moves as the work does.
const (
	// weightExtract is writing the payload's files (49.18 ms).
	weightExtract = 49
	// weightUninstaller is copying setup in as the uninstaller (9.09 ms).
	weightUninstaller = 9
	// weightRecord is writing or deleting the Apps list entry (below resolution).
	weightRecord = 1
	// weightShortcut is placing one shortcut through the shell object (Start Menu 24.32 ms,
	// Desktop 21.42 ms).
	weightShortcut = 23
	// weightClearShortcut is removing one shortcut (below resolution).
	weightClearShortcut = 1
	// weightStartup is writing or deleting the Start with Windows value (below resolution).
	weightStartup = 1
	// weightForget is deleting the settings folder (below resolution).
	weightForget = 1
	// weightSchedule is starting the removal that runs once setup closes (13.67 ms).
	weightSchedule = 14
)
