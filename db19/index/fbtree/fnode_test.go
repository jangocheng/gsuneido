// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package fbtree

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/apmckinlay/gsuneido/db19/stor"
	"github.com/apmckinlay/gsuneido/util/assert"
	"github.com/apmckinlay/gsuneido/util/str"
)

func TestFAppendRead(t *testing.T) {
	type ent struct {
		offset uint64
		npre   int
		diff   string
	}
	var fn fnode
	var data []ent
	add := func(offset uint64, npre int, diff string) {
		fn = fn.append(offset, npre, diff)
		data = append(data, ent{offset, npre, diff})
	}
	add(123, 2, "bar")
	add(456, 1, "foo")
	for _, e := range data {
		var npre int
		var diff []byte
		var off uint64
		npre, diff, off = fn.read()
		fn = fn[fLen(diff):]
		assert.T(t).This(npre).Is(e.npre)
		assert.T(t).This(string(diff)).Is(e.diff)
		assert.T(t).This(off).Is(e.offset)
	}
}

func TestFnodeInsert(*testing.T) {
	datas := []string{
		"a b c d",
		"xa xb xc xd",
		"a ab abc abcd",
		"ant ants bun bunnies bunny buns cat a anti b bunn ca cats",
		"bbb bbc abc aa ab bc c aaa ba bba bb b a",
		"1000 1001 1002 1003",
	}
	for _, s := range datas {
		data := strings.Fields(s)
		get := func(i uint64) string { return data[i] }
		// forward
		fn := fnode{}
		for i, d := range data {
			fn = fn.insert(d, uint64(i), get)
			fn.checkUpTo(i, data, get)
		}
		assert.That(fn.check() == len(data))
		// reverse
		str.List(data).Reverse()
		fn = nil
		for i, d := range data {
			fn = fn.insert(d, uint64(i), get)
			fn.checkUpTo(i, data, get)
		}
		// builder
		fn = build(data)
		fn.checkData(data, get)
	}
}

func build(data []string) fnode {
	sort.Strings(data)
	b := fNodeBuilder{}
	for i, d := range data {
		b.Add(d, uint64(i), 255)
	}
	return b.Entries()
}

func (fn fnode) checkData(data []string, get func(uint64) string) {
	fn.checkUpTo(len(data)-1, data, get)
}

// checkUpTo is used during inserting.
// It checks that inserted keys are present
// and uninserted keys are not present.
func (fn fnode) checkUpTo(i int, data []string, get func(uint64) string) {
	n := fn.check()
	nn := 0
	for j, d := range data {
		if (d != "" && j <= i) != fn.contains(d, get) {
			panic("can't find " + d)
		}
		if d != "" && j <= i {
			nn++
		}
	}
	if nn != n {
		panic("check count expected " + strconv.Itoa(n) +
			" got " + strconv.Itoa(nn))
	}
}

func TestFnodeRandom(*testing.T) {
	const nData = 100
	var nGenerate = 20
	var nShuffle = 20
	if testing.Short() {
		nGenerate = 1
		nShuffle = 4
	}
	var data = make([]string, nData)
	get := func(i uint64) string { return data[i] }
	for gi := 0; gi < nGenerate; gi++ {
		data = data[0:nData]
		randKey := str.UniqueRandomOf(1, 6, "abcdef")
		for di := 0; di < nData; di++ {
			data[di] = randKey()
		}
		for si := 0; si < nShuffle; si++ {
			rand.Shuffle(len(data),
				func(i, j int) { data[i], data[j] = data[j], data[i] })
			var fn fnode
			for i, d := range data {
				fn = fn.insert(d, uint64(i), get)
				// fe.checkUpTo(i, data, get)
			}
			fn.checkData(data, get)
		}
	}
}

func TestDelete(*testing.T) {
	var fn fnode
	const nData = 8 + 32
	var data = make([]string, nData)
	get := func(i uint64) string { return data[i] }
	randKey := str.UniqueRandomOf(1, 6, "abcdef")
	for i := 0; i < nData; i++ {
		data[i] = randKey()
	}
	sort.Strings(data)
	for i := 0; i < len(data); i++ {
		fn = fn.insert(data[i], uint64(i), get)
	}
	// fn.printLeafNode(get)

	var ok bool

	// delete at end, simplest case
	for i := 0; i < 8; i++ {
		fn, ok = fn.delete(uint64(len(data) - 1))
		assert.That(ok)
		data = data[:len(data)-1]
		fn.checkData(data, get)
	}
	// print("================================")
	// fn.printLeafNode(get)

	// delete at start
	const nStart = 8
	for i := 0; i < nStart; i++ {
		fn, ok = fn.delete(uint64(i))
		assert.That(ok)
		data[i] = ""
		fn.checkData(data, get)
	}
	// print("================================")
	// fn.printLeafNode(get)

	for i := 0; i < len(data)-nStart; i++ {
		off := rand.Intn(len(data))
		for data[off] == "" {
			off = (off + 1) % len(data)
		}
		// print("================================ delete", data[off])
		fn, ok = fn.delete(uint64(off))
		assert.That(ok)
		// fn.printLeafNode(get)
		data[off] = ""
		fn.checkData(data, get)
	}
}

func TestDelete2(*testing.T) {
	data := []string{"a", "b", "c", "d", "e"}
	get := func(i uint64) string { return data[i] }
	var fn fnode
	for i := 0; i < len(data); i++ {
		fn = fn.insert(data[i], uint64(i), get)
	}
	// fn.printLeafNode(get)

	var ok bool
	for i := 1; i < len(data); i++ {
		fn, ok = fn.delete(uint64(i))
		assert.That(ok)
		// print("================================")
		// fn.printLeafNode(get)
		data[i] = ""
		fn.checkData(data, get)
	}
}

func TestWords(*testing.T) {
	data := words
	const nShuffle = 100
	get := func(i uint64) string { return data[i] }
	for si := 0; si < nShuffle; si++ {
		rand.Shuffle(len(data),
			func(i, j int) { data[i], data[j] = data[j], data[i] })
		var fn fnode
		for i, d := range data {
			fn = fn.insert(d, uint64(i), get)
			// fe.checkUpto(i, data, get)
		}
		fn.checkData(data, get)
	}
}

var words = []string{
	"tract",
	"pluck",
	"rumor",
	"choke",
	"abbey",
	"robot",
	"north",
	"dress",
	"pride",
	"dream",
	"judge",
	"coast",
	"frank",
	"suite",
	"merit",
	"chest",
	"youth",
	"throw",
	"drown",
	"power",
	"ferry",
	"waist",
	"moral",
	"woman",
	"swipe",
	"straw",
	"shell",
	"class",
	"claim",
	"tired",
	"stand",
	"chaos",
	"shame",
	"thigh",
	"bring",
	"lodge",
	"amuse",
	"arrow",
	"charm",
	"swarm",
	"serve",
	"world",
	"raise",
	"means",
	"honor",
	"grand",
	"stock",
	"model",
	"greet",
	"basic",
	"fence",
	"fight",
	"level",
	"title",
	"knife",
	"wreck",
	"agony",
	"white",
	"child",
	"sport",
	"cheat",
	"value",
	"marsh",
	"slide",
	"tempt",
	"catch",
	"valid",
	"study",
	"crack",
	"swing",
	"plead",
	"flush",
	"awful",
	"house",
	"stage",
	"fever",
	"equal",
	"fault",
	"mouth",
	"mercy",
	"colon",
	"belly",
	"flash",
	"style",
	"plant",
	"quote",
	"pitch",
	"lobby",
	"gloom",
	"patch",
	"crime",
	"anger",
	"petty",
	"spend",
	"strap",
	"novel",
	"sword",
	"match",
	"tasty",
	"stick",
}

var S1 []byte
var S2 []byte
var FN fnode

func BenchmarkFnode(b *testing.B) {
	get := func(i uint64) string { return words[i] }
	var fn fnode
	for i, d := range words {
		fn = fn.insert(d, uint64(i), get)
	}
	FN = fn

	for i := 0; i < b.N; i++ {
		iter := fn.iter()
		for iter.next() {
			S1 = iter.known
			S2 = iter.diff
		}
	}
}

func ExampleFnodeBuilderSplit() {
	var fb fNodeBuilder
	fb.Add("1234xxxx", 1234, 1)
	fb.Add("1235xxxx", 1235, 1)
	fb.Add("1299xxxx", 1299, 1)
	fb.Add("1300xxxx", 1300, 1)
	fb.Add("1305xxxx", 1305, 1)
	store := stor.HeapStor(8192)
	leftOff, splitKey := fb.Split(store)
	// assert.T(t).This(splitKey).Is("13")
	fmt.Println("splitKey", splitKey)
	fmt.Println("LEFT ---")
	readNode(store, leftOff).print()
	fmt.Println("RIGHT ---")
	fb.fe.print()

	// Output:
	// splitKey 13
	// LEFT ---
	// 1234 ''
	// 1235 1235
	// 1299 129
	// RIGHT ---
	// 1300 ''
	// 1305 1305
}

func ExampleFbmergeSplit() {
	var fb fNodeBuilder
	fb.Add("1234xxxx", 1234, 1)
	fb.Add("1235xxxx", 1235, 1)
	fb.Add("1299xxxx", 1299, 1)
	fb.Add("1300xxxx", 1300, 1)
	fb.Add("1305xxxx", 1305, 1)
	m := merge{node: fb.fe, modified: true}
	left, right, splitKey := m.split()
	// assert.T(t).This(splitKey).Is("13")
	fmt.Println("splitKey", splitKey)
	fmt.Println("LEFT ---")
	left.print()
	fmt.Println("RIGHT ---")
	right.print()

	// Output:
	// splitKey 13
	// LEFT ---
	// 1234 ''
	// 1235 1235
	// 1299 129
	// RIGHT ---
	// 1300 ''
	// 1305 1305
}