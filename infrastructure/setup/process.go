//go:build windows

package setup

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ErrAppRunning says the application is open, so nothing that would write over or delete its files
// may start (FR-807). Processes answers it after the application's name.
var ErrAppRunning = errors.New("is running")

// ErrStillRunning says the application was asked to close and was still there when the wait ran
// out. Processes answers it after the application's name.
var ErrStillRunning = errors.New("could not be closed; please close it by hand, then try again")

const (
	// closeTimeout is how long a running copy is given to go once asked (FR-807).
	closeTimeout = 5 * time.Second
	// pollStep is how often a wait looks again.
	pollStep = 100 * time.Millisecond
	// forcedExitCode is what a process ended by setup reports.
	forcedExitCode = 1
)

// Processes finds and ends the running copies of one executable, by its image name alone. Setup
// never ends a process TREE: descent is read from recorded parent ids, which churn, so setup could
// find itself counted a descendant and end itself.
type Processes struct {
	image string
	// name is the application's name, which its errors begin with.
	name string
	// end ends one process; terminate, unless a test stands in for a copy that will not go.
	end func(pid uint32)
	// wait is how long a copy is given to go once asked; closeTimeout, unless a test shortens it.
	wait time.Duration
}

// AppProcesses answers the running copies of product.
func AppProcesses(product Product) Processes {
	return Processes{image: product.Exe(), name: product.App.Name, end: terminate, wait: closeTimeout}
}

// Running reports whether any copy is running.
func (p Processes) Running() bool { return len(p.ids()) > 0 }

// Refusal answers ErrAppRunning after the application's name, for work refused while it runs.
func (p Processes) Refusal() error { return p.named(ErrAppRunning) }

// named answers err after the application's name, so it reads as a sentence about it.
func (p Processes) named(err error) error { return fmt.Errorf("%s %w", p.name, err) }

// Close ends every running copy and waits for them to go, answering ErrStillRunning when one is
// still there once its wait runs out (FR-807).
func (p Processes) Close() error {
	for _, pid := range p.ids() {
		p.end(pid)
	}
	deadline := time.Now().Add(p.wait)
	for p.Running() {
		if time.Now().After(deadline) {
			return p.named(ErrStillRunning)
		}
		time.Sleep(pollStep)
	}
	return nil
}

// ids answers the process ids running the image, compared without regard to case as Windows does.
// A snapshot that cannot be taken yields none, so a failing snapshot never blocks an install.
func (p Processes) ids() []uint32 {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil
	}
	defer windows.CloseHandle(snapshot)
	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	var found []uint32
	for err = windows.Process32First(snapshot, &entry); err == nil; err = windows.Process32Next(snapshot, &entry) {
		if strings.EqualFold(windows.UTF16ToString(entry.ExeFile[:]), p.image) {
			found = append(found, entry.ProcessID)
		}
	}
	return found
}

// terminate ends one process. A process already gone needs nothing more; one this user cannot open
// is not this user's to end. So every failure is let pass.
func terminate(pid uint32) {
	handle, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, pid)
	if err != nil {
		return
	}
	defer windows.CloseHandle(handle)
	_ = windows.TerminateProcess(handle, forcedExitCode)
}

// hidden keeps a child process from flashing a console over the window.
func hidden() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{HideWindow: true, CreationFlags: windows.CREATE_NO_WINDOW}
}

// user32 carries the window calls setup makes.
var (
	user32                  = windows.NewLazySystemDLL("user32.dll")
	procFindWindow          = user32.NewProc("FindWindowW")
	procIsWindowVisible     = user32.NewProc("IsWindowVisible")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procAllowSetForeground  = user32.NewProc("AllowSetForegroundWindow")
	procGetWindow           = user32.NewProc("GetWindow")
	procGetWindowThread     = user32.NewProc("GetWindowThreadProcessId")
	procAttachThreadInput   = user32.NewProc("AttachThreadInput")
	procSetFocus            = user32.NewProc("SetFocus")
)

// Launch starts program detached in its own folder, then waits up to wait for its window of class
// to show and brings it forward, so setup closes only once the application is in front: closing
// first hands the foreground back to whatever was behind and the application only flashes on the
// taskbar. A window that does not show in time is no failure; the program was started.
func Launch(program, class string, wait time.Duration) error {
	cmd := exec.Command(program)
	cmd.Dir = filepath.Dir(program)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting %s: %w", program, err)
	}
	_, _, _ = procAllowSetForeground.Call(uintptr(cmd.Process.Pid))
	_ = cmd.Process.Release()
	deadline := time.Now().Add(wait)
	for {
		// The deadline is read before the window is looked for, so nothing in the look can leave
		// the wait running for ever.
		if time.Now().After(deadline) {
			return nil
		}
		if window := visibleWindow(class); window != 0 {
			_, _, _ = procSetForegroundWindow.Call(window)
			return nil
		}
		time.Sleep(pollStep)
	}
}

// visibleWindow answers the top-level window of class when it is showing; zero otherwise.
func visibleWindow(class string) uintptr {
	name, err := windows.UTF16PtrFromString(class)
	if err != nil {
		return 0
	}
	window, _, _ := procFindWindow.Call(uintptr(unsafe.Pointer(name)), 0)
	if window == 0 {
		return 0
	}
	if visible, _, _ := procIsWindowVisible.Call(window); visible == 0 {
		return 0
	}
	return window
}

// gwChild asks GetWindow for a window's first child, which for a Wails window is the WebView2 host.
const gwChild = 5

// TakeFocus gives the web view inside the window of class the keyboard, reporting whether it could.
//
// Wails hands the web view its keyboard from the window's WM_SETFOCUS, raised only on a change of
// focus and handled inside an asynchronous callback, so a cold launch can lose the race and leave
// the page with no keyboard at all (measured in BridgeTalk's setup). SetFocus acts only within the
// caller's input queue, so the two threads' queues are joined for the call.
func TakeFocus(class string) bool {
	window := visibleWindow(class)
	if window == 0 {
		return false
	}
	child, _, _ := procGetWindow.Call(window, gwChild)
	if child == 0 {
		return false
	}
	windowThread, _, _ := procGetWindowThread.Call(window, 0)
	current := uintptr(windows.GetCurrentThreadId())
	attached := false
	if windowThread != 0 && windowThread != current {
		joined, _, _ := procAttachThreadInput.Call(current, windowThread, 1)
		attached = joined != 0
	}
	_, _, _ = procSetForegroundWindow.Call(window)
	focused, _, _ := procSetFocus.Call(child)
	if attached {
		_, _, _ = procAttachThreadInput.Call(current, windowThread, 0)
	}
	return focused != 0
}
