// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package runtime

import (
	"testing"

	. "github.com/apmckinlay/gsuneido/util/hamcrest"
)

func TestGlobals(t *testing.T) {
	foo := Global.Num("foo")
	Assert(t).That(Global.Num("foo"), Is(foo))
	Assert(t).That(Global.Add("bar", nil), Is(foo+1))
	Assert(t).That(Global.Num("bar"), Is(foo+1))
	Global.Add("baz", True)
	Assert(t).That(func() { Global.Add("baz", False) }, Panics("duplicate"))
	Assert(t).That(Global.Name(foo), Is("foo"))
	Assert(t).That(Global.Name(foo+1), Is("bar"))
}
