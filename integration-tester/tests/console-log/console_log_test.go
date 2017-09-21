package main

import (
	"testing"
)

// This is run after integration-tester successfully runs console-log.js.
// XXX: Ideally, we would check that the expected output was actually
// printed to stdout, but the current integration-tester testing structure
// doesn't allow this file to see the output. The best we can do for now is
// ensure that `console.log` didn't break anything, which is implied if this
// file is invoked because integration-tester only runs checks after the
// containers specified in the spec are running.
func TestCalled(t *testing.T) {
}
