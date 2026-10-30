package main

/*
#include <signal.h>

static void reset_sigusr1(void) { signal(SIGUSR1, SIG_DFL); }
*/
import "C"

// init hands SIGUSR1 back to the default disposition. The Go runtime installs
// a handler for it at startup, and WebKit, which takes SIGUSR1 for its garbage
// collector, prints "Overriding existing handler for signal 10" when it finds
// one there. Go itself doesn't use SIGUSR1. (JSC_SIGNAL_FOR_GC would move
// WebKit to another signal, but WebKitGTK 2.4x then logs "invalid option"
// from every process.)
func init() {
	C.reset_sigusr1()
}
