package str

import (
	"math/rand"
	"strings"
)

const alpha = "abcdefghijklmnopqrstuvwxyz"

func Random(min, max int) string {
	return RandomOf(min, max, alpha)
}

func RandomOf(min, max int, chars string) string {
	n := min + rand.Intn(1+max-min)
	var b strings.Builder
	b.Grow(n)
	for i := 0; i < n; i++ {
		b.WriteByte(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

func UniqueRandom(min, max int) func() string {
	return UniqueRandomOf(min, max, alpha)
}

func UniqueRandomOf(min, max int, chars string) func() string {
	prev := map[string]bool{}
	return func() string {
		var key string
		for i := 0; i < 10; i++ {
			key = RandomOf(min, max, chars)
			if !prev[key] {
				prev[key] = true
				return key
			}
		}
		panic("str.UniqueRandomOf too many duplicates")
	}
}
