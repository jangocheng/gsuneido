/*
Package verify provides a simple assertion facility

For example:

	verify.That(size >= 0)
*/
package verify

import (
	"runtime/debug"
)

// That panics if its argument is false
func That(cond bool) {
	if !cond {
		debug.PrintStack()
		panic("verify failed")
	}
}
