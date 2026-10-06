# Decisions and trade-offs

The deliberate choices ribbonkit rests on, as the kit makes them today: what was chosen, what was
given up and why. The detail and the tests behind each live in [ARCHITECTURE.md](ARCHITECTURE.md)
and [TESTING.md](TESTING.md); [TECH_DEBT.md](TECH_DEBT.md) holds what is still open and what only
looks like debt.

## The kit as a whole

### A ribbon's desktop behaviour is a module of its own

Everything about being a ribbon on a desktop was carved out of the first ribbon whole: placing,
dragging, snapping, the tab, the grip, opacity, the tray and menus, one copy at a time, sign-in, the
update check, the run log and setup. Each application keeps only what its ribbon shows.

- **Rather than:** copying the desktop code into each new ribbon.
- **Gains:** a desktop fix lands once for every ribbon; two ribbons on one desktop can keep off each
  other because they share the code that places them.
- **Costs:** a change to the ribbon is made and tagged here before an application can ship it.

### One repository, one tag, for the Go module and the npm package

The window and its page half change together, so both live side by side and are released under the
same tag; an application names that one tag in its module and its front end alike.

- **Rather than:** two repositories with two version lines.
- **Gains:** a window can never be paired with the wrong page.
- **Costs:** a change to either half releases both; an application must keep its two references in
  step.

### The kit names no product

Every name, window class, repository and piece of artwork arrives from the application: its name
and id, its GitHub repository for the update check, its publisher for setup.

- **Rather than:** constants in the kit.
- **Gains:** a name written once can never be the wrong one for the other ribbon; one product never
  asks after another's releases.
- **Costs:** names chosen for any product rather than one; every entry point takes the application's
  names as an argument.

### Only for ribbons, only on Wails and React

The kit decides the window's shape, its place against an edge and how it hides. The window is a
Wails facade and the page half is React.

- **Rather than:** a general desktop toolkit; support for other frameworks.
- **Gains:** every decision can assume a ribbon, so none is a compromise between shapes.
- **Costs:** an application that is not a ribbon cannot use it; nor can one off Wails v2 and React.

### The page half shipped as source

- **Rather than:** a bundle built per release.
- **Gains:** the application's own build compiles it under its own settings, which may extend the
  kit's own type check, lint rules and test set-up.
- **Costs:** an application's build must be able to compile the kit's TypeScript.

### No stable API before 1.0

- **Rather than:** freezing the calls an application makes from the first tag.
- **Gains:** the kit can still be reshaped as the second ribbon finds what it needs.
- **Costs:** a minor release may change what an application calls.

## Boundaries and ports

### Layers with one place where they meet: the application's own

Domain, application, infrastructure and interface, each depending only inward. Nothing in the kit
outside its adapters imports both the application layer and the infrastructure: each application
wires the kit in its own composition root. Structural tests hold every boundary.

- **Rather than:** convention; the kit wiring itself.
- **Gains:** placement, hovering and the ribbon's choices are tested with no disk, network, clock or
  screen; an application decides what stands behind each port.
- **Costs:** more packages and explicit wiring in every application.

### The desktop behind one port

The window reaches the tray, the menus, the desktop's events and its own native window through one
port, which each platform's desktop implements. The interface layer imports no infrastructure, its
tests included.

- **Rather than:** the window calling the platform directly.
- **Gains:** every decision the window makes is tested over a stand-in desktop.
- **Costs:** the port grows with every new thing the ribbon asks of the desktop.

### What the page may call and what only the application may call are kept apart

Wails binds every method promoted from an embedded type, so every exported method of the window is
page API. What only the application may call sits on a separate hold that is never embedded.

- **Rather than:** one type serving both.
- **Gains:** the page can never reach a call meant for the application.
- **Costs:** two surfaces to keep in mind when adding a call.

### The wire written twice and compared; refusals that cannot be dropped

The shapes crossing between Go and the page are stated in both languages and compared by test, as
are the events the window emits and the panels the page keys. Every page call Go can refuse takes a
refusal handler and answers null rather than rejecting.

- **Rather than:** Wails' generated bindings; promises that reject.
- **Gains:** a contract both sides are checked against; a call without a handler does not compile,
  so no refusal is lost.
- **Costs:** every wire change is made twice; every call site names its handler.

### The structural checks are shared from the kit

The mechanics of the layer, size, network, palette, contrast, timer and wire checks live in a package
of the kit's own, run over a description of each repository's layout. The kit's own tests use it; so
does each application.

- **Rather than:** a copy of the checks per repository.
- **Gains:** an application and the kit are held by the same code; a fix to a check reaches both.
- **Costs:** a check must be written for any layout rather than one; the package's own coverage
  figure counts only its recognisers, since the calls that prove each check bite live elsewhere.

### The window's context is held atomically

Wails runs its start-up on a goroutine of its own and each bound call on another, nothing ordering
the two. The window's context and the setup facade's window are each held atomically; before
start-up the setup facade does nothing rather than hand Wails no context.

- **Rather than:** a plain field written at start-up.
- **Gains:** a race reproduced under the race detector is gone, on both windows.
- **Costs:** every read goes through an accessor.

## The window

### One window for the ribbon and every panel

Opening a panel resizes the window to it, centred on the ribbon's display; closing returns the
ribbon to where it was. An application adds panels of its own by the word its page names each by.

- **Rather than:** a second window per panel, which Wails does not offer.
- **Gains:** one window to place, hide and show.
- **Costs:** a panel replaces the ribbon while open; a move or change of length waits for it to
  close.

### A panel fits its content

Once a panel has opened, the page measures it at its own width and Go centres it again at that
height, capped by the work area.

- **Rather than:** fixed panel heights; measuring before the window became the panel, which took the
  ribbon's narrower window.
- **Gains:** on a display with room nothing scrolls.
- **Costs:** the window changes height as a panel's content changes.

### Displays and placing through the system's own calls

What macOS and Linux share, beyond their toolkits, is written once for both.

- **Rather than:** Wails' screen list and position calls, which know nothing of work areas and mix
  relative with absolute coordinates; a copy of the shared code per platform.
- **Gains:** the ribbon goes exactly where it is put; the two platforms cannot drift apart.
- **Costs:** three platforms' worth of desktop code; cgo on macOS and Linux.

### Hidden until placed, shown only once sized

The window opens hidden, is taken off the taskbar and placed, then shown once the page has reported
its scale and widest text. Once shown it is placed again if the desktop moved it.

- **Rather than:** showing it at once and growing it in view, which on a fractionally scaled Linux
  desktop often left it cut off.
- **Gains:** the ribbon first appears whole and where it belongs.
- **Costs:** it appears a moment later, up to a bounded wait where the page never reports.

### Closing hides; one copy runs; a launch toggles

One copy runs per user under the application's id; a second launch shows or hides the first.

- **Rather than:** a second launch only showing it; several copies.
- **Gains:** a single launcher button both shows and hides the ribbon.
- **Costs:** a copy left running makes a newer build's first launch only toggle the old ribbon.

### No taskbar button, no Dock icon

On Windows and Linux the window's taskbar button is removed before it is first shown. On macOS the
application becomes an accessory once Wails has finished launching it, so it has no Dock icon; its
menu-bar icon is how it is reached.

- **Rather than:** accepting the button everywhere; `LSUIElement` in the bundle, which Wails
  overrides as it launches.
- **Gains:** the ribbon stays out of the way; no Dock Quit reaches a close handler that hides rather
  than quits.
- **Costs:** on Windows it rests on swapping Wails' window style and on macOS on undoing Wails'
  activation policy, both checked by hand on any Wails upgrade; with no Dock icon there is no app
  switcher entry either.

### The early window veiled on Linux

Wails puts its window on screen before GTK's main loop starts, even when told to start hidden. A
window mapped before the loop runs is made fully transparent until it is next mapped.

- **Rather than:** a smaller starting size, which GTK holds at the web view's minimum (measured: 400
  by 400 pixels, still a black square); a transparent starting background, applied too late to change
  the first frame; patching Wails, whose v2.16.0 has the same code.
- **Gains:** no black window at login.
- **Costs:** a hook on every widget's map signal for the life of the process, which does nothing
  once the loop runs and the window is shown; it depends on Wails mapping before the loop, checked by
  hand on any Wails upgrade.

### See-through background, solid content

The opacity choice fades the ribbon's background between a faint floor and opaque, never invisible;
the content and every panel stay solid.

- **Rather than:** fading out entirely; fading the whole window.
- **Gains:** the ribbon sits over other work yet can always be found and read.
- **Costs:** the window is drawn translucent, so its own paint must be cleared below full opacity
  and the page must report its true surface colour.

### Resized by a grip, within bounds, from its corner

A corner grip scales everything in the ribbon together, between fixed bounds; it grows from its
top-left corner as a window does, with the pull out held still until let go. The pointer is read from
the desktop, the page's reading taken only where the desktop cannot answer.

- **Rather than:** free resizing; re-centring on every step; the page's pointer events, which jumped
  backwards while the window resized under them.
- **Gains:** no shape the layout was not made for; the ribbon holds still and follows the pointer
  smoothly.
- **Costs:** a ribbon centred on an edge is off-centre once resized, until it is next arranged; one
  pointer reading per platform to keep.

### The web view's data lives inside the settings folder

- **Rather than:** Wails' default, a folder named for the executable beside the settings folder.
- **Gains:** forgetting the settings at uninstall removes the web view's data with them.
- **Costs:** none recorded.

## Placing and dragging

### A home edge per orientation

A ribbon with nothing stored goes to its orientation's home edge on the primary display; so does one
whose display has gone. A stored place is kept as the display, its work area, its DPI and the offset
from its corner. It is restored scaled by any change of DPI.

- **Rather than:** an absolute position; one default for both orientations.
- **Gains:** a ribbon that has lost its place comes back somewhere predictable; a change of scaling
  keeps it where it was.
- **Costs:** none recorded.

### A move onto another display is not saved

The least move that keeps the ribbon inside a work area is saved only on the display its place was
saved on.

- **Rather than:** saving wherever the ribbon was pushed.
- **Gains:** a monitor that comes back finds the ribbon's placement kept.
- **Costs:** a ribbon pushed to another display returns to its old one at the next launch.

### Re-centred only when its length changes

Where its content changes its length, the ribbon is centred along that length on its work area, its
position across kept; one against the right or bottom edge stays flush. A change of scale does not
re-centre it.

- **Rather than:** growing from its corner on every change, which would pull it off the far edges.
- **Gains:** a ribbon centred on an edge stays centred as its content comes and goes.
- **Costs:** a ribbon dragged off-centre is re-centred at the next change of length.

### Dragging is the platform's own

A press that moves past the system's drag distance is handed to the platform's move loop through the
message Wails' drag regions send. Windows holds the ribbon on a display throughout; macOS and Linux
give no say, so it is put back once it stands still. macOS publishes no drag distance, so it borrows
Windows'.

- **Rather than:** a move loop of the kit's own, which would fight the desktop.
- **Gains:** the drag behaves as every other window's; a press on a control never starts one.
- **Costs:** the message is internal to Wails and checked by hand on any upgrade; off Windows the
  ribbon can hang off a display until let go.

### A drop near an edge snaps flush

- **Rather than:** leaving a drop exactly where it fell.
- **Gains:** a ribbon is easily put flush, which the tab needs.
- **Costs:** a drop near an edge snaps whether meant or not.

### Two ribbons never land on each other

Every running ribbon keeps an entry in a folder all ribbons share, listing the rectangles it holds;
the entry is live while its ribbon holds an operating system lock on a separate file. An entry whose
lock is free is removed; one that cannot be believed is logged once and passed over. Every placement
ends clear of the others, else at the opposite edge; none is moved aside while the grip is dragged.

- **Rather than:** one shared file listing process ids; each ribbon unaware of the others.
- **Gains:** a crashed ribbon's entry is known to be dead, since a lock ends with its process; on
  Windows others can still read an entry, since its lock is on another file; a hostile entry cannot
  cost more than a bounded read.
- **Costs:** a lock file per product stays behind; a Flatpak must be granted the shared folder.

### A pull out shares the window

What is pulled out stands beside the ribbon on the side away from its edge, else the side it already
has while that side has room, else the side with more room. Placements are decided for the ribbon
alone and the pull out is added beside it; on Windows the window is cut to the two.

- **Rather than:** a second window; a side fixed for good.
- **Gains:** one window; on Windows nothing is hidden that the ribbon does not draw.
- **Costs:** on macOS and Linux the window stays a rectangle, its spare area painted as the ribbon.

## The unpinned ribbon

### A thin tab, only against an edge, always on top

Unpinned, the ribbon shrinks to a thin tab shortly after the pointer leaves, only while flush
against an edge; elsewhere it shows in full with the choice kept. It stays above other windows and
opens on a resting pointer, keeping the tab's frame and the page's colour as it grows. An open menu
holds it open.

- **Rather than:** a tab wherever it stands; a tab big enough to press; one other windows can cover.
- **Gains:** the tab costs almost no room, is never lost and opens cleanly.
- **Costs:** touch reports no resting pointer, so touch users keep it pinned.

### When to open and close is a pure rule

Told when the pointer arrived or left and what the time is, the rule answers whether the ribbon is
open and when to ask again; it reads no clock. The window owns the timer.

- **Rather than:** timers scattered through the window.
- **Gains:** every hover case is tested without waiting.
- **Costs:** the window must ask again when told.

### The pointer is read as each platform was measured to need

Windows and macOS read the pointer at short intervals (Windows against the window's cut shape); Linux
listens for GTK's crossing events.

- **Rather than:** the page's own events, which were measured blind on an inactive macOS application
  and under XWayland.
- **Gains:** the tab opens and closes reliably everywhere.
- **Costs:** three ways of reading the pointer.

## Menus, choices and the tray

### Native menus from one item list

The ribbon's menu and the tray's are the system's own, built from items the kit makes for every
ribbon's actions with their words; each application lists them in its own order among its own. The
window carries out every ribbon's action itself and hands any other to the application.

- **Rather than:** a menu drawn in the page; each application writing the shared items again.
- **Gains:** the small window never clips a menu; the shared items read alike in every ribbon.
- **Costs:** each desktop draws the menu model natively, three times over.

### Every choice applies at once; a value not offered is refused

Settings offers the menus' choices from the same items. Choosing refuses anything the menus do not
offer.

- **Rather than:** two lists written separately; accepting any value.
- **Gains:** the menus and Settings cannot disagree; no call can set a value the ribbon never offers.
- **Costs:** a new choice has to fit both.

### A Position item is greyed only when pressing it would not move the ribbon

Beside another ribbon, putting the ribbon against an edge can leave it exactly where it stands. The
arranger answers ahead, saving nothing, whether it would move; that item alone is greyed in both menus
and in Settings.

- **Rather than:** offering a press that does nothing; greying every edge whose centre is taken,
  which would hide moves that still work.
- **Gains:** the ribbon never offers what cannot happen.
- **Costs:** whether a press would move it is worked out afresh each time a menu or Settings is
  built.

### A Linux tray icon of its own

The kit's own status item and menu over D-Bus, registered by object path.

- **Rather than:** a tray library that could not rebuild its menu as it opens and kept global state.
- **Gains:** the menu always shows the current ticks; no bus name or sandbox permission is needed.
- **Costs:** the tray code is the kit's own to maintain.

### Desktop events never block the desktop

The desktop hands events on without waiting, dropping one with a log line when nobody reads; both
the desktop's thread and the window's handling recover any panic. A tray icon that cannot be made is
not fatal.

- **Rather than:** calling back into the window on the desktop's thread.
- **Gains:** one fault cannot leave a ribbon that reacts to nothing.
- **Costs:** under a flood an event can be lost.

## Privacy and the network

### One network request, held by a test

The update check is the only code that may import a network package. Structural tests fail for any
other Go code that could open a connection, start a program or load a library by name outside a named
list; they fail too for any request or web address in either page.

- **Rather than:** a promise that the network is used sparingly.
- **Gains:** "nothing else touches the network" is a test result, for the kit and every application
  that runs the same checks.
- **Costs:** any new outward feature has to change the test that forbids it.

### The update check: shortly after start, then daily, quiet unless asked

The check asks GitHub for the latest published release without signing in, gives up quickly, never
retries and reads little. It speaks only of a newer release not skipped; a check from Help always
answers. A version it cannot read as dotted numbers is never newer. A release's addresses are taken
only as https on GitHub itself.

- **Rather than:** no check; one that reports every outcome.
- **Gains:** a new release is found without nagging; a draft, prerelease or malformed tag never
  prompts; a release cannot send the user elsewhere.
- **Costs:** one unprompted request to GitHub a day.

### It never installs an update itself

Download hands this platform's file, else the release's page, to the browser. The addresses stay in
Go: the page asks Go to act on what it offered.

- **Rather than:** downloading and running the new version from inside the application; the page
  holding the address.
- **Gains:** a ribbon never fetches or writes an executable; the page cannot be made to open an
  address Go did not offer.
- **Costs:** every update is a manual install.

### The browser opens through the desktop, not through Wails

- **Rather than:** Wails' own call, which reports nothing, so a machine with no browser left a
  button silent.
- **Gains:** a refusal is shown with the address.
- **Costs:** one opener per platform to keep.

## What a ribbon keeps

### Every file a ribbon keeps is replaced whole

Each is written to a temporary file beside the real one, flushed, then renamed over it; on any
failure the temporary file goes and the old copy stays.

- **Rather than:** writing in place.
- **Gains:** a crash mid-write leaves the old file whole.
- **Costs:** the longest name a file can have is shortened by the temporary name's length.

### One settings file, tolerant, never lost

What every ribbon's settings file does alike is the kit's; each application says what its own holds.
A bad value leaves its default and a bad entry of a list costs nothing else, kept as found. A file
that cannot be trusted is kept aside, never overwritten; one that is there but cannot be read is
never saved over. Keys a later version wrote are written back as found.

- **Rather than:** a database; each application writing its own reader; starting afresh over a
  damaged file.
- **Gains:** a person can read and repair it; the only copy of somebody's settings is never lost; an
  older version does not strip a newer one's keys.
- **Costs:** the ribbon's own choices are stored under the kit's keys in every application.

### A run leaves a log

The first act of a run points its error output at a log in the settings folder, which starts afresh
past a fixed size.

- **Rather than:** standard error, which a windowed program on Windows hands to nobody.
- **Gains:** a crash, the runtime's own panic report included, leaves a record rather than a silence.
- **Costs:** a log file per application that the user may never read.

### Time zones with the rules built in

Zone ids resolve with the tz database built into the executable, each loaded once; an empty id and
the machine's own zone are refused. Windows reads only the built-in rules; macOS and Linux read their
own first.

- **Rather than:** the machine's rules alone; letting an empty id read as UTC.
- **Gains:** every ribbon that shows places resolves them alike and offline.
- **Costs:** the rules add to the executable; on Windows a change of clocks shows only from the next
  release.

## Linux

### Linux on X11, with WebKit's faster paths off

GTK runs through X11, XWayland on Wayland. WebKit's DMABUF renderer is off unless the user chose a
value; hardware acceleration is off as Wails would choose.

- **Rather than:** Wayland, where a window may not choose where it stands; telling drivers apart.
- **Gains:** the ribbon stands where placed; NVIDIA's own driver draws the page.
- **Costs:** machines that could use the faster path lose it, which a ribbon never needs.

### Waiting for the size before moving

Linux sets the window's size and awaits it before moving, since the window manager clamps a move by
the size the window has when it arrives. The end of a move is the ribbon standing still for a moment.

- **Rather than:** setting size and place together, as macOS does.
- **Gains:** the ribbon lands where it was put.
- **Costs:** a placing can wait briefly; a move ends a moment after the pointer stops.

## The setup program

### Per user, with a setup program of its own, its policy apart from its window

Everything is written under the user's own folders and registry. One bespoke program installs,
updates, goes back, repairs, reinstalls and removes; its window holds no install logic and imports
nothing of the kit but the install policy. Every payload entry is checked to land inside the install
folder before any is written.

- **Rather than:** a machine-wide install; a generic installer.
- **Gains:** nothing asks for administrator rights; a hostile payload writes nothing; every decision
  of the policy is tested.
- **Costs:** each account installs separately; the window and its first reading of the machine are
  checked by hand.

### One reading of the machine decides the screen

Install, update, go back or repair is decided once from what is installed against what setup carries,
so the screen, its heading, its boxes and its buttons cannot drift apart. Removal is a screen
reachable from every route rather than a route of its own.

- **Rather than:** controls enabled and disabled in place.
- **Gains:** every route is a pure function of the reading and is tested; a machine that cannot be
  read is shown as the problem, never guessed at.
- **Costs:** the page cannot choose a screen of its own; every new screen starts as a route in Go.

### The facade is tested over ports of its own

The setup window reaches the machine, the running copies, the step log and its window through ports
the real parts satisfy unchanged.

- **Rather than:** a facade that touches the real machine directly.
- **Gains:** the route, the refusal while the application runs, the bar and a failed step are tested
  without changing the machine.
- **Costs:** four ports to keep beside the real parts.

### Nothing is touched while the application runs

Setup offers to close a running copy and does nothing until it has gone.

- **Rather than:** replacing files under a running program.
- **Gains:** no half-replaced install.
- **Costs:** the user must let the running ribbon close.

### The bar is weighted by measured time

Each step weighs what it was measured to take, so the bar moves as the work does.

- **Rather than:** weighting by step count, which showed a sixth of the bar once the files were
  written, nearly half the time.
- **Gains:** a bar that tells the truth.
- **Costs:** the weights are measurements that age as the payload grows.

### The settings are removed only when asked; the folder goes after setup

Uninstall forgets the settings only when the user ticks the box. The install folder holds the running
setup program, so a hidden PowerShell deletes it once setup has exited. Start with Windows is the same
value Settings writes.

- **Rather than:** always removing the settings; leaving the folder behind.
- **Gains:** an uninstall to reinstall keeps somebody's settings; forgetting them leaves nothing
  behind.
- **Costs:** a hidden process outlives setup for a moment.

### The setup page has no build step

The setup page is plain files that can import nothing. Its keyboard ring is the window's model
written again, held to the same behaviour by tests; the self-reading cycle for long pages is the
setup page's one script, which the window's build imports. The page names nothing: the product's name
arrives on its state.

- **Rather than:** a built page; the cycle written twice.
- **Gains:** the setup program carries its page as it is; one self-reading script serves both pages.
- **Costs:** the ring is kept twice; the shared script is plain JavaScript typed through a wrapper.

## Building and delivery

### The build tools' work and setup's wiring live in the kit

The version resource, the names the macOS and Linux scripts read, the Linux icon theme's sizes, the
setup program's payload and the whole of a setup program's wiring are commands an application's tools
hand its product's names. The icon generator is the kit's too, each application naming its own
artwork.

- **Rather than:** each application keeping its own copy.
- **Gains:** every ribbon is built and packaged alike; a fix reaches each.
- **Costs:** an application's tool must find the kit where Go builds it from.

### The generators hold no Go

- **Rather than:** generators written as Go commands inside the module.
- **Gains:** nothing outside the layer rules sits in the module.
- **Costs:** a generator needs Python on the machine that runs it.

## Engineering

### Complete coverage where it means something

The domain and application are held to 100 percent; every other gated package to the coverage it
measured, its fraction dropped. The desktop's macOS and Linux halves carry no floor.

- **Rather than:** one figure over everything; aspirational floors.
- **Gains:** a shortfall in the pure layers is a decision nobody made; every other one is named.
- **Costs:** the desktop code sits far lower and relies on checks by hand; a floor measured on one
  desktop would not hold on another, so macOS and Linux carry no figure.

### Small files

- **Rather than:** letting files grow.
- **Gains:** files split at real seams.
- **Costs:** many small files.

### Tests with real parts; every guard proved

No mocking library; hand-written doubles against real interfaces. No test reaches the network or the
user's own settings, sign-in entry or Apps list. Every guard is proved by planting a violation.

- **Rather than:** mocks; guards trusted because they pass.
- **Gains:** a passing test means the real behaviour holds; a guard never seen to fail is not
  counted as one.
- **Costs:** fakes are kept by hand; the macOS and Linux checks are a practice nothing enforces.
