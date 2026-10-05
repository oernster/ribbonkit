// Package heldfile holds a file open with no sharing, as a backup tool, a sync client or a scanner
// may hold one at sign-in, so a test can show what a ribbon does with a file it can neither read nor
// replace. Windows alone refuses others while a file is held, so Hold exists there alone and its
// callers are tests built for Windows.
package heldfile
