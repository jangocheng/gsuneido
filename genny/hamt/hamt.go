// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package hamt

import (
	"math/bits"

	"github.com/cheekybits/genny/generic"
)

type Item generic.Type
type KeyType generic.Type

type ItemHamt struct {
	root       *node
	mutable    bool
	generation uint32 // if mutable, nodes with this generation are mutable
}

type node struct {
	generation uint32
	bmVal      uint32
	bmPtr      uint32
	vals       []Item
	ptrs       []*node
}

const bitsPerNode = 5
const mask = 1<<bitsPerNode - 1

func (ht ItemHamt) Get(key KeyType) (Item, bool) {
	var zero Item
	nd := ht.root
	if nd == nil {
		return zero, false
	}
	hash := ItemHash(key)
	for shift := 0; shift < 32; shift += bitsPerNode { // iterative
		bit := nd.bit(hash, shift)
		iv := bits.OnesCount32(nd.bmVal & (bit - 1))
		if (nd.bmVal & bit) != 0 {
			if nd.vals[iv].Key() != key {
				return zero, false
			}
			return nd.vals[iv], true
		}
		if (nd.bmPtr & bit) == 0 {
			return zero, false
		}
		ip := bits.OnesCount32(nd.bmPtr & (bit - 1))
		nd = nd.ptrs[ip]
	}
	// overflow node, linear search
	for i := range nd.vals {
		if nd.vals[i].Key() == key {
			return nd.vals[i], true
		}
	}
	return zero, false // not found
}

func (*node) bit(hash uint32, shift int) uint32 {
	return 1 << ((hash >> shift) & mask)
}

//-------------------------------------------------------------------

func (ht ItemHamt) Mutable() ItemHamt {
	gen := ht.generation + 1
	nd := ht.root
	if nd == nil {
		nd = &node{generation: gen}
	}
	nd = nd.dup()
	nd.generation = gen
	return ItemHamt{root: nd, mutable: true, generation: gen}
}

func (ht ItemHamt) Put(item *Item) {
	if !ht.mutable {
		panic("can't modify an immutable Hamt")
	}
	key := item.Key()
	hash := ItemHash(key)
	ht.root.with(ht.generation, item, key, hash, 0)
}

func (nd *node) with(gen uint32, item *Item, key KeyType, hash uint32, shift int) *node {
	// recursive
	if nd.generation != gen {
		// path copy on the way down the tree
		nd = nd.dup()
		nd.generation = gen // now mutable in this generation
	}
	if shift >= 32 {
		// overflow node
		for i := range nd.vals { // linear search
			if nd.vals[i].Key() == key {
				nd.vals[i] = *item // update if found
				return nd
			}
		}
		nd.vals = append(nd.vals, *item) // not found, add it
		return nd
	}
	bit := nd.bit(hash, shift)
	ip := bits.OnesCount32(nd.bmPtr & (bit - 1))
	if (nd.bmPtr & bit) != 0 {
		// recurse to child node
		nd.ptrs[ip] = nd.ptrs[ip].with(gen, item, key, hash, shift+bitsPerNode)
		return nd
	}
	iv := bits.OnesCount32(nd.bmVal & (bit - 1))
	if (nd.bmVal & bit) == 0 {
		// slot is empty, insert new value
		nd.bmVal |= bit
		nd.vals = append(nd.vals, Item{})
		copy(nd.vals[iv+1:], nd.vals[iv:])
		nd.vals[iv] = *item
		return nd
	}
	if nd.vals[iv].Key() == key {
		// already exists, update it
		nd.vals[iv] = *item
		return nd
	}
	// collision, create new child node
	nu := &node{generation: gen}
	if shift+bitsPerNode < 32 {
		oldval := &nd.vals[iv]
		oldkey := oldval.Key()
		nu = nu.with(gen, oldval, oldkey, ItemHash(oldkey), shift+bitsPerNode)
		nu = nu.with(gen, item, key, hash, shift+bitsPerNode)
	} else {
		// overflow node, no bitmaps, just list values
		nu.vals = append(nu.vals, nd.vals[iv], *item)
	}

	// remove old colliding value from node
	nd.bmVal &^= bit
	copy(nd.vals[iv:], nd.vals[iv+1:])
	nd.vals = nd.vals[:len(nd.vals)-1]

	// point to new child node instead
	nd.ptrs = append(nd.ptrs, nil)
	copy(nd.ptrs[ip+1:], nd.ptrs[ip:])
	nd.ptrs[ip] = nu
	nd.bmPtr |= bit

	return nd
}

func (nd *node) dup() *node {
	dup := *nd // shallow copy
	dup.vals = append(nd.vals[0:0:0], nd.vals...)
	dup.ptrs = append(nd.ptrs[0:0:0], nd.ptrs...)
	return &dup
}

func (ht ItemHamt) Freeze() ItemHamt {
	return ItemHamt{root: ht.root, generation: ht.generation}
}

//-------------------------------------------------------------------

func (ht ItemHamt) ForEach(fn func(*Item)) {
	if ht.root != nil {
		ht.root.forEach(fn)
	}
}

func (nd *node) forEach(fn func(*Item)) {
	for i := range nd.vals {
		fn(&nd.vals[i])
	}
	for _, p := range nd.ptrs {
		p.forEach(fn)
	}
}
