"""Generate a ribbon application's icons from the master artwork in its assets/ folder.

An application's own tools/genicons.py finds this file where Go builds the kit from and calls
generate with its repository and its button artwork. What comes out, every file from one of the
masters so the masters stay the one home:

  build/windows/icon.ico               the multi-size Windows icon both executables wear; the
                                       taskbar, the tray and the shortcuts read it out of the binary
  build/appicon.png                    the artwork Wails keeps beside it, which the tray on Linux
                                       and the menu bar on macOS are handed
  frontend/src/assets/app-icon.png     the application icon at the head of About
  frontend/src/assets/<button>         each button's artwork, at the height the page draws it
  installer/frontend/dist/icon.png     the setup window's header mark
  installer/frontend/dist/light-mode.png, dark-mode.png
                                       the setup window's theme toggle, named for the appearance
                                       each switches to
  docs/favicon.ico, favicon-32.png, apple-touch-icon.png
                                       the site's browser tab icon, the PNG a modern browser prefers
                                       and the icon iOS puts on a home screen; site_icons alone
                                       writes these three

The setup page has no build step, so it loads each file as it finds it; shipping the masters there
would put megabytes behind a badge. Each is written at about twice the size it is drawn at, crisp on
a high-density display without carrying detail nothing shows. It is not part of any build: the
output is committed, so a clone needs neither Python nor Pillow to build.
"""

from __future__ import annotations

import pathlib
import sys

try:
    from PIL import Image
except ImportError:  # pragma: no cover - a plain message beats a traceback
    sys.exit("Pillow is required: python -m pip install pillow")

# TOGGLE_MASTERS are the setup toggle's two faces, named for the appearance each switches to.
TOGGLE_MASTERS = ("light-mode.png", "dark-mode.png")

# ICO_SIZES are the sizes Windows chooses between: the tray and menu sizes, the taskbar and
# shortcut sizes, then the large one Explorer uses in its biggest view. Leaving one out makes
# Windows scale a neighbour, which looks soft.
ICO_SIZES = [(16, 16), (24, 24), (32, 32), (48, 48), (64, 64), (128, 128), (256, 256)]

# APPICON_SIZE is the square Wails expects its build/appicon.png to be.
APPICON_SIZE = 1024

# HEADER_SIZE is about twice the setup header mark's 126 pixels; TOGGLE_SIZE about twice the
# toggle's 44.
HEADER_SIZE = 256
TOGGLE_SIZE = 96

# Button artwork is drawn at a button's height, not as an icon: each is cropped to its artwork and
# scaled by height alone, rendered RENDER_SCALE times the height the page draws it at (.art img and
# .art.large img in the page's CSS) so it stays crisp under display scaling.
RENDER_SCALE = 4

# ABOUT_ICON_DRAWN is the application icon's size at the head of About (.about-head img in the
# kit's help.css); it is rendered RENDER_SCALE times that, square.
ABOUT_ICON_DRAWN = 128

# FAVICON_SIZES are the sizes a browser picks between for a tab, a bookmark and a pinned site; a
# reader that skips the page's link tags asks for /favicon.ico at the site root instead.
# FAVICON_PNG_SIZE is the PNG a modern browser prefers; TOUCH_ICON_SIZE is the square iOS asks for.
FAVICON_SIZES = [(16, 16), (32, 32), (48, 48)]
FAVICON_PNG_SIZE = 32
TOUCH_ICON_SIZE = 180


def squared(master: pathlib.Path) -> Image.Image:
    """Open a master, trim any transparent margin and centre it on a transparent square."""
    image = Image.open(master).convert("RGBA")
    box = image.getbbox()
    if box is not None:
        image = image.crop(box)
    side = max(image.width, image.height)
    square = Image.new("RGBA", (side, side), (0, 0, 0, 0))
    square.paste(image, ((side - image.width) // 2, (side - image.height) // 2), image)
    return square


def button_art(master: pathlib.Path, height: int) -> Image.Image:
    """Button artwork cropped to its pixels and scaled to height, keeping its shape."""
    image = Image.open(master).convert("RGBA")
    box = image.getbbox()
    if box is not None:
        image = image.crop(box)
    width = round(image.width * height / image.height)
    return image.resize((width, height), Image.LANCZOS)


def write_png(repo: pathlib.Path, image: Image.Image, size: int, target: pathlib.Path) -> None:
    """Write image at size square to target, reporting the file against repo."""
    target.parent.mkdir(parents=True, exist_ok=True)
    image.resize((size, size), Image.LANCZOS).save(target, "PNG", optimize=True)
    print(f"{target.relative_to(repo).as_posix():<42} {size:>5} px {target.stat().st_size:>9,} bytes")


def write_ico(repo: pathlib.Path, image: Image.Image, sizes: list, target: pathlib.Path) -> None:
    """Write image as a multi-size ICO at target, reporting the file against repo."""
    target.parent.mkdir(parents=True, exist_ok=True)
    image.save(target, "ICO", sizes=sizes)
    listed = ", ".join(str(width) for width, _ in sizes)
    print(f"{target.relative_to(repo).as_posix():<42} {listed} {target.stat().st_size:>9,} bytes")


def site_icons(repo: pathlib.Path) -> int:
    """Write the site's favicon set into repo/docs from the application master."""
    app_master = repo / "assets" / "application-icon.png"
    if not app_master.exists():
        sys.exit(f"no master artwork at {app_master}")
    app = squared(app_master)
    site = repo / "docs"
    write_ico(repo, app, FAVICON_SIZES, site / "favicon.ico")
    write_png(repo, app, FAVICON_PNG_SIZE, site / "favicon-32.png")
    write_png(repo, app, TOUCH_ICON_SIZE, site / "apple-touch-icon.png")
    return 0


def generate(repo: pathlib.Path, buttons: dict[str, int]) -> int:
    """Write every icon of the application at repo; buttons names each button master with the
    height the page draws it at."""
    masters = repo / "assets"
    app_master = masters / "application-icon.png"
    page_assets = repo / "frontend" / "src" / "assets"
    setup = repo / "installer" / "frontend" / "dist"
    names = (*TOGGLE_MASTERS, *buttons)
    for master in (app_master, *(masters / name for name in names)):
        if not master.exists():
            sys.exit(f"no master artwork at {master}")

    app = squared(app_master)
    write_ico(repo, app, ICO_SIZES, repo / "build" / "windows" / "icon.ico")
    write_png(repo, app, APPICON_SIZE, repo / "build" / "appicon.png")
    write_png(repo, app, HEADER_SIZE, setup / "icon.png")
    write_png(repo, app, RENDER_SCALE * ABOUT_ICON_DRAWN, page_assets / "app-icon.png")
    for name in TOGGLE_MASTERS:
        write_png(repo, squared(masters / name), TOGGLE_SIZE, setup / name)
    for name, drawn in buttons.items():
        art = button_art(masters / name, RENDER_SCALE * drawn)
        target = page_assets / name
        target.parent.mkdir(parents=True, exist_ok=True)
        art.save(target, "PNG", optimize=True)
        print(f"{target.relative_to(repo).as_posix():<42} {art.width}x{art.height} {target.stat().st_size:>9,} bytes")
    return site_icons(repo)
