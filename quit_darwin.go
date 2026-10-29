//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

// Stops the app's run loop from the main thread. -[NSApp stop:] only takes effect once the
// loop handles another event, so post an empty one to wake it.
static void quitApp(void) {
	dispatch_async(dispatch_get_main_queue(), ^{
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
	});
}
*/
import "C"

import "context"

// beforeClose quits the whole app when the window closes or on Cmd+Q. Wails stops the run
// loop from a background goroutine, which Cocoa doesn't notice until the next event, so the
// app stayed running with no window and OnShutdown never ran (#176). Returning true skips
// Wails' own quit; once the loop ends, Wails calls OnShutdown and the process exits.
func beforeClose(context.Context) bool {
	C.quitApp()
	return true
}
