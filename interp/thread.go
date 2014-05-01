package interp

import . "github.com/apmckinlay/gsuneido/core/value"

type Thread struct {
	// frames are the Frame's making up the call stack.
	// The end of the slice is top of the stack (the current frame).
	frames []Frame
	// stack is the Value stack for arguments and expressions.
	// The end of the slice is the top of the stack.
	stack []Value
}

func (t *Thread) Push(x Value) {
	t.stack = append(t.stack, x)
}

func (t *Thread) Pop() Value {
	last := len(t.stack) - 1
	x := t.stack[last]
	t.stack = t.stack[:last]
	return x
}

// Call executes a Function and returns the result.
// The arguments must be already on the stack as per the ArgSpec.
// On return, the arguments are removed from the stack.
func (t *Thread) Call(fn Function, as ArgSpec) Value {
	defer func(sp int) { t.stack = t.stack[:sp] }(len(t.stack) - as.Nargs())
	t.args(fn, as)
	base := len(t.stack) - fn.nparams
	for i := fn.nparams; i < fn.nlocals; i++ {
		t.Push(nil)
	}
	frame := Frame{fn: fn, ip: 0, locals: t.stack[base:]}
	t.frames = append(t.frames, frame)
	defer func(fp int) { t.frames = t.frames[:fp] }(len(t.frames) - 1)
	return t.Interp()
}

// args converts the arguments on the stack as per the ArgSpec
// into the parameters expected by the function.
// On return, the stack is guaranteed to match the Function.
func (t *Thread) args(fn Function, as ArgSpec) {
	if fn.nparams == as.N_unnamed() {
		return // simple fast path
	}
	panic("not implemented") // TODO
}