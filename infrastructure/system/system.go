// Package system answers the wall clock and new clock ids: the two things the application may not
// make for itself (FR-207).
package system

import (
	"crypto/rand"
	"time"
)

// Clock is the application's Clock port over the Windows clock.
type Clock struct{}

// Now answers the current instant.
func (Clock) Now() time.Time { return time.Now() }

// IDs is the application's IDs port.
type IDs struct{}

// NewID answers a random id of 26 characters, never reused in practice.
func (IDs) NewID() string { return rand.Text() }
