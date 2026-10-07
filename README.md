# ribbonkit

> **Commercial licences available.** ribbonkit is free and open source under the GNU General Public License, version 3 (GPL-3.0). If those terms do not suit what you are building, such as a closed-source product, a commercial licence can be bought from me separately. It covers my own code; third-party libraries keep their own licences. See [commercial licensing](https://ernster.dev/commercial-licensing.html).

The desktop half of a ribbon application, written once: a narrow always-available window docked to
an edge of the screen, with the content of the application drawn inside it. TimeRibbon (a ribbon of
clocks) and WeatherRibbon are built on it, so a fix to how a ribbon is placed, dragged, scaled or
installed lands once and reaches both.

It is a Go module and an npm package in one repository, released together under one tag.

## Who it is for

- An application built with Wails v2 and React that shows its content in a ribbon on Windows, macOS
  and Linux.
- Its author, who wants the ribbon's desktop behaviour without writing it again.

## Who it is not for

- An application that is not a ribbon. The kit decides the window's shape, its place against an edge
  and how it hides; it is not a general desktop toolkit.
- An application without Wails v2 and React: the window is a Wails facade and the page half is React.
- Anyone after a stable API. Until the first major release a minor release may change what an
  application calls.

## What it provides

- **Placement:** the ribbon against an edge of a display's work area, restored at the display's DPI,
  kept on a display while dragged, snapped to an edge, centred along it, the tab it collapses to and
  the pull out beside it.
- **Two ribbons on one desktop never land on each other:** every running ribbon holds an entry in a
  shared folder and the one being placed yields.
- **The window:** a Wails facade an application embeds, with the drag, the corner grip that scales the
  ribbon, opacity, panels in the same window, the tray and the native menus, start at sign-in, the
  update check and Help, About and Licence.
- **Standing well with each desktop:** no taskbar button on Windows or Linux and no Dock icon on
  macOS; on macOS it quits when the system asks at log out or restart; on Linux no window shows
  before it is placed and it leaves as soon as a restart is announced.
- **What a ribbon keeps:** its settings file, read and written alike in every ribbon, the folder it
  lives in, files replaced whole and a log of each run.
- **Time zones** resolved with the rules built in, for a ribbon that shows places.
- **The page half:** the bridge to the window, the band the content is drawn in, the pull out, the
  shell that routes panels and redraws, the grip and opacity controls, the ribbon's palette and a
  stand-in bridge for tests.
- **The setup program for Windows:** the install policy (install, update, go back, repair, reinstall,
  uninstall, all per user) and the window over it, named by the application.
- **The build tools' work:** the Windows version resource, the names the macOS and Linux scripts
  read, the Linux icon theme's sizes, the setup program's payload and the icon generator.
- **The structural tests' mechanics:** the `structure` package, which holds an application's own
  layers, size limit and network rules with the same code that holds the kit's.

The kit names no product. An application hands it an `identity.App` (its name and reverse-domain id),
its GitHub repository for the update check and a `setup.Product` for setup.

## Stack

| Part | What |
|---|---|
| Go | at the versions `go.mod` declares: Wails v2; `golang.org/x/sys`, `go-ole` on Windows; `godbus` on Linux |
| Native | Win32 on Windows; AppKit through cgo on macOS; GTK 3 through cgo on Linux |
| Page | React 18 and TypeScript, shipped as source for the application's own build |
| Tests | Go's `testing`; Vitest with jsdom and Testing Library |

## Taking it up

In the application's module, with `<tag>` the release to take up:

```powershell
go get github.com/oernster/ribbonkit@<tag>
```

In its front end's `package.json`, the same tag:

```json
"@oernster/ribbonkit": "github:oernster/ribbonkit#<tag>"
```

The page imports the kit from `@oernster/ribbonkit` and its stand-in bridge from
`@oernster/ribbonkit/testing`. The front end may extend the kit's `tsconfig.json`, spread its eslint
rules (`@oernster/ribbonkit/eslint`) and name its test set-up (`@oernster/ribbonkit/test-setup`), so
both are held to the same settings.

## Testing

```powershell
./test.ps1
```

Read the exit code: `0` means every check passed and every floor held. [TESTING.md](TESTING.md) says
what that covers and what is checked on macOS and Linux.

## More

- [ARCHITECTURE.md](ARCHITECTURE.md): the layers, what each package does and the rules the tests hold.
- [TESTING.md](TESTING.md): the gate, the floors and their reasons.
- [DEVELOPMENT.md](DEVELOPMENT.md): working on the kit beside an application and cutting a release.
- [DECISIONS-TRADEOFFS.md](DECISIONS-TRADEOFFS.md): the decisions the kit rests on, with what each
  costs.
- [TECH_DEBT.md](TECH_DEBT.md): what is still open, what is deliberately left and what only looks
  like debt.

## Licence

GNU General Public License, version 3: see [LICENSE](LICENSE).

Commercial licences are available: see [commercial licensing](https://ernster.dev/commercial-licensing.html).
