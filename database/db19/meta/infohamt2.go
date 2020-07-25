// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package meta

import (
	"sort"

	"github.com/apmckinlay/gsuneido/database/db19/stor"
	"github.com/apmckinlay/gsuneido/util/verify"
)

// list returns a list of the keys in the table
func (ht InfoHamt) list() []string {
	keys := make([]string, 0, 16)
	ht.ForEach(func(it *Info) {
		keys = append(keys, InfoKey(it))
	})
	return keys
}

const blockSizeInfo = 2000
const perFingerInfo = 16

func (ht InfoHamt) Write(st *stor.Stor) uint64 {
	nitems := 0
	size := 2
	ht.ForEach(func(it *Info) {
		size += it.storSize()
		nitems++
	})
	if nitems == 0 {
		off, buf := st.Alloc(2)
		stor.NewWriter(buf).Put2(0)
		return off
	}
	nfingers := 1 + nitems/perFingerInfo
	size += 3 * nfingers
	off, buf := st.Alloc(size)
	w := stor.NewWriter(buf)
	w.Put2(nitems)

	keys := ht.list()
	sort.Strings(keys)
	w2 := *w
	for i := 0; i < nfingers; i++ {
		w.Put3(0) // leave room
	}
	fingers := make([]int, 0, nfingers)
	for i, k := range keys {
		if i%16 == 0 {
			fingers = append(fingers, w.Len())
		}
		it, _ := ht.Get(k)
		it.Write(w)
	}
	verify.That(len(fingers) == nfingers)
	for _, f := range fingers {
		w2.Put3(f) // update with actual values
	}
	return off
}

func ReadInfoHamt(st *stor.Stor, off uint64) InfoHamt {
	r := st.Reader(off)
	nitems := r.Get2()
	t := InfoHamt{}.Mutable()
	if nitems == 0 {
		return t
	}
	nfingers := 1 + nitems/perFingerInfo
	for i := 0; i < nfingers; i++ {
		r.Get3() // skip the fingers
	}
	for i := 0; i < nitems; i++ {
		t.Put(ReadInfo(st, r))
	}
	return t.Freeze()
}

//-------------------------------------------------------------------

type InfoPacked struct {
	stor    *stor.Stor
	off     uint64
	buf     []byte
	fingers []InfoFinger
}

type InfoFinger struct {
	table string
	pos   int
}

func NewInfoPacked(st *stor.Stor, off uint64) *InfoPacked {
	buf := st.Data(off)
	r := stor.NewReader(buf)
	nitems := r.Get2()
	nfingers := 1 + nitems/perFingerInfo
	fingers := make([]InfoFinger, nfingers)
	for i := 0; i < nfingers; i++ {
		fingers[i].pos = r.Get3()
	}
	for i := 0; i < nfingers; i++ {
		fingers[i].table = stor.NewReader(buf[fingers[i].pos:]).GetStr()
	}
	return &InfoPacked{stor: st, off: off, buf: buf, fingers: fingers}
}

func (p InfoPacked) MustGet(key string) *Info {
	if item, ok := p.Get(key); ok {
		return item
	}
	panic("item not found")
}

func (p InfoPacked) Get(key string) (*Info, bool) {
	pos := p.binarySearch(key)
	r := stor.NewReader(p.buf[pos:])
	for n := 0; n <= perFingerInfo; n++ {
		item := ReadInfo(p.stor, r)
		if item.Table == key {
			return item, true
		}
	}
	var zero *Info
	return zero, false
}

// binarySearch does a binary search of the fingers
func (p InfoPacked) binarySearch(table string) int {
	i, j := 0, len(p.fingers)
	for i < j {
		h := int(uint(i+j) >> 1) // i ≤ h < j
		if table >= p.fingers[h].table {
			i = h + 1
		} else {
			j = h
		}
	}
	// i is first one greater, so we want i-1
	return int(p.fingers[i-1].pos)
}

func (p InfoPacked) Offset() uint64 {
	return p.off
}
