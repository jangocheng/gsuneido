// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package metadata

import (
	"sort"

	"github.com/apmckinlay/gsuneido/database/stor"
	"github.com/apmckinlay/gsuneido/util/verify"
)

//go:generate genny -in ../../genny/flathash/flathash.go -out tables.go -pkg metadata gen "Key=int Item=TableInfo"

type TableInfo struct {
	table   int
	nrows   int
	size    uint64
	indexes []IndexInfo
}

type IndexInfo struct {
	root       uint64 //TODO should be fbtree
	treeLevels int
}

func (*TableInfoHtbl) hash(k int) uint32 {
	return uint32(k)
}

func (*TableInfoHtbl) keyOf(ti *TableInfo) int {
	return ti.table
}

//-------------------------------------------------------------------

// Merge combines two TableInfoHtbl into a new one.
// t2 takes precedence.
func (t *TableInfoHtbl) Merge(t2 *TableInfoHtbl) *TableInfoHtbl {
	// Important - bulk copy rather than inserting individually
	t3 := t.Dup()
	iter := t2.Iter()
	for ti2 := iter(); ti2 != nil; ti2 = iter() {
		ti := t.Get(ti2.table)
		if ti == nil {
			t3.Put(ti2)
		} else {
			t3.Put(ti.Merge(ti2))
		}
	}
	return t3
}

// Merge combines two TableInfo into a new one.
// ti2 takes precedence.
func (ti *TableInfo) Merge(ti2 *TableInfo) *TableInfo {
	return &TableInfo{table: ti2.table,
		nrows:   ti.nrows + ti2.nrows,
		size:    ti.size + ti2.size,
		indexes: append([]IndexInfo(nil), ti2.indexes...),
	}
}

//-------------------------------------------------------------------

const blockSize = 4 * 1024
const itemsPerFinger = 16

// Write converts a TableInfoHtbl to external packed format in a byte slice.
// Tables are sorted by table number.
func (t *TableInfoHtbl) Write(st *stor.Stor) uint64 {
	w := st.Writer(blockSize)
	keys := t.List()
	sort.Ints(keys)
	w.Put2(t.nitems)
	nfingers := 1 + t.nitems/itemsPerFinger
	w2 := *w
	for i := 0; i < nfingers; i++ {
		w.Put3(0) // leave room
	}
	fingers := make([]int, 0, nfingers)
	for i, k := range keys {
		if i%16 == 0 {
			fingers = append(fingers, w.Pos())
		}
		t.Get(k).Write(w)
	}
	verify.That(len(fingers) == nfingers)
	for _, f := range fingers {
		w2.Put3(f) // update with actual values
	}
	return w.Close()
}

// Write converts a TableInfo to external packed format in a byte slice
func (ti *TableInfo) Write(w *stor.Writer) {
	w.Put3(ti.table).
		Put4(ti.nrows).
		Put5(ti.size).
		Put1(len(ti.indexes))
	for _, ii := range ti.indexes {
		w.Put5(ii.root).Put1(ii.treeLevels)
	}
}

func ReadTablesInfo(st *stor.Stor, off uint64) *TableInfoHtbl {
	r := st.Reader(off)
	nitems := r.Get2()
	nfingers := 1 + nitems/itemsPerFinger
	for i := 0; i < nfingers; i++ {
		r.Get3() // skip the fingers
	}
	t := NewTableInfoHtbl(nitems)
	for i := 0; i < nitems; i++ {
		t.Put(ReadTableInfo(r))
	}
	return t
}

func ReadTableInfo(r *stor.Reader) *TableInfo {
	var ti TableInfo
	ti.table = r.Get3()
	ti.nrows = r.Get4()
	ti.size = r.Get5()
	ni := r.Get1()
	ti.indexes = make([]IndexInfo, ni)
	for i := 0; i < ni; i++ {
		ti.indexes[i].Read(r)
	}
	return &ti
}

func (ii *IndexInfo) Read(r *stor.Reader) {
	ii.root = r.Get5()
	ii.treeLevels = r.Get1()
}

//-------------------------------------------------------------------

type TableInfoPacked struct {
	r       *stor.Reader
	fingers []finger
}

type finger struct {
	table int
	pos   int
}

func NewTableInfoPacked(st *stor.Stor, off uint64) *TableInfoPacked {
	r := st.Reader(off)
	nitems := r.Get2()
	nfingers := 1 + nitems/itemsPerFinger
	fingers := make([]finger, nfingers)
	for i := 0; i < nfingers; i++ {
		fingers[i].pos = r.Get3()
	}
	for i := 0; i < nfingers; i++ {
		fingers[i].table = r.Pos(fingers[i].pos).Get3()
	}
	return &TableInfoPacked{r: r, fingers: fingers}
}

func (p TableInfoPacked) Get(table int) *TableInfo {
	p.r.Pos(p.binarySearch(table))
	count := 0
	for {
		ti := ReadTableInfo(p.r)
		if ti.table == table {
			return ti
		}
		count++
		if count > 20 {
			panic("linear search too long")
		}
	}
}

func (p TableInfoPacked) binarySearch(table int) int {
	i, j := 0, len(p.fingers)
	count := 0
	for i < j {
		h := int(uint(i+j) >> 1) // i ≤ h < j
		if table >= p.fingers[h].table {
			i = h + 1
		} else {
			j = h
		}
		count++
		if count > 20 {
			panic("binary search too long")
		}
	}
	// i is first one greater, so we want i-1
	return int(p.fingers[i-1].pos)
}
