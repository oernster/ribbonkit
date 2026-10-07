// How the application stands with macOS: kept off the Dock with its menu-bar icon staying (FR-101),
// never recorded by the Dock as a recent app, and quitting when macOS asks. These belong to the
// application rather than the window, so they live apart from the window's half in desktop_darwin.m.
// Everything here runs on AppKit's main thread, or before it starts.

#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>

// The type encoding of applicationShouldTerminate:, used where the delegate has no such method to
// copy one from: an NSUInteger answer, the receiver, the selector and the NSApplication.
#define SHOULD_TERMINATE_TYPES "Q@:@"

// The class of Wails' application delegate, as Wails names it.
#define WAILS_DELEGATE_CLASS "AppDelegate"

// ribbon_leave_dock makes the application an accessory: no Dock icon and no place in the
// application switcher, while its menu-bar icon stays (FR-101). Wails makes the application
// regular in applicationWillFinishLaunching, which overrides LSUIElement and any earlier switch
// (measured 2026-10-06), so where launching has not finished the switch is made again once it has.
void ribbon_leave_dock(void)
{
    [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
    if ([[NSRunningApplication currentApplication] isFinishedLaunching]) {
        return;
    }
    __block id observer = [[NSNotificationCenter defaultCenter]
        addObserverForName:NSApplicationDidFinishLaunchingNotification
                    object:nil
                     queue:[NSOperationQueue mainQueue]
                usingBlock:^(NSNotification *note) {
        [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
        [[NSNotificationCenter defaultCenter] removeObserver:observer];
        observer = nil;
    }];
}

// ribbon_skips_dock answers whether the application is an accessory, kept off the Dock.
int ribbon_skips_dock(void)
{
    return [NSApp activationPolicy] == NSApplicationActivationPolicyAccessory;
}

// The implementations replaced while the delegate finishes launching: the delegate's own handler,
// and AppKit's setActivationPolicy: for the moment it is set aside.
static IMP delegate_will_finish;
static IMP appkit_set_policy;

// ribbon_refuse_regular stands in for setActivationPolicy: while the delegate finishes launching: a
// switch to a regular application is refused, any other goes through.
static BOOL ribbon_refuse_regular(id self, SEL command, NSApplicationActivationPolicy policy)
{
    if (policy == NSApplicationActivationPolicyRegular) {
        return NO;
    }
    return ((BOOL (*)(id, SEL, NSApplicationActivationPolicy))appkit_set_policy)(self, command, policy);
}

// ribbon_will_finish runs the delegate's applicationWillFinishLaunching: whole, except that a switch
// to a regular application is refused. Wails makes that switch, which made the Dock record the
// ribbon as a recent app at every launch, though the ribbon became an accessory a moment later; with
// it refused the Dock made no tile at all (measured 2026-10-07 from the Mac's log, two log ins).
static void ribbon_will_finish(id self, SEL command, NSNotification *note)
{
    Method policy = class_getInstanceMethod([NSApplication class], @selector(setActivationPolicy:));
    appkit_set_policy = method_setImplementation(policy, (IMP)ribbon_refuse_regular);
    ((void (*)(id, SEL, NSNotification *))delegate_will_finish)(self, command, note);
    method_setImplementation(policy, appkit_set_policy);
}

// ribbon_wrap_will_finish wraps delegate's applicationWillFinishLaunching: in ribbon_will_finish.
// One delegate is wrapped per process; a class without the method is left alone.
void ribbon_wrap_will_finish(Class delegate)
{
    if (delegate == Nil || delegate_will_finish != NULL) {
        return;
    }
    Method finish = class_getInstanceMethod(delegate, @selector(applicationWillFinishLaunching:));
    if (finish == NULL) {
        return;
    }
    delegate_will_finish = method_setImplementation(finish, (IMP)ribbon_will_finish);
}

// ribbon_keep_agent wraps Wails' delegate at process start, before AppKit can call it; the ribbon's
// window cannot be found until after, which is too late. Where Wails is not linked, as in the kit's
// own tests, there is no such class and nothing is wrapped.
__attribute__((constructor)) static void ribbon_keep_agent(void)
{
    ribbon_wrap_will_finish(objc_getClass(WAILS_DELEGATE_CLASS));
}

// ribbon_terminate_now answers macOS's request to quit with yes.
static NSApplicationTerminateReply ribbon_terminate_now(id self, SEL command, NSApplication *sender)
{
    return NSTerminateNow;
}

// ribbon_honour_quit makes the application quit when macOS asks it to: at log out, restart and shut
// down, and from Activity Monitor or a script. Wails answers every such request with
// NSTerminateCancel and hands it to the window's close handler, which hides the ribbon rather than
// quitting, so the request was refused and a restart interrupted (measured 2026-10-06: a quit Apple
// event was answered "User cancelled"). Closing the window does not come here, so it still hides.
// Without a delegate AppKit already quits, so there is nothing to answer.
void ribbon_honour_quit(void)
{
    id delegate = [NSApp delegate];
    if (delegate == nil) {
        return;
    }
    SEL selector = @selector(applicationShouldTerminate:);
    Method existing = class_getInstanceMethod([delegate class], selector);
    const char *types = existing != NULL ? method_getTypeEncoding(existing) : SHOULD_TERMINATE_TYPES;
    class_replaceMethod([delegate class], selector, (IMP)ribbon_terminate_now, types);
}

// RibbonRefusingDelegate refuses to quit as Wails' delegate does, for the tests.
@interface RibbonRefusingDelegate : NSObject <NSApplicationDelegate>
@end

@implementation RibbonRefusingDelegate
- (NSApplicationTerminateReply)applicationShouldTerminate:(NSApplication *)sender
{
    return NSTerminateCancel;
}
@end

// The test delegate, kept alive while it is the application's delegate, which holds it weakly.
static RibbonRefusingDelegate *testDelegate = nil;

// test_refusing_delegate makes a delegate that refuses to quit the application's delegate.
void test_refusing_delegate(void)
{
    testDelegate = [[RibbonRefusingDelegate alloc] init];
    [NSApp setDelegate:testDelegate];
}

// test_quits answers whether the application's delegate agrees to quit.
int test_quits(void)
{
    return [[NSApp delegate] applicationShouldTerminate:NSApp] == NSTerminateNow;
}

// test_no_delegate takes the test delegate away again.
void test_no_delegate(void)
{
    [NSApp setDelegate:nil];
    testDelegate = nil;
}

// RibbonRegularDelegate makes the application regular as it finishes launching, as Wails' delegate
// does, for the tests.
@interface RibbonRegularDelegate : NSObject <NSApplicationDelegate>
@end

@implementation RibbonRegularDelegate
- (void)applicationWillFinishLaunching:(NSNotification *)note
{
    [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];
}
@end

// test_finish_launching makes the application an accessory, has a RibbonRegularDelegate finish
// launching it, then answers whether the application is still an accessory. The application is left
// an accessory either way.
int test_finish_launching(void)
{
    [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
    NSNotification *note =
        [NSNotification notificationWithName:NSApplicationWillFinishLaunchingNotification object:NSApp];
    [[[RibbonRegularDelegate alloc] init] applicationWillFinishLaunching:note];
    int stayed = ribbon_skips_dock();
    [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
    return stayed;
}

// test_wrap_regular_delegate wraps RibbonRegularDelegate as the kit wraps Wails' delegate.
void test_wrap_regular_delegate(void)
{
    ribbon_wrap_will_finish([RibbonRegularDelegate class]);
}

// test_policy_settable answers whether setActivationPolicy: works again outside the wrapped
// handler: it switches to regular and back, answering whether the first switch was made.
int test_policy_settable(void)
{
    [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];
    int made = [NSApp activationPolicy] == NSApplicationActivationPolicyRegular;
    [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
    return made;
}
