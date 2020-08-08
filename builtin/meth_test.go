// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package builtin

import (
	"testing"

	. "github.com/apmckinlay/gsuneido/runtime"
	. "github.com/apmckinlay/gsuneido/util/hamcrest"
)

func Test_MethodCall(t *testing.T) {
	n := NumFromString("12.34")
	th := NewThread()
	f := n.Lookup(th, "Round")
	th.Push(IntVal(1))
	result := f.Call(th, n, &ArgSpec1)
	Assert(t).That(result, Is(NumFromString("12.3")))
}
