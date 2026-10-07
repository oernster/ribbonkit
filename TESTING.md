# Testing

What is tested, what is not and why the line falls where it does. Every figure was measured by the
commands in [Running it](#running-it) when it was written; every shortfall is named with its reason.

## Before the first run on Windows

Some anti-virus programs quarantine a freshly built test binary in Go's scratch folder, which stops
the suite. If a run fails with access denied or a missing file on a test binary, allow the folder
`go env GOTMPDIR` names (the system temporary folder where it prints nothing). Install the web half's
tools once:

```powershell
npm install
```

## The standard

**A floor is a measurement, never an aspiration.** Each floor in `test.ps1` is its package's measured
figure with the fraction dropped, so it fails once cover is lost.

**A gap is named or it is closed.** An unexplained shortfall cannot be told from an oversight.

## What the numbers are

### Go

| Package | Coverage | Floor |
|---|---|---|
| `domain/hover`, `localtime`, `placement`, `ribbon` | 100% | 100% |
| `application/arranger`, `controls`, `menus`, `release` | 100% | 100% |
| `infrastructure/appdata`, `heldfile`, `iconscale`, `system`, `update`, `zones` | 100% | 100% |
| `infrastructure/settingsfile` | 97.3% | 97% |
| `ui/window` | 93.9% | 93% |
| `infrastructure/occupancy` | 90.1% | 90% |
| `infrastructure/delivery` | 96.5% | 96% |
| `infrastructure/atomicfile` | 85.7% | 85% |
| `infrastructure/setup` | 84.5% | 84% |
| `infrastructure/monitors` | 82.6% | 82% |
| `infrastructure/startup` | 80.6% | 80% |
| `infrastructure/runlog` | 77.8% | 77% |
| `installer` | 68.3% | 68% |
| `infrastructure/desktop` | 49.7% | 49% |
| `structure` | 33.6% | 33% |

`domain/identity` and `application/shell` hold types and ports alone, with no statement to cover.
On Windows `infrastructure/platform` holds a `Prepare` that does nothing and one constant, with no
test file; its Linux half is tested there ([On macOS and Linux](#on-macos-and-linux)).

Every figure is the Windows build's, which `test.ps1` measures. That build compiles 396 Go test
functions, counted from the test files `go list` selects, plus one `TestMain` in
`infrastructure/setup`. Twenty-four are the structural tests, which read the source and are the same
on every platform; [ARCHITECTURE.md](ARCHITECTURE.md) lists each against its rule. The macOS build
compiles 350 and the Linux build 356 ([On macOS and Linux](#on-macos-and-linux)).

### The web half

105 tests in 16 files under Vitest with jsdom: the ribbon's band, its tab and the report that it has
been drawn (`Band.test.tsx`, FR-614, FR-615); the pull out's handle (`PullOut.test.tsx`); the shell's
panels, refreshes, theme and colour scheme (`shell.test.tsx`); the menus' choices as Settings draws them (`MenuChoices.test.tsx`, FR-624); the opacity and its slider
(`opacity.test.ts`, `OpacitySlider.test.tsx`, FR-622); the corner grip (`ScaleGrip.test.tsx`,
FR-623); Help, About and Licence and their self-reading cycle (`Help.test.tsx`,
`autoScroll.test.ts`); a panel fitting its content (`panelFit.test.tsx`, FR-621); the background
colour reported to Go (`background.test.ts`); the `devicePixelRatio` watch (`pixelRatio.test.ts`); the scan
an application's suite holds its page's timers to, over a page of the same shape (`testing/timers.test.ts`); the setup page's screens, its keyboard ring and
the cases where the program cannot be reached (`web/setup`, over the page's own files loaded into
jsdom by `setupPage.ts`). No coverage provider is installed, so no figure is claimed. An application's own suite runs the page it
builds from the kit as a whole.

## How each part is tested

| Part | Kind of test | Touches |
|---|---|---|
| `domain` | pure unit | nothing |
| `application` | unit over hand-written fakes of the ports | nothing |
| `infrastructure` | integration over temporary folders and scratch registry keys | the filesystem, `HKCU` under a scratch key, child processes, the real displays |
| `ui/window` | unit over a scripted service, with Wails and the desktop stood in for | nothing |
| `structure` | unit over its recognisers and arithmetic | temporary files |
| `tests/structural` | source and AST scans over the kit | reads files |
| `web` | component tests under jsdom over the stand-in bridge (`web/testing`) | nothing |

No Go test uses a mocking library. **No test writes to the user's own settings, sign-in entry or
Apps list** and **no test reaches the network**: the update adapter runs over a stand-in HTTP client.

## What is not tested and why

### The platform owns it

- **`desktop` (49.7%).** The tray, native menus, move fence and broadcasts run on a hidden window's
  message loop and act on the real ribbon window. Tested: menu identifier numbering, a disabled item
  greyed in a menu Windows builds and reads back, the fence's
  arithmetic, a work area at a point, the drag distance, an address Windows refuses, the clock watch,
  the window's cut (FR-913) with the pointer read against it, the cursor read for the grip and every
  operation of the `shell.Desktop` port, on a hidden window that is never shown. The loop itself, the
  menus as drawn, the broadcasts and a browser opening are checked by hand in an application's real
  build.
- **`monitors` (82.6%):** Windows refusing to enumerate or describe a display.
- **`atomicfile` (85.7%).** Tested over a temporary folder: a file written whole then replaced whole,
  no folder, a replace refused with the temporary file removed, the longest name its temporary name
  allows written while the file system's own limit is refused. Not reached: the disk failing to
  take the bytes, flush them, close the file or set its mode.
- **`heldfile` (100%, Windows only).** Exists for tests: a held file can be neither read nor removed
  until it is let go; a missing file or an impossible path cannot be held.
- **`settingsfile` (97.3%).** Tested over a temporary folder through a small test product holding the
  ribbon's choices, a value of its own and a list: no file, a file kept aside (a second never
  replacing the first, none when every name is taken), a file that could not be read never saved
  over (held open on Windows), unknown keys and a byte order mark kept, every choice written and
  read, a bad value leaving its default, one bad entry leaving the rest, ids told apart, ordering by
  position, a value that cannot be written refused. Each application's own suite proves its keys.
  Not reached: a parsed object failing to tokenise and indenting failing on text built from valid
  JSON, which valid input cannot produce; the system failing to look up a kept-aside name for any
  reason but absence.
- **`occupancy` (90.1%).** Tested over a temporary folder with two products' places in it, each lock
  a real one: what one holds the other sees, a closed or crashed ribbon's entry passed over and
  removed, an entry that cannot be believed said once, a second copy holding nothing, no folder at
  all. Not reached: the system refusing to open or lock a lock file, list the folder or remove an
  entry. Whether a lock held in one Flatpak is seen from another was measured by hand on Linux.
- **`runlog` (77.8%):** the log refusing to open, its first line failing and `SetStdHandle` refusing.
- **`ui/window` (93.9%).** The window's decisions are tested over a scripted service: which calls
  refit the ribbon, panels and their fit, the tab, the window holding and cut to the pull out, the
  menu actions and the hand-over of those the kit does not know, closing, a signal ending the run, the
  recover round each event and update check, the update watch's timing, the grip's drag, the window's
  paint below full opacity, the first showing, Help and every call into the desktop going through the
  `shell.Desktop` port. Not reached: `run.go`, the one-line calls into Wails, `startup`, `listen` and
  `shutdown`, which only Wails runs.
- **`structure` (33.6%).** Its recognisers, the contrast arithmetic and the contrast check's findings
  are tested on their own. Its
  checks are called by `tests/structural`; every check was proved there by planting a violation
  and watching it fail; Go counts none of those calls towards this package's figure.

### It would change the machine

- **`installer` (68.3%).** The setup window is tested for the pictures it asks for and `Main` for
  refusing pictures it cannot open. The facade is tested over fakes of its `Machine`, `Processes`,
  `Log` and window: a machine that could not be read as the verdict, the route from one reading,
  the Uninstall screen, nothing touched while the application runs, the bar over each step, a failed
  step stopping the work, the boxes applied without the bar, closing the running copy, a program that
  will not start, the window left alone before startup. Not reached: `Run`, which opens the window;
  the three calls into Wails; the rest of `Main`, which reads the real machine before opening a
  window; a launch that starts. The policy beneath it is tested in `setup`; the page in `web/setup`.
- **`delivery` (96.5%).** Each command is tested over temporary folders: the version resource for an
  application and its setup program, a version refused, the shell names with Wails' bus name, every
  icon size, the payload's two entries, a refused packing leaving the archive that was there and every
  missing or unknown flag. Not reached: the disk refusing to take an encoded picture or to close the
  archive being packed.
- **`setup` (84.5%).** Tested over temporary folders, a scratch key and real stand-in processes. Not
  reached: the real Apps list record, deleting the install folder after setup exits, COM or a shortcut
  refusing, a copy failing part way, `TakeFocus`, finding the launched ribbon and `Places`.
- **`startup` (80.6%):** the registry refusing to open, read, write or delete.

## On macOS and Linux

Their halves of infrastructure compile only there; `cocoamain`, `gtkmain`, `monitors` and `desktop`
also need cgo against AppKit or GTK, so `test.ps1` reaches none of them. Check them on a machine of
that platform with the tools [DEVELOPMENT.md](DEVELOPMENT.md) names.

| What | macOS | Linux |
|---|---|---|
| Tags | `desktop,production` | `desktop,production,webkit2_41` |
| Go test functions | 350, plus 3 `TestMain` | 356, plus 3 `TestMain` |
| Tests that need cgo | `cocoamain` 3, `monitors` 3, `desktop` 19 | `gtkmain` 5, `monitors` 2, `desktop` 23, `platform` 1 |
| Needs | a signed-in desktop | a signed-in desktop with a display and a tray host |

With the platform's tags in `TAGS` and the kit's own folders in `KIT` (`./...` would also reach a Go
package an npm dependency ships under `node_modules`), run each and read its exit code:

```bash
export KIT="./application/... ./domain/... ./infrastructure/... ./installer/... ./structure/... ./tests/... ./ui/..."
```

```bash
test -z "$(gofmt -l application domain infrastructure installer structure tests ui)"
```

```bash
go vet -tags "$TAGS" $KIT
```

```bash
go run honnef.co/go/tools/cmd/staticcheck@v0.8.1 -tags "$TAGS" $KIT
```

```bash
go test -count=1 -tags "$TAGS" $KIT
```

```bash
npm install && npm run lint && npm run typecheck && npm test
```

What they prove: work handed to the toolkit's loop runs there and a panic reaches the caller;
displays are read with their work areas, AppKit's rectangles turned over to count from the top; the
ribbon is found by its title, kept off the taskbar, placed where asked and returned from a panel; a
move ends only when the ribbon was not placed there; the drag distance is the desktop's own; the tray
answers its host with a greyed disabled item. On macOS, a delegate that makes the application
regular as it launches, as Wails' does, cannot; the application agrees when macOS asks it to quit.
On Linux, GTK runs through X11 with the DMABUF renderer off unless chosen; a window mapped before
GTK's loop is veiled until it is shown and no other window is touched; only logind announcing that a
shutdown is starting ends the ribbon, a cancelled one not.

staticcheck is the version `test.ps1` pins. The desktop tests open real windows, place them and read
back where they stand; one registers a real tray or menu bar icon, so a person at the desktop sees
windows come and go. Each Linux `TestMain` fails at once, saying so, where no display opens. These
packages carry no coverage gate: a floor measured on one desktop would not hold on another.

## Running it

The whole gate:

```powershell
./test.ps1
```

It checks formatting, vet and staticcheck, runs every Go test, runs the web half's `lint`,
`typecheck` and `test`, holds the domain and application to 100% and every other gated package to its
floor. A missing `node_modules` stops the gate. Read the exit code: `0` means every check passed and
every floor held.

Another floor for the domain and application, for a deliberate check:

```powershell
./test.ps1 -Floor 95
```

The web half alone:

```powershell
npm test
```

One package's coverage in detail:

```powershell
go test -coverprofile=cover.out ./infrastructure/occupancy
```

```powershell
go tool cover -func=cover.out
```

## Keeping this honest

**Prove a new guard bites:** plant the violation, read the exit code, restore the file in a
`finally`. Confirm first that the clean tree passes, so a compile error cannot pass for a guard
biting. **Re-measure before quoting:** a figure copied forward describes a repository that no longer
exists. **Read the exit code, never the last line.**

See also [DEVELOPMENT.md](DEVELOPMENT.md), [ARCHITECTURE.md](ARCHITECTURE.md) and
[TECH_DEBT.md](TECH_DEBT.md).
