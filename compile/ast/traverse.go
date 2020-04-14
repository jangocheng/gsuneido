// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package ast

type Visitor interface {
	Before(node Node) bool // return false to skip children
	After(node Node) Node // return non-nil to update
}

// Traverse calls visitor.Before for node.
// If Before returns true,
// Traverse is called recursively for each child node,
// and then visitor.After is called for node.
// NOTE: it will not traverse nested functions and classes
// because they will be constants.
func Traverse(node Node, visitor Visitor) Node {
	if node == nil || !visitor.Before(node) {
		return node
	}
	node.Children(func(child Node) Node { return Traverse(child, visitor) })
	return visitor.After(node)
}
