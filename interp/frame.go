package interp

import . "github.com/apmckinlay/gsuneido/core/value"

/*
Frame is the context for a function/method/block invocation.
*/
type Frame struct {
	// fn is the Function being executed
	fn Function
	// ip is the current index into the Function's code
	ip int
	// locals references a slice of the Thread stack
	// containing the parameters and local variables
	locals []Value
}

// Local returns a pointer to a local variable (including parameters)
// A pointer is returned so that the variable can be modified.
func (fr Frame) Local(i int) *Value {
	return &fr.locals[i]
}