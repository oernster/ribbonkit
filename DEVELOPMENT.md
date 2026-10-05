# Development

How to work on ribbonkit and release it. The kit builds nothing of its own: an application builds it
into its executables and its page. Commands run from the repository root: PowerShell on Windows, the
Terminal's shell elsewhere. Testing is in [TESTING.md](TESTING.md).

## What a machine needs

| Tool | Version | What for | Where from |
|---|---|---|---|
| Go | 1.26.3, as `go.mod` declares | the Go module and its tests | [go.dev/dl](https://go.dev/dl/) |
| Node.js with npm | a current LTS | the web half's lint, type check and tests | [nodejs.org](https://nodejs.org/) or `winget install OpenJS.NodeJS.LTS` |
| Xcode (macOS) | current | cgo's compiler for the AppKit half | the App Store |
| GTK 3 and WebKitGTK 4.1 headers (Linux) | as the distribution ships | cgo's compiler for the GTK half | `sudo apt-get install -y build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev` |

The gate fetches staticcheck through `go run` on its first run, so it needs the network once. Windows
needs no C compiler.

## Getting the source

```powershell
git clone https://github.com/oernster/ribbonkit.git
```

```powershell
cd ribbonkit
```

```powershell
go mod download
```

```powershell
npm install
```

## Working on the kit beside an application

An application depends on one tag of the kit, in its `go.mod` and its front end's `package.json`
alike. To change the kit and see the change in the application before tagging, clone the kit beside
the application and give the application a `go.work` (gitignored there), from the application's
root:

```powershell
go work init . ../ribbonkit
```

Go then builds the application against the kit's working copy. The application's release build sets
`GOWORK=off`, so what it ships is always the tagged kit. Delete the `go.work` once the kit is tagged
and the application names the new tag.

## Cutting a release

The tag is the version: Go reads it from git and the application's `package.json` names it. A tag
before `v1.0.0` may change what an application calls.

1. Run `./test.ps1` on Windows and read its exit code.
2. Run the checks in [TESTING.md](TESTING.md#on-macos-and-linux) on a Mac and on Linux.
3. Tag the commit `v` plus the version and push the tag.
4. In each application, name the new tag in `go.mod` (`go get github.com/oernster/ribbonkit@<tag>`)
   and in the front end's `package.json`, then run its own gate and its checks by hand in a real
   build. The window, the tray, focus, paint and the install are seen only there.

## Where things live

| Path | What it holds |
|---|---|
| `domain`, `application` | the pure rules (placement, the ribbon's choices, hovering, the application's names) and the use cases over their ports (arranging the ribbon, the menus, the update check, the desktop's port) |
| `infrastructure` | the adapters: the desktop, displays, sign-in start, the log, the data folder, the folder every running ribbon shares, a file replaced whole, the update check, the install policy and a file held open for tests; a file's platform is in its name (`_windows`, `_linux`, `_darwin`, `_unix`) |
| `ui/window` | the ribbon's window: its life and the desktop's events, the tab, the pull out, panels, the grip, opacity, menu choices, the update check, Help, the first showing and the Wails options (`run.go`) |
| `installer` | the setup program's window: its page (`page/`, no build step), the facade the page calls and `Run` |
| `web`, `package.json` | the page's half, an npm package shipped as source; `web/testing` is the stand-in bridge, Go's events and the one test set-up |
| `structure` | the structural tests' mechanics, shared with every application |
| `tests/structural` | the kit's own structural tests |
| `test.ps1` | the gate |
| `tsconfig.json`, `eslint.config.js`, `vite.config.ts` | the web half's settings, which an application's front end may extend or spread |

## House rules worth knowing before a first change

- **The layer direction is enforced:** the domain imports nothing outside itself and reads no clock;
  the application imports neither infrastructure nor Wails; the UI imports no infrastructure (its
  tests included); nothing below the UI imports it; nothing in the kit wires the application to the
  infrastructure.
- **The kit names no product and no repository.** Everything that writes a name or asks a network
  takes it from the application.
- **Every exported method of `window.Window` is page API,** since an application embeds it and Wails
  binds promoted methods too. What only the application may call goes on `window.Control`.
- **No file over 400 lines,** tests included; one between 381 and 400 goes down to 350 or fewer.
- **No magic numbers:** a literal needing a comment is a named constant or derived from data.
- **The wire is written twice,** in `ui/window/wire.go` and `web/wire.ts`; change both sides.
- **The update check is the one network request** (`network_test.go`).
- **A page call Go can refuse takes a refusal handler** and answers null rather than rejecting.
- **Every new guard is proved by planting a violation** ([TESTING.md](TESTING.md#keeping-this-honest)).

See also [README.md](README.md), [TESTING.md](TESTING.md), [ARCHITECTURE.md](ARCHITECTURE.md) and
[TECH_DEBT.md](TECH_DEBT.md).
