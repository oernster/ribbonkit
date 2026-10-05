// The Objective-C half of the macOS desktop: the ribbon's window, its moves, its popup menu and the
// icon in the menu bar. It lives apart from the Go files because a Go file that exports to C may
// only declare C functions. Everything here runs on AppKit's main thread.
//
// AppKit counts upward from the bottom-left corner of the menu-bar display; the kit counts
// downward from its top-left corner. Every position crossing to Go is turned over against that
// display's height, as the monitors package turns over the work areas.

#import <Cocoa/Cocoa.h>
#include "_cgo_export.h"

// The test window's size before a test places it; placing it replaces both.
#define TEST_WINDOW_WIDTH 200
#define TEST_WINDOW_HEIGHT 100

// How much of the menu bar's thickness the icon fills, leaving the margin other icons keep.
#define TRAY_ICON_SHARE 0.75

// primary_height answers the menu-bar display's height, which AppKit lists first.
static CGFloat primary_height(void)
{
    NSArray<NSScreen *> *screens = [NSScreen screens];
    return screens.count == 0 ? 0 : [screens[0] frame].size.height;
}

// ribbon_find answers the process's window titled title, else NULL. Wails gives the ribbon's
// window the product's name even though it has no title bar to show it.
void *ribbon_find(const char *title)
{
    NSString *wanted = [NSString stringWithUTF8String:title];
    for (NSWindow *window in [NSApp windows]) {
        if ([[window title] isEqualToString:wanted]) {
            return (__bridge void *)window;
        }
    }
    return NULL;
}

// ribbon_place stands the window with its top-left corner at x, y at width by height.
void ribbon_place(void *ribbon, int x, int y, int width, int height)
{
    NSRect frame = NSMakeRect(x, primary_height() - y - height, width, height);
    [(__bridge NSWindow *)ribbon setFrame:frame display:YES animate:NO];
}

// ribbon_frame answers where the window's top-left corner stands and its size.
void ribbon_frame(void *ribbon, int *x, int *y, int *width, int *height)
{
    NSRect frame = [(__bridge NSWindow *)ribbon frame];
    *x = (int)lround(frame.origin.x);
    *y = (int)lround(primary_height() - (frame.origin.y + frame.size.height));
    *width = (int)lround(frame.size.width);
    *height = (int)lround(frame.size.height);
}

// ribbon_pointer_inside answers whether the pointer is on the window now (FR-615, FR-616). The page
// hears nothing of the pointer while the ribbon is not the active application, so this is asked
// instead (measured 2026-09-28).
int ribbon_pointer_inside(void *ribbon)
{
    return NSMouseInRect([NSEvent mouseLocation], [(__bridge NSWindow *)ribbon frame], NO);
}

// ribbon_cursor answers where the pointer is, in points counted down from the menu-bar display's
// top-left corner, as ribbon_frame counts. The corner grip's drag reads it (FR-623).
void ribbon_cursor(int *x, int *y)
{
    NSPoint at = [NSEvent mouseLocation];
    *x = (int)lround(at.x);
    *y = (int)lround(primary_height() - at.y);
}

// test_warp moves the pointer to x, y in Quartz's global points, which count down from the
// menu-bar display's top-left corner as the kit does.
void test_warp(int x, int y)
{
    CGWarpMouseCursorPosition(CGPointMake(x, y));
}

// The observers desktop_watch registers, held so they live as long as the process.
static NSMutableArray *observers;

// desktop_watch reports each move of the ribbon and each change of the displays to the desktop
// handle holds.
void desktop_watch(void *ribbon, uintptr_t handle)
{
    if (observers == nil) {
        observers = [NSMutableArray array];
    }
    NSNotificationCenter *centre = [NSNotificationCenter defaultCenter];
    [observers addObject:[centre addObserverForName:NSWindowDidMoveNotification
                                             object:(__bridge NSWindow *)ribbon
                                              queue:nil
                                         usingBlock:^(NSNotification *note) {
        int x = 0, y = 0, width = 0, height = 0;
        ribbon_frame(ribbon, &x, &y, &width, &height);
        desktopMoved(handle, x, y);
    }]];
    [observers addObject:[centre addObserverForName:NSApplicationDidChangeScreenParametersNotification
                                             object:nil
                                              queue:nil
                                         usingBlock:^(NSNotification *note) {
        desktopDisplaysChanged(handle);
    }]];
}

// RibbonKitMenuTarget receives every menu item's choice and reports its number.
@interface RibbonKitMenuTarget : NSObject
- (void)chosen:(NSMenuItem *)item;
@end

@implementation RibbonKitMenuTarget
- (void)chosen:(NSMenuItem *)item
{
    desktopMenuChosen((int)[item tag]);
}
@end

static RibbonKitMenuTarget *menuTarget;

// menu_new answers an empty menu, owned by the caller until desktop_popup takes it.
void *menu_new(void)
{
    NSMenu *menu = [[NSMenu alloc] init];
    [menu setAutoenablesItems:NO];
    return (__bridge_retained void *)menu;
}

// menu_add_item adds an item numbered index, checked or not where it can be, greyed where not enabled.
void menu_add_item(void *menu, const char *label, int checkable, int checked, int enabled, int index)
{
    if (menuTarget == nil) {
        menuTarget = [RibbonKitMenuTarget new];
    }
    NSMenuItem *item = [[NSMenuItem alloc] initWithTitle:[NSString stringWithUTF8String:label]
                                                  action:@selector(chosen:)
                                           keyEquivalent:@""];
    item.target = menuTarget;
    item.tag = index;
    item.enabled = enabled ? YES : NO;
    if (checkable) {
        item.state = checked ? NSControlStateValueOn : NSControlStateValueOff;
    }
    [(__bridge NSMenu *)menu addItem:item];
}

// menu_add_submenu adds an item opening a submenu, answering the submenu; the item owns it.
void *menu_add_submenu(void *menu, const char *label)
{
    NSString *title = [NSString stringWithUTF8String:label];
    NSMenuItem *item = [[NSMenuItem alloc] initWithTitle:title action:nil keyEquivalent:@""];
    NSMenu *sub = [[NSMenu alloc] initWithTitle:title];
    [sub setAutoenablesItems:NO];
    item.submenu = sub;
    [(__bridge NSMenu *)menu addItem:item];
    return (__bridge void *)sub;
}

void menu_add_separator(void *menu)
{
    [(__bridge NSMenu *)menu addItem:[NSMenuItem separatorItem]];
}

// desktop_popup shows menu at the pointer, taking it over from the caller. The menu runs its own
// loop until it closes, so it is started once the current work returns rather than inside it, which
// would hold the caller waiting on the main thread for as long as the menu stays open.
void desktop_popup(void *menu)
{
    NSMenu *owned = (__bridge_transfer NSMenu *)menu;
    dispatch_async(dispatch_get_main_queue(), ^{
        [owned popUpMenuPositioningItem:nil atLocation:[NSEvent mouseLocation] inView:nil];
        // The menu has closed, its choice already sent (FR-616).
        desktopMenuClosed();
    });
}

// RibbonKitTray rebuilds the menu-bar icon's menu each time it is about to open, so it always
// says what is true now.
@interface RibbonKitTray : NSObject <NSMenuDelegate>
@property (nonatomic) uintptr_t handle;
@end

@implementation RibbonKitTray
- (void)menuNeedsUpdate:(NSMenu *)menu
{
    [menu removeAllItems];
    trayMenuNeedsUpdate(self.handle, (__bridge void *)menu);
}
@end

static NSStatusItem *statusItem;
static RibbonKitTray *trayDelegate;

// tray_start puts the icon, a PNG of length bytes, in the menu bar with tooltip. The bytes are
// copied at once; the icon is made once AppKit's loop runs, since the application may not have
// finished launching yet.
void tray_start(uintptr_t handle, const void *png, int length, const char *tooltip)
{
    NSData *data = [NSData dataWithBytes:png length:length];
    NSString *tip = [NSString stringWithUTF8String:tooltip];
    dispatch_async(dispatch_get_main_queue(), ^{
        NSStatusBar *bar = [NSStatusBar systemStatusBar];
        CGFloat side = [bar thickness] * TRAY_ICON_SHARE;
        NSImage *image = [[NSImage alloc] initWithData:data];
        [image setSize:NSMakeSize(side, side)];
        statusItem = [bar statusItemWithLength:NSSquareStatusItemLength];
        statusItem.button.image = image;
        statusItem.button.toolTip = tip;
        NSMenu *menu = [[NSMenu alloc] init];
        [menu setAutoenablesItems:NO];
        trayDelegate = [RibbonKitTray new];
        trayDelegate.handle = handle;
        menu.delegate = trayDelegate;
        statusItem.menu = menu;
    });
}

void tray_stop(void)
{
    if (statusItem != nil) {
        [[NSStatusBar systemStatusBar] removeStatusItem:statusItem];
    }
    statusItem = nil;
    trayDelegate = nil;
}

int tray_shown(void)
{
    return statusItem != nil && statusItem.button.image != nil;
}

// test_window makes a window set up as Wails sets up the ribbon's, titled and without a title bar,
// then shows it. The caller owns it until test_window_close.
void *test_window(const char *title)
{
    NSWindow *window = [[NSWindow alloc]
        initWithContentRect:NSMakeRect(0, 0, TEST_WINDOW_WIDTH, TEST_WINDOW_HEIGHT)
                  styleMask:NSWindowStyleMaskBorderless | NSWindowStyleMaskMiniaturizable
                    backing:NSBackingStoreBuffered
                      defer:NO];
    window.title = [NSString stringWithUTF8String:title];
    window.releasedWhenClosed = NO;
    [window orderFrontRegardless];
    return (__bridge_retained void *)window;
}

// test_window_close closes a test window. AppKit may keep a closed window in its list for a while,
// so its title is cleared first: a later test looking for the ribbon by title must not find it.
void test_window_close(void *ribbon)
{
    NSWindow *window = (__bridge_transfer NSWindow *)ribbon;
    window.title = @"";
    [window close];
}
