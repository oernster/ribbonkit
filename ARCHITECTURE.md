# ribbonkit Architecture

The desktop half of a ribbon application for Windows, macOS and Linux: placement, the window, the
tray and menus, sign-in start, the update check, the setup program and the page's half that draws the
ribbon. The application supplies what the ribbon shows and its own names; the kit supplies
everything about the ribbon itself. The domain and application are the same code everywhere; each
platform's own half sits in files its build tags or names select.

A ribbon's one network request is its update check: only `infrastructure/update` imports a network
package; the structural tests hold the other ways out (a program started, a Windows library
loaded by name, a request from either page) to the ones they name. A release's addresses are taken
only as `https` on `github.com`.

FR, NFR and CON numbers are those of TimeRibbon's REQUIREMENTS.md, where each rule was first
specified.

## Invariant

`UI -> Application -> Domain <- Infrastructure`

Dependencies point inward. Every rule below is a test under `tests/structural`, written over the
mechanics in `structure` ([below](#outside-the-layers)); a guard not listed here does not exist.

| Invariant | Enforcing test | File |
|---|---|---|
| Domain imports nothing of the kit outside a domain | `TestDomainHasNoOutwardImports` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Domain is pure: no network, filesystem, process, random or tz package; no wall clock read | `TestDomainIsPure` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Application never imports infrastructure or Wails | `TestApplicationDoesNotImportInfrastructure` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| The UI imports no infrastructure, its tests included: it reaches the desktop through `shell.Desktop` | `TestTheUIDependsOnTheApplicationOnly` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Neither application nor infrastructure imports the UI | `TestNothingBelowTheUIImportsIt` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Infrastructure never imports Wails | `TestWailsStaysOutOfInfrastructure` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Nothing in the kit outside infrastructure imports both application and infrastructure: each application wires the kit in its own composition root. An infrastructure file importing both is an adapter implementing a port, not wiring | `TestNothingWiresTheApplicationToTheInfrastructure` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| No source file exceeds 400 lines: Go, the web half's TypeScript and CSS, the setup page | `TestNoFileExceedsLineLimit` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| No source file sits in the danger band of 381 to 400 lines | `TestNoFileInDangerBand` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Every exported type carries a doc comment | `TestEveryExportedTypeIsDocumented` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| The kit holds its four layers, `web`, `installer`, `structure` and `tests` and no other folder, so none escapes the layer rules | `TestTheKitHoldsOnlyItsLayersThePageAndTheSetupProgram` | [`folders_test.go`](tests/structural/folders_test.go) |
| The setup program imports nothing of the kit but the install policy | `TestTheSetupProgramReachesOnlyTheInstallPolicy` | [`folders_test.go`](tests/structural/folders_test.go) |
| The web half reaches nothing outside the kit and imports only its peer dependencies and its test tools | `TestTheWebHalfReachesNothingOutsideTheKit` | [`page_test.go`](tests/structural/page_test.go) |
| No Go file outside `infrastructure/update` imports `net`, `crypto/tls` or `golang.org/x/net`; the exemption names a directory that exists (NFR-S-1) | `TestOnlyTheUpdateCheckImportsANetworkPackage` | [`network_test.go`](tests/structural/network_test.go) |
| Only the files in `processStarters` start a program or hand an address to the desktop, each of them exists; no Go file names a Windows library outside `structure.SystemLibraries` (NFR-S-1) | `TestOnlyNamedFilesStartAProcess` | [`network_test.go`](tests/structural/network_test.go) |
| Neither page half uses a request API or names a web address, the SVG namespace aside (NFR-S-1) | `TestNeitherPageHalfMakesARequest` | [`network_test.go`](tests/structural/network_test.go) |
| The setup page loads every script beside it | `TestTheSetupPageLoadsEveryScript` | [`setup_test.go`](tests/structural/setup_test.go) |
| Every scheme the menus offer has a block in `colours.css` stating each of Classic's tokens (the problem colour aside); every block is offered (FR-611) | `TestEveryOfferedSchemeHasItsOwnCompleteBlock` | [`colours_test.go`](tests/structural/colours_test.go) |
| Classic's dark colours are the same under the system's dark mode as under a chosen dark theme | `TestClassicDarkIsTheSameUnderTheSystemAsWhenChosen` | [`colours_test.go`](tests/structural/colours_test.go) |
| The window's wire is stated alike in `ui/window/wire.go` and `web/wire.ts` | `TestTheWireIsStatedAlikeOnBothSides` | [`wire_test.go`](tests/structural/wire_test.go) |
| The page listens for every event the window emits and keys every panel it names | `TestThePageNamesEveryEventTheWindowEmits` | [`wire_test.go`](tests/structural/wire_test.go) |
| The setup page listens for every event the setup facade emits | `TestTheSetupPageNamesEveryEventSetupEmits` | [`wire_test.go`](tests/structural/wire_test.go) |
| Lines are counted as an editor numbers them; network packages, process starters, libraries and requests are recognised exactly; contrast is computed as WCAG 2.x states it | `TestLineCountCountsTheLinesAnEditorShows` and three more | [`structure_test.go`](structure/structure_test.go) |
| Every picture the setup page shows is one `installer.Pictures` asks for | `TestPicturesNamesEveryPictureThePageShows` | [`run_test.go`](installer/run_test.go) |

A kit that is its own module cannot import an application without a `require` in its `go.mod`, so
no test watches for that.

## Layers

- **Domain**, pure Go.
  - `placement`, in physical pixels: the default place, a stored placement restored at its
    monitor's DPI, the least move into a work area (`Clamp`, `Recover`), the ribbon's length (`Fit`),
    centring along a work area (`CentredAlong`) or against an edge (`AgainstEdge`), the tab (`Tab`,
    FR-614), the edge a ribbon stands flush against (`FlushAgainst`), the snap of a drop within
    `SnapReach` (`Snapped`, FR-410) and a place clear of other ribbons (`Clear`, FR-412). The pull
    out goes on the side away from the ribbon's edge, else on the side it already has while that
    side has room (`PullOutSideOf`), else the side with more room; `PullOutBeside` sizes it;
    `PullOutHeld` keeps it still while the grip is dragged (FR-623).
  - `ribbon`: the ribbon's own choices as one value, `Choices`: colour, orientation and its home
    edge (FR-409), theme, Always on top, the pin and the pin in effect (FR-619), the stay-on-top rule
    (FR-617), the skipped release, the placement, the last edge (FR-411), the pull out's side, the
    opacity bounds (FR-622), the scale bounds and `ScaleAfter`, the scale a drag of the grip has
    reached (FR-623). Also a cell's label: trimmed, cut to `MaxLabelLength` characters, the
    application's own default when nothing is left (`Label`).
  - `localtime`, for ribbons that show places: a time written in 24-hour or 12-hour form (`Text`),
    the mark naming a zone (`ZoneMark`), the next minute boundary (`NextRefresh`), the order places
    run in east from Greenwich (`EastFromGreenwich`) plus every time of a day in a format, for the
    page to measure the widest (`TimeSamples`, over `Distinct`). Instants and zones arrive as
    arguments.
  - `hover`: told the pointer arrived or left and the time, it answers whether an unpinned ribbon is
    open and when to ask again (FR-615, FR-616).
  - `identity`: the application's names, `App{Name, AppID}`, which the kit holds none of its own.
- **Application.**
  - `arranger`: arranging the ribbon (`Launch`, `Rearrange`, `Moved`, `ToEdge`, `ToLastEdge`,
    `Centred`, `Collapsed`), the scroll bar, the scale and the grip's preview. It asks its `Host`
    (the application) for the ribbon's choices and content read together and saves through the
    host's one save path. Its lock is never held while it calls the host. Every placement ends clear
    of any other ribbon its `Neighbours` port reports, its own pull out included, else at the
    opposite edge, never while the grip is dragged; the result is then held for the others (FR-412).
    A nil port is `NoNeighbours`, a ribbon alone. `EdgeMoves` asks ahead, saving nothing, whether
    `ToEdge` would move the ribbon from where it was last arranged, so a menu can grey an edge whose
    press would leave it where it stands.
  - `controls`: the ribbon's own use cases over the same `Host`: choosing its colour, orientation,
    theme, opacity, Always on top and the pin, a value not offered refused; start at sign-in through
    a `Startup` port; the update check and the release the user skipped. An application embeds
    `Controls` in its service beside the arranger.
  - `menus`: the menu model (`Item`), the actions every ribbon offers and their items with their
    words (show or hide, Settings, Colour, Orientation, Position, Always on top, Pin, Help, Exit),
    built from the ribbon's choices. Each application lists them in its own order among its own.
    An item may be `Disabled`; `Offered` greys each Position item whose edge would not move the
    ribbon. The window passes its own menu through it; an application passes its tray menu and its
    Settings choices through `Control.Offered`. Each desktop draws a disabled item greyed.
  - `release`: the update check's rules over a `Source` (FR-509).
  - `shell`: the desktop port, `shell.Desktop`, which `desktop` implements.
- **Infrastructure.** On every platform `system` (wall clock, ids), `update`, `iconscale`, `zones`
  (zone ids resolved with the tz database built in, each loaded once, an empty id and `Local`
  refused), `atomicfile` (a file replaced whole, every file a ribbon keeps written through it, with
  the longest name it can replace) and `settingsfile` (a ribbon's settings file: what every
  application's does alike, each application saying through a `Codec` what its own holds, with the
  ribbon's choices written under the kit's keys and a list read one entry at a time, an entry it
  could not read written back as found); per
  platform `monitors`, `startup`, `appdata`, `runlog`, `desktop` (tray, native menus, the ribbon's
  window, the end of a move, the desktop's broadcasts, the pointer, the browser opener) and
  `occupancy` (the folder every running ribbon shares). Windows only: `setup`, the install policy;
  `heldfile`, a file held open with no sharing, which only tests import.
  Linux only: `gtkmain`. macOS only: `cocoamain`. `platform` is what every composition root does
  for its platform before the window opens: on Linux, GTK sent through X11 at import; `Prepare`, the
  icon handed to the tray and a request to end from outside heard, on Linux and macOS; and
  `GeneratingBindings`, true only in the run `wails build` makes to generate bindings.
- **UI**: `window`, the ribbon's window as the page and the desktop see it.

### Outside the layers

- **`web`**, the page's half, an npm package (`@oernster/ribbonkit`) shipped as TypeScript source for
  the application's own build. It holds the bridge to the window's methods (`bridge.ts`, over a
  guarded call that answers null where Go refused and tells a refusal handler why; the application
  adds its own calls over the same guard), the window's half of the wire (`wire.ts`), the drag and the
  right-click menu (`drag.ts`), the opacity, the `devicePixelRatio` watch, the scroll bar's measure,
  the background colour reported to Go, a panel fitting its content (`panelFit.ts`), the corner grip
  (`ScaleGrip.tsx`) and the opacity slider (`OpacitySlider.tsx`) styled by `controls.css`. The ribbon
  itself: `Band.tsx` is the band the application draws its content in, with the tab, the drag, the
  wheel, the menu, the grip and the report that an opening ribbon has been drawn; `PullOut.tsx`
  places the band and what is pulled out beside it at the boxes Go sends (`Box`), with the handle
  named in the application's words; `ribbon.css` styles both. `shell.ts` (`useShell`) holds the
  application's snapshot, taken by a call the application passes in, routes the window's open-panel
  words to its panels, reloads on Go's refresh and draws the theme, colour scheme and opacity.
  Help, About and Licence are its panels (`Help.tsx`, `help.css`). The palette's ribbon half is
  `theme.css` (Classic) and `colours.css` (every other scheme). `web/testing` is the stand-in bridge
  and Go's events for an application's tests; `web/testing/setup.ts` is the one test set-up both run
  under. Each component takes only the values it draws; a module that reaches Go is handed the calls
  it needs.
- **`installer`**, the setup program's window over `setup`: the setup page (`page/`, no build step),
  the facade the page calls (`facade.go`) and `Run`, which opens the window. It is a program's
  window rather than a layer and imports nothing of the kit but the install policy.
- **`structure`**, the structural tests' mechanics: finding a repository's files, reading imports,
  counting lines and the layer, size, network, palette and wire checks over a `Layout` (its module,
  its layered folders, the modules it depends on and its composition root). The kit's own tests use
  it; so does each application, so both are held by the same code.

## What an application hands in

| To | What |
|---|---|
| everything that writes a name | `identity.App`, built once at the application's composition root |
| `update.New` | its GitHub repository as `owner/name`; the kit names none, so one product never asks after another's releases |
| `window.New` | `window.Config`: its `window.Service` (the application's half behind the window's port), the `shell.Desktop`, the log, the panel sizes, the `window.Product` (names, window class, version, author, copyright, credits, licence text, donation address) and `Act`, which carries out a menu action the kit does not know |
| `setup` | `setup.Product`: the `identity.App` and the publisher, from which the install folder, the executable, the shortcuts and the Apps list entry take their name |
| `installer.Run` | the payload and the page's pictures (`installer.Pictures` names those the page asks for) |

The object Wails binds embeds `*window.Window`, so every exported method of the window is page API:
Wails binds methods promoted from an embedded struct. What only the application may call is on
`window.Control`, a named field that is never embedded and so never bound.

## One window

Wails v2 offers one window, so the ribbon and every panel share it (CON-6). Opening a panel resizes
the window to its size in `PanelSizes`, centred on the ribbon's display within its work area. The
kit's panels are Settings, About, Licence and the update check's; an application adds its own in
`PanelSizes.Own` by the word its page names each by; `useShell` takes those words as its second
type parameter. A word of the kit's stays the kit's.
closing returns the ribbon to where it was. While a panel is open a move is not recorded and a change
of length waits for the close.

**A panel fits its content (FR-621).** Once a panel has opened, `panelFit.ts` measures it at its own
width and hands the height to `FitPanel`, which centres it again at that height, capped by the work
area. Measuring before the window became the panel took the ribbon's narrower window.

**Opacity (FR-622).** On Windows the web view is transparent and the window translucent, since Wails
otherwise paints the window solid behind the page; on macOS the web view is transparent; on Linux the
window is translucent (`run.go`). The window's own paint shows behind anything less than opaque, so
`opacity.go` paints it the page's colour at full opacity and clear below it.

**Scale (FR-623).** The grip sends `BeginScale`, `DragScale` and `EndScale`; Go turns the pointer's
travel into a scale through `ribbon.ScaleAfter`, fractional while dragging and rounded when kept. Go
reads the cursor from the desktop (`GetCursorPos`, `NSEvent mouseLocation`, GDK's seat pointer),
because the page's pointer events jumped backwards on Windows while the window resized under them;
where the desktop cannot answer, it takes the page's reading. A change of scale keeps the top-left
corner. While dragging, the pull out is held at its size and its place from the ribbon's corner
(`PullOutHeld`), so the window's corner stays still.

**Hidden until placed.** The window opens hidden; `startup` finds it, takes it off the taskbar, fences
its moves and places it first. On Windows it is found by its class and `HideFromTaskbar` swaps
Wails' application-window style for a tool window's. A launched ribbon is shown once the page has
reported its scale and widest text, else after `sizeWait`. Once shown, `keepLaunchedPlace` places it
again if the desktop put it somewhere else: GNOME may place a newly shown window by its own rule.

**The unpinned ribbon.** The window owns `hover`'s timer (`unpinned.go`). Opening tells the page
first and grows the window once the page reports `RibbonDrawn` (else after `drawWait`); the open
ribbon keeps the tab's frame; the window wears the page's colour. The pointer is read every 50 ms on
Windows (against the window's cut shape) and macOS; Linux hears GTK's crossing events instead, since
under XWayland neither the page nor the X server sees it leave.

**The pull out** shares the window: placements are decided for the ribbon alone and the pull out is
added beside it. On Windows the window is cut to the two (`placement.Shape`, `desktop.Shape` over
`SetWindowRgn`, FR-913) before every placing. On macOS and Linux it stays a rectangle.

## Place and drag

On Windows coordinates are physical pixels on the virtual desktop. Wails' `WindowSetPosition` is
relative to the current monitor's work area while `WindowGetPosition` is absolute; its screen list
has no origin, device name or work area, so displays are read through `EnumDisplayMonitors` and
`GetMonitorInfoW` and the window placed with `SetWindowPos` (CON-7). A drag's end is heard through a
WinEvent hook on `EVENT_SYSTEM_MOVESIZEEND`; a placement stores the monitor's device name, work area,
DPI and the offset from its corner, restored scaled by any change of DPI. With nothing stored (or the
monitor gone) the ribbon goes to its home edge on the primary.

A press on empty ribbon that moves past the desktop's drag distance (Windows' `SM_CXDRAG` and
`SM_CYDRAG`, GTK's `gtk-dnd-drag-threshold`, Windows' 4 DIP on macOS) is handed to the platform's
move loop through `window.WailsInvoke('drag')`, internal to Wails v2. On Windows a window procedure
in front of Wails' (`desktop.KeepOnDisplays`) keeps the rectangle inside the display under the
pointer.

**Two ribbons (FR-412).** `occupancy` keeps a folder every running ribbon shares (under
`%LOCALAPPDATA%`, `~/Library/Application Support` or `$XDG_RUNTIME_DIR`, each `ribbonkit`): one
`<id>.json` per product listing the rectangles it holds, live while its ribbon holds an OS lock on a
separate `<id>.lock`. An entry whose lock is free belongs to a ribbon that has gone and is removed. An
entry that cannot be believed (not JSON, over 64 KiB, over 16 rectangles) is said once in the log and
passed over. A Flatpak needs `--filesystem=xdg-run/ribbonkit:create`; a lock held in one sandbox was
measured to be seen from another.

## The desktop

On Windows `desktop` owns a hidden top-level window on its own locked thread for the tray icon, the
native menus and the broadcasts (a message-only window would not hear them). It re-adds the icon on
`TaskbarCreated`. It reports on a buffered channel, dropping an event with a log line rather than
blocking Windows' thread; the window's `listen` loop acts on it. Both recover a panic and log it.

Both menus are native popups, so the small window never clips them. The application builds their
items from `menus`; the window carries out every ribbon's actions itself and hands any other to the
application's `Act`. Settings offers every menu choice from the same items; `Choose` refuses anything
not offered (FR-624). A tray icon that cannot be made is not fatal.

**On Linux and macOS** both reach the desktop through cgo: GTK 3 and AppKit. What does not depend on
the toolkit is written once in `_unix.go` files.

- **One thread.** `gtkmain.Do` and `cocoamain.Do` run a function on the toolkit's loop and wait,
  raising a panic again on the caller.
- **Finding and hiding (FR-101).** The ribbon is the top-level window titled with the product's name.
  Linux marks it to skip the taskbar and switcher. macOS keeps its Dock icon: switching the
  application to an accessory after Wails launched it never removed the icon on a real Mac, so
  `HideFromTaskbar` does nothing there.
- **Coordinates.** Both count in DIP, every display reported at `placement.BaseDPI`; AppKit's
  bottom-left origin is turned over into the domain's top-left reckoning.
- **Placing.** macOS uses one `setFrame`. Linux sets the size and awaits it (up to 500 ms) before
  moving, since the window manager clamps a move by the size the window has when it arrives.
- **The end of a move (FR-404)** is the ribbon standing still for 300 ms, heard through
  `configure-event` or `NSWindowDidMoveNotification`; a position the kit chose is never a move.
- **Menus.** The Linux tray is the kit's own StatusNotifierItem with a `com.canonical.dbusmenu` menu
  over godbus, registered by object path so it needs no bus name or sandbox permission; macOS uses an
  `NSStatusItem` rebuilt as it opens.
- **Sign-in (FR-605).** Windows writes a value under `HKCU\...\Run`; Linux an XDG autostart entry
  (running `flatpak run` under the Flatpak); macOS a launchd agent.

## Help, About and Licence

About shows the application's picture, name, version, author, copyright line and its credits;
Licence shows the licence text the application hands in, its type sized so the widest line fits
(FR-608). Both read themselves when they overflow (FR-609) through one script,
`installer/page/auto-scroll.js`, shared with the setup page, which can import nothing;
`web/autoScroll.ts` types it and wraps it in a React hook.

## The update check

`update` asks GitHub's `releases/latest` for the repository the application names, which answers
only a published release, never a draft or prerelease. It is unauthenticated, times out after 5
seconds, never retries and reads at most a megabyte. `release.Check` compares the tag with the
version as dotted integers (anything else is never newer), picks this platform's asset by its ending
and honours the skipped release except on a manual check. The window's `updates.go` checks 3 seconds
after start, then every 24 hours, recovering any panic. The addresses stay in Go: Download and Skip
ask Go to act on what it offered.

## The setup program

`setup` is the install policy: the paths, the extraction with its fence against an entry leaving the
install folder (every entry checked before any is written, FR-803), the version comparison, the Apps
list record, the shortcuts (COM via go-ole), Start with Windows (through `startup`, the same value
Settings writes) and the step log. Everything written is per user (FR-810).

| Reading of the machine | Screen |
|---|---|
| started with `-uninstall` | Uninstall |
| nothing installed | Install |
| the version carried is newer | Update |
| the version carried is older | Go back |
| the versions match | Installed: Repair, Reinstall, Uninstall |

Uninstall deletes the settings only when **Also forget my settings** is ticked, then a hidden
PowerShell deletes the install folder once setup has exited. Setup offers to close a running copy
(FR-807). The setup page names nothing: the product's name arrives on its state. Its keyboard ring
(`setup-ring.js`) is the window's model written again, held by `setupRing.test.ts`.

## Errors

Errors are wrapped with `%w` at each boundary, so `errors.Is` finds a sentinel beneath. A page call Go
can refuse takes a refusal handler and answers null rather than rejecting, so a call without one does
not compile. A dropped desktop event, a panic in the desktop's thread or its handling and a failure to
place, hide from the taskbar or fence the ribbon are logged and the ribbon carries on.

## Quality enforcement

- The structural tests above run with the suite.
- `test.ps1` checks formatting, vet and staticcheck, runs the Go suite and the web half's lint, type
  check and tests, holds the domain and application to 100% and every other gated package to its
  measured floor ([TESTING.md](TESTING.md)).
- The Linux and macOS code is checked on its own platform ([TESTING.md](TESTING.md#on-macos-and-linux)).

## Design decisions

| Decision | Why | Rejected alternative |
|---|---|---|
| One repository for the Go module and the npm package, released under one tag | The window and its page half change together; one tag cannot pair a window with the wrong page | Two repositories, two version lines |
| The page half shipped as TypeScript source | The application's own build compiles it with its own settings | A built bundle per release |
| The kit names no product or repository | Two applications share it; a name written into the kit would be the wrong one for one of them | Constants in the kit |
| Displays and placement through each desktop's own calls | Wails' screen list lacks origin, device name and work area; its position calls mix relative and absolute coordinates | Wails' position calls |
| One window for the ribbon and every panel | Wails v2 offers one | A second window per panel |
| Native popup menus | A page-drawn menu would be clipped by the small window | A menu drawn in the page |
| The grip's cursor read from the desktop | The page's pointer events jumped backwards while the window resized under them | The page's `screenX`, `screenY` |
| An entry per product in a shared folder, live while its ribbon holds a lock on a separate file | Each Flatpak sandbox has private folders, so the folder must be one both are granted; on Windows a lock on the entry itself would block others reading it; a lock ends with its process, so a crashed ribbon's entry is known to be dead | One shared file listing process ids |
| A StatusNotifierItem of the kit's own | `fyne.io/systray` could not rebuild its menu as it opens and kept global state (v1.12.2) | `fyne.io/systray` |
| Toolkit-free code shared in `_unix.go` | Linux and macOS cannot drift apart | A copy per platform |
| The structural mechanics in the kit | An application and the kit are held by the same code | A copy of the checks per repository |

See also [TESTING.md](TESTING.md) and [DEVELOPMENT.md](DEVELOPMENT.md).
