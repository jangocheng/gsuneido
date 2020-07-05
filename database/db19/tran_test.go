// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package db19

import (
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	. "github.com/apmckinlay/gsuneido/util/hamcrest"

	"github.com/apmckinlay/gsuneido/database/db19/btree"
	"github.com/apmckinlay/gsuneido/database/db19/meta"
	"github.com/apmckinlay/gsuneido/database/db19/stor"
)

func TestConcurrent(t *testing.T) {
	createDb()
	c := StartConcur(100 * time.Millisecond)
	var nclients = 8
	var ntrans = 4000
	if testing.Short() {
		nclients = 4
		ntrans = 100
	}
	var wg sync.WaitGroup
	for i := 0; i < nclients; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < ntrans; j++ {
				ut := output1()
				c.commitChan <- ut
				// time.Sleep(time.Duration(rand.Intn(900)) * time.Microsecond)
			}
		}()
	}
	wg.Wait()
	c.Stop()
	// fmt.Println("under", btree.Under)

	var nout = nclients * ntrans
	rt := NewReadTran()
	ti := rt.meta.GetRoInfo("mytable")
	Assert(t).That(ti.Nrows, Equals(nout).Comment("nrows"))
	Assert(t).That(ti.Size, Equals(nout*12).Comment("size"))

	GetState().store.Close()
}

func TestTran(t *testing.T) {
	store := createDb()

	const nout = 100
	for i := 0; i < nout; i++ {
		Merge(output1().Commit())
	}

	Persist()
	store.Close()

	f, _ := os.OpenFile("test.tmp", os.O_WRONLY|os.O_APPEND, 0644)
	f.Write([]byte("some garbage to add on the end of the file"))
	f.Close()

	store, _ = stor.MmapStor("test.tmp", stor.UPDATE)
	off := store.LastOffset([]byte(magic1))
	UpdateState(func(state *DbState) {
		*state = *ReadState(store, off)
	})

	rt := NewReadTran()
	ti := rt.meta.GetRoInfo("mytable")
	Assert(t).That(ti.Nrows, Equals(nout).Comment("nrows"))
	Assert(t).That(ti.Size, Equals(nout*12).Comment("size"))

	store.Close()
	os.Remove("test.tmp")
}

func createDb() *stor.Stor {
	btree.GetLeafKey = func(st *stor.Stor, _ interface{}, off uint64) string {
		return string(st.DataSized(off))
	}

	// store := stor.HeapStor(16 * 1024)
	store, err := stor.MmapStor("test.tmp", stor.CREATE)
	if err != nil {
		panic("can't create test.tmp")
	}

	schema := meta.NewSchemaHtbl(0)
	schema.Put(&meta.Schema{
		Table: "mytable",
		Columns: []meta.ColumnSchema{
			{Name: "one", Field: 0},
			{Name: "two", Field: 1},
		},
		Indexes: []meta.IndexSchema{
			{Fields: []int{0}},
		},
	})
	baseSchema := meta.NewSchemaPacked(store, schema.Write(store))

	info := meta.NewInfoHtbl(0)
	info.Put(&meta.Info{
		Table:   "mytable",
		Indexes: []*btree.Overlay{btree.NewOverlay(store).Save()},
	})
	baseInfo := meta.NewInfoPacked(store, info.Write(store))

	roSchema := meta.NewSchemaHtbl(0)
	roSchemaOff := roSchema.Write(store)
	roInfo := meta.NewInfoHtbl(0)
	roInfoOff := roInfo.Write(store)

	UpdateState(func(state *DbState) {
		state.store = store
		state.meta = meta.NewOverlay(baseSchema, baseInfo,
			roSchema, roSchemaOff, roInfo, roInfoOff, nil)
	})

	return store
}

func output1() *UpdateTran {
	ut := NewUpdateTran()
	// write some data
	data := (strconv.Itoa(ut.num) + "transaction")[:12]
	off, buf := ut.store.AllocSized(len(data))
	copy(buf, data)
	// add it to the indexes
	ti := ut.meta.GetRwInfo("mytable", ut.num)
	ti.Nrows++
	ti.Size += uint64(len(data))
	ti.Indexes[0].Insert(data, off)
	return ut
	// NOTE: does not commit
}