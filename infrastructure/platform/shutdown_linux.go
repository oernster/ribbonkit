package platform

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/godbus/dbus/v5"
)

// logind's name, object and interface on the system bus; the signal it sends before the machine
// shuts down or restarts.
const (
	login1Name         = "org.freedesktop.login1"
	login1Path         = "/org/freedesktop/login1"
	login1Manager      = "org.freedesktop.login1.Manager"
	prepareForShutdown = login1Manager + ".PrepareForShutdown"
)

// The inhibitor lock asked of logind: held over shutdown in delay mode, so logind waits for the
// lock to go before shutting down, up to its own limit (InhibitDelayMaxSec).
const (
	inhibitWhat = "shutdown"
	inhibitWhy  = "Closing the ribbon before the desktop shuts down"
	inhibitMode = "delay"
)

// signalBuffer is how many bus signals may wait to be read.
const signalBuffer = 4

// watchShutdown ends the application before a shutdown or restart begins, sending SIGTERM on
// signals, which takes the path a request to end from outside always takes. At a restart systemd
// stops GNOME Shell and the ribbon together; the ribbon's window then went while GNOME Shell was
// losing Xwayland, so GNOME Shell waited for Xwayland's answer until it was killed (measured
// 2026-10-06 from a core dump: meta_x11_display_get_current_time_roundtrip, handling TimeRibbon's
// UnmapNotify). Leaving when logind announces the shutdown, while both are well, avoids it. A
// delay lock makes logind wait for that; it is released when the process ends. Where the system
// bus or logind cannot be reached, as in a Flatpak not granted it, the ribbon carries on as before.
func watchShutdown(signals chan<- os.Signal, log io.Writer) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		fmt.Fprintf(log, "shutdown watch: the system bus cannot be reached: %v\n", err)
		return
	}
	var lock dbus.UnixFD
	who := filepath.Base(os.Args[0])
	err = conn.Object(login1Name, login1Path).
		Call(login1Manager+".Inhibit", 0, inhibitWhat, who, inhibitWhy, inhibitMode).Store(&lock)
	if err != nil {
		fmt.Fprintf(log, "shutdown watch: logind gave no delay lock: %v\n", err)
	}
	if err := conn.AddMatchSignal(
		dbus.WithMatchObjectPath(login1Path),
		dbus.WithMatchInterface(login1Manager),
		dbus.WithMatchMember("PrepareForShutdown"),
	); err != nil {
		fmt.Fprintf(log, "shutdown watch: logind's shutdown signal cannot be heard: %v\n", err)
		return
	}
	incoming := make(chan *dbus.Signal, signalBuffer)
	conn.Signal(incoming)
	go func() {
		for each := range incoming {
			if shutdownStarting(each) {
				fmt.Fprintln(log, "shutdown watch: the machine is shutting down; ending")
				signals <- syscall.SIGTERM
				return
			}
		}
	}()
}

// shutdownStarting answers whether s is logind announcing that a shutdown or restart is about to
// begin; it sends the same signal with false when one is cancelled.
func shutdownStarting(s *dbus.Signal) bool {
	if s == nil || s.Name != prepareForShutdown || len(s.Body) != 1 {
		return false
	}
	starting, ok := s.Body[0].(bool)
	return ok && starting
}
