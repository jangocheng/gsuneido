package base

import (
	"strconv"

	"github.com/apmckinlay/gsuneido/util/dnum"
	"github.com/apmckinlay/gsuneido/util/hash"
)

// SuConcat is a Value used to optimize string concatenation
// NOTE: Not thread safe
type SuConcat struct {
	b *shared
	n int
}

type shared struct {
	a []byte
	// MAYBE have a string to cache?
}

var _ Value = SuConcat{}
var _ Packable = SuConcat{}

// NewSuConcat returns an empty SuConcat
func NewSuConcat() SuConcat {
	return SuConcat{b: &shared{}}
}

// Len returns the length of an SuConcat
func (c SuConcat) Len() int {
	return c.n
}

// Add appends a string to an SuConcat
func (c SuConcat) Add(s string) SuConcat {
	bb := c.b
	if len(bb.a) != c.n {
		// another reference has appended their own stuff so make our own buf
		a := append(make([]byte, 0, c.n+len(s)), bb.a[:c.n]...)
		bb = &shared{a}
	}
	bb.a = append(bb.a, s...)
	return SuConcat{bb, c.n + len(s)}
}

// AddSuConcat appends an SuConcat to an SuConcat
func (c SuConcat) AddSuConcat(cv2 SuConcat) SuConcat {
	// avoid converting cv2 to string
	bb := c.b
	if len(bb.a) != c.n {
		// another reference has appended their own stuff so make our own buf
		a := append(make([]byte, 0, c.n+cv2.Len()), bb.a[:c.n]...)
		bb = &shared{a}
	}
	bb.a = append(bb.a, cv2.b.a...)
	return SuConcat{bb, c.n + cv2.Len()}
}

// Value interface --------------------------------------------------

// ToInt converts an SuConcat to an integer (Value interface)
func (c SuConcat) ToInt() int {
	i, _ := strconv.ParseInt(c.ToStr(), 0, 32)
	return int(i)
}

// ToDnum converts an SuConcat to a Dnum (Value interface)
func (c SuConcat) ToDnum() dnum.Dnum {
	return dnum.FromStr(c.ToStr())
}

// ToStr converts an SuConcat to a string (Value interface)
func (c SuConcat) ToStr() string {
	return string(c.b.a[:c.n])
}

// String returns a quoted string (Value interface)
// TODO: handle escaping
func (c SuConcat) String() string {
	return "'" + c.ToStr() + "'"
}

// Get returns the character at a given index (Value interface)
func (c SuConcat) Get(key Value) Value {
	return SuStr(string(c.b.a[:c.n][key.ToInt()]))
}

// Put is not applicable to SuConcat (Value interface)
func (SuConcat) Put(Value, Value) {
	panic("strings do not support put")
}

// Hash returns a hash value for an SuConcat (Value interface)
func (c SuConcat) Hash() uint32 {
	return hash.HashBytes(c.b.a[:c.n])
}

// hash2 is used to hash nested values (Value interface)
func (c SuConcat) hash2() uint32 {
	return c.Hash()
}

// Equals returns true if other is an equal SuConcat or SuStr (Value interface)
func (c SuConcat) Equals(other interface{}) bool {
	if c2, ok := other.(SuConcat); ok {
		return c == c2 // FIXME: slices aren't comparable
	}
	if s2, ok := other.(SuStr); ok && c.n == len(s2) {
		for i := 0; i < c.n; i++ {
			if c.b.a[i] != string(s2)[i] {
				return false
			}
			return true
		}
	}
	return false
}

// TypeName returns the name of this type (Value interface)
func (SuConcat) TypeName() string {
	return "String"
}

// Order returns the ordering of SuDnum (Value interface)
func (SuConcat) Order() ord {
	return ordStr
}

// Cmp compares an SuDnum to another Value (Value interface)
func (c SuConcat) Cmp(other Value) int {
	// COULD optimize this to not convert Concat to string
	s1 := c.ToStr()
	s2 := other.ToStr()
	switch {
	case s1 < s2:
		return -1
	case s1 > s2:
		return +1
	default:
		return 0
	}
}

// Packable interface -----------------------------------------------

func (c SuConcat) PackSize() int {
	if c.n == 0 {
		return 0
	}
	return 1 + c.n
}

func (c SuConcat) Pack(buf []byte) []byte {
	buf = append(buf, packString)
	buf = append(buf, c.b.a[:c.n]...)
	return buf
}