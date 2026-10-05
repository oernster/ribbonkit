// The Objective-C half of cocoamain. It lives apart from the Go file because a Go file that exports
// a function to C may only declare C functions, never define them.

#import <Cocoa/Cocoa.h>
#include "_cgo_export.h"

// cocoamain_invoke runs the handle's function on the main thread: at once when this is the main
// thread, otherwise queued there and waited for.
void cocoamain_invoke(uintptr_t handle)
{
    if ([NSThread isMainThread]) {
        cocoamainRun(handle);
        return;
    }
    dispatch_sync(dispatch_get_main_queue(), ^{
        cocoamainRun(handle);
    });
}

// cocoamain_serve runs the application's loop until cocoamain_stop. The activation policy is left
// as a bare program has it, with no Dock icon (measured 2026-09-28: Prohibited), so a test that
// hides the ribbon from the Dock proves the hiding rather than finding it done already.
void cocoamain_serve(void)
{
    [NSApplication sharedApplication];
    [NSApp run];
}

// cocoamain_stop ends the loop. The loop only notices a stop after its next event, so one is
// posted to wake it.
void cocoamain_stop(void)
{
    [NSApp stop:nil];
    NSEvent *wake = [NSEvent otherEventWithType:NSEventTypeApplicationDefined
                                       location:NSZeroPoint
                                  modifierFlags:0
                                      timestamp:0
                                   windowNumber:0
                                        context:nil
                                        subtype:0
                                          data1:0
                                          data2:0];
    [NSApp postEvent:wake atStart:YES];
}
