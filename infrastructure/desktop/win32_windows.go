package desktop

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// Window messages, notify-icon values, menu flags and system metrics, named as the headers name
// them.
const (
	wmNull            = 0x0000
	wmDestroy         = 0x0002
	wmClose           = 0x0010
	wmTimeChange      = 0x001E
	wmDisplayChange   = 0x007E
	wmPowerBroadcast  = 0x0218
	wmLButtonUp       = 0x0202
	wmRButtonUp       = 0x0205
	wmApp             = 0x8000
	wmTrayCallback    = wmApp + 1
	pbtResumeSuspend  = 0x0007
	pbtResumeAuto     = 0x0012
	nimAdd            = 0x0
	nimDelete         = 0x2
	nifMessage        = 0x01
	nifIcon           = 0x02
	nifTip            = 0x04
	mfString          = 0x0000
	mfByCommand       = 0x0000
	mfGrayed          = 0x0001
	mfChecked         = 0x0008
	mfPopup           = 0x0010
	mfSeparator       = 0x0800
	tpmRightButton    = 0x0002
	tpmNonotify       = 0x0080
	tpmReturnCmd      = 0x0100
	idiApplication    = 32512
	trayIconID        = 1
	eventMoveSizeEnd  = 0x000B // EVENT_SYSTEM_MOVESIZEEND
	winEventOutOfCtx  = 0x0000 // WINEVENT_OUTOFCONTEXT
	gwlExStyle        = -20    // GWL_EXSTYLE
	gwlStyle          = -16    // GWL_STYLE
	wsExToolWindow    = 0x00000080
	wsExAppWindow     = 0x00040000
	wsCaption         = 0x00C00000
	wsSysMenu         = 0x00080000
	wsMinimizeBox     = 0x00020000
	wsPopup           = 0x80000000
	swpNoZOrder       = 0x0004
	swpNoActivate     = 0x0010
	swpFrameChanged   = 0x0020
	smCxDrag          = 68
	smCyDrag          = 69
	taskbarCreatedMsg = "TaskbarCreated"
	tipLength         = 128
	menuIDBase        = 1
	rgnOr             = 2 // RGN_OR
)

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	shell32  = windows.NewLazySystemDLL("shell32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
	gdi32    = windows.NewLazySystemDLL("gdi32.dll")

	procCreateRectRgn = gdi32.NewProc("CreateRectRgn")
	procCombineRgn    = gdi32.NewProc("CombineRgn")
	procDeleteObject  = gdi32.NewProc("DeleteObject")
	procSetWindowRgn  = user32.NewProc("SetWindowRgn")
	procGetWindowRgn  = user32.NewProc("GetWindowRgn")
	procPtInRegion    = gdi32.NewProc("PtInRegion")

	procShellNotifyIcon        = shell32.NewProc("Shell_NotifyIconW")
	procExtractIconEx          = shell32.NewProc("ExtractIconExW")
	procGetModuleHandle        = kernel32.NewProc("GetModuleHandleW")
	procRegisterClassEx        = user32.NewProc("RegisterClassExW")
	procCreateWindowEx         = user32.NewProc("CreateWindowExW")
	procDestroyWindow          = user32.NewProc("DestroyWindow")
	procDefWindowProc          = user32.NewProc("DefWindowProcW")
	procGetMessage             = user32.NewProc("GetMessageW")
	procTranslateMessage       = user32.NewProc("TranslateMessage")
	procDispatchMessage        = user32.NewProc("DispatchMessageW")
	procPostQuitMessage        = user32.NewProc("PostQuitMessage")
	procPostMessage            = user32.NewProc("PostMessageW")
	procCreatePopupMenu        = user32.NewProc("CreatePopupMenu")
	procAppendMenu             = user32.NewProc("AppendMenuW")
	procTrackPopupMenu         = user32.NewProc("TrackPopupMenu")
	procDestroyMenu            = user32.NewProc("DestroyMenu")
	procGetCursorPos           = user32.NewProc("GetCursorPos")
	procLoadIcon               = user32.NewProc("LoadIconW")
	procSetForegroundWindow    = user32.NewProc("SetForegroundWindow")
	procRegisterWindowMessage  = user32.NewProc("RegisterWindowMessageW")
	procSetWinEventHook        = user32.NewProc("SetWinEventHook")
	procUnhookWinEvent         = user32.NewProc("UnhookWinEvent")
	procFindWindow             = user32.NewProc("FindWindowW")
	procGetWindowLongPtr       = user32.NewProc("GetWindowLongPtrW")
	procSetWindowLongPtr       = user32.NewProc("SetWindowLongPtrW")
	procSetWindowPos           = user32.NewProc("SetWindowPos")
	procGetWindowRect          = user32.NewProc("GetWindowRect")
	procGetSystemMetricsForDpi = user32.NewProc("GetSystemMetricsForDpi")
	procGetCurrentProcessID    = kernel32.NewProc("GetCurrentProcessId")
)

// notifyIconData is NOTIFYICONDATAW; cbSize is the whole structure, so the shell reads the modern
// version.
type notifyIconData struct {
	cbSize            uint32
	hWnd              windows.HWND
	uID               uint32
	uFlags            uint32
	uCallbackMessage  uint32
	hIcon             windows.Handle
	szTip             [tipLength]uint16
	dwState           uint32
	dwStateMask       uint32
	szInfo            [256]uint16
	uVersionOrTimeout uint32
	szInfoTitle       [64]uint16
	dwInfoFlags       uint32
	guidItem          windows.GUID
	hBalloonIcon      windows.Handle
}

// wndClassEx is WNDCLASSEXW.
type wndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     windows.Handle
	hIcon         windows.Handle
	hCursor       windows.Handle
	hbrBackground windows.Handle
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       windows.Handle
}

// point is POINT; rect is RECT.
type (
	point struct{ x, y int32 }
	rect  struct{ left, top, right, bottom int32 }
)

// msg is MSG.
type msg struct {
	hwnd     windows.HWND
	message  uint32
	wParam   uintptr
	lParam   uintptr
	time     uint32
	pt       point
	lPrivate uint32
}

// utf16Pointer answers text as a NUL-terminated UTF-16 pointer for a Win32 call.
func utf16Pointer(text string) uintptr {
	pointer, err := windows.UTF16PtrFromString(text)
	if err != nil {
		return 0
	}
	return uintptr(unsafe.Pointer(pointer))
}
