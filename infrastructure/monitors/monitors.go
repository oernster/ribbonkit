// Package monitors reads the displays (CON-7): each one's device name, work area, effective DPI
// and whether it is the primary. Wails' own screen list carries none of the first three, which is
// why this exists. Each platform supplies it in a file of its own.
package monitors
