# Technical debt

What is still open, what is deliberately left and what only looks like debt.

Every item is a behaviour-preserving internal concern; nothing here reverts a feature or changes what
the user sees. `ARCHITECTURE.md` and the structural tests are the authority on the invariants.

Open items are numbered sections, so a scan for `## <number>.` tells whether the file is clear. The
two standing sections at the end are unnumbered. A resolved item is deleted outright; history belongs
in the release notes.

## 1. The setup program's facade has no tests

`Setup` in `installer/facade.go` is what the setup page calls. It holds a few decisions of its own:
the route answered when the machine cannot be read, the refusal while the application is running, the
line each act writes to the step log. None has a test; the package is gated at its measured 11%
(TESTING.md). It reaches the machine through `setup.Machine` and `setup.Processes`, which are concrete
types, so no test can stand in for them.

Cost of leaving it: low. Every act hands straight to `infrastructure/setup`, tested at 84%; each
decision is seen whenever setup is run by hand. Resolving it means giving the facade ports for the
machine and the processes, which its tests then fake. Not blocked.

## 2. A Linux and macOS test writes a panel's size in

`TestTheRibbonReturnsFromAPanelToWhereItIsPlaced` in `infrastructure/desktop/ribbon_unix_test.go`
writes the panel's size in as 560 by 760, which was TimeRibbon's About panel when the test was
written. The kit knows no application's panels, so the number can no longer be a copy of anything; it
stands for any window larger than the ribbon.

Cost of leaving it: low. The test still proves the ribbon returns to where it was placed after a
larger window. Blocked on an owner decision: whether this is now a sample size (which would make it
not debt) or whether the test should name its size as one.

## Looks like debt, not worth touching

**The drag sends Wails an internal message.** `startDrag` calls `window.WailsInvoke('drag')`, the
message Wails' own drag regions send: the one way to hand a press to the platform's own move loop.
The Wails version in `go.mod` pins it; check dragging by hand on every platform on any Wails upgrade.

**macOS borrows Windows' drag distance** of 4 DIP, since macOS publishes none.

**`placement.Fit` still makes room for one cell when handed none.** An application always counts at
least one cell, so that branch is the domain's own tested contract at the cost of one line.

## Not debt (do not "fix" these)

**The setup page's keyboard ring is the window's model written again.** The setup page has no build
step and can import nothing; `setupRing.test.ts` holds the shipped script to the same behaviour. The
self-reading cycle went the other way because the window's build can import from the setup page's
folder.

**The setup program holds no install logic.** Every act goes through `infrastructure/setup`.

**The browser opens through the desktop, not Wails.** Wails' `BrowserOpenURL` reports no error, so a
machine with no browser left a button silent. `desktop.OpenInBrowser` reports the refusal, which the
page shows with the address.

**Nothing holds the ribbon on a display during a drag on macOS and Linux.** Neither offers a say while
a drag lasts; a ribbon left partly off every display is put back when the move ends. A move loop of
our own would fight the desktop.

**Linux has a tray icon of its own.** `fyne.io/systray` was measured and rejected; going back would
lose a menu rebuilt as it opens.

**WebKit's DMABUF renderer is off on every Linux machine** (`gtkmain`). It drew a blank window on
NVIDIA's own driver; telling drivers apart would buy a faster path a ribbon never needs. A value the
user sets in `WEBKIT_DISABLE_DMABUF_RENDERER` is kept.

**The scroll bar's thickness comes from the page.** It is the web engine's bar, which Windows' own
metric does not describe.
