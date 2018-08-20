package base

import (
	"math"

	"github.com/apmckinlay/gsuneido/util/ints"

	"github.com/apmckinlay/gsuneido/util/dnum"
	"github.com/apmckinlay/gsuneido/util/regex"
)

func Is(x Value, y Value) Value {
	return SuBool(x.Equals(y))
}

func Isnt(x Value, y Value) Value {
	return SuBool(!x.Equals(y))
}

func Lt(x Value, y Value) Value {
	return SuBool(x.Cmp(y) < 0)
}

func Lte(x Value, y Value) Value {
	return SuBool(x.Cmp(y) <= 0)
}

func Gt(x Value, y Value) Value {
	return SuBool(x.Cmp(y) > 0)
}

func Gte(x Value, y Value) Value {
	return SuBool(x.Cmp(y) >= 0)
}

func Add(x Value, y Value) Value {
	if xi, xok := SmiToInt(x); xok {
		if yi, yok := SmiToInt(y); yok {
			return IntToValue(xi + yi)
		}
	}
	return SuDnum{dnum.Add(x.ToDnum(), y.ToDnum())}
}

func Sub(x Value, y Value) Value {
	if xi, xok := SmiToInt(x); xok {
		if yi, yok := SmiToInt(y); yok {
			return IntToValue(xi - yi)
		}
	}
	return SuDnum{dnum.Sub(x.ToDnum(), y.ToDnum())}
}

func Mul(x Value, y Value) Value {
	if xi, xok := SmiToInt(x); xok {
		if yi, yok := SmiToInt(y); yok {
			return IntToValue(xi * yi)
		}
	}
	return SuDnum{dnum.Mul(x.ToDnum(), y.ToDnum())}
}

func Div(x Value, y Value) Value {
	if xi, xok := SmiToInt(x); xok {
		if yi, yok := SmiToInt(y); yok {
			if xi % yi == 0 {
				return IntToValue(xi / yi)
			}
		}
	}
	return SuDnum{dnum.Div(x.ToDnum(), y.ToDnum())}
}

func Mod(x Value, y Value) Value {
	return IntToValue(x.ToInt() % y.ToInt())
}

func Lshift(x Value, y Value) Value {
	return IntToValue(int(uint(x.ToInt()) << uint(y.ToInt())))
}

func Rshift(x Value, y Value) Value {
	return IntToValue(int(uint(x.ToInt()) >> uint(y.ToInt())))
}

func Bitor(x Value, y Value) Value {
	return IntToValue(x.ToInt() | y.ToInt())
}

func Bitand(x Value, y Value) Value {
	return IntToValue(x.ToInt() & y.ToInt())
}

func Bitxor(x Value, y Value) Value {
	return IntToValue(x.ToInt() ^ y.ToInt())
}

func Bitnot(x Value) Value {
	return IntToValue(^x.ToInt())
}

func Not(x Value) Value {
	if x == True {
		return False
	} else if x == False {
		return True
	}
	panic("not requires boolean")
}

func Uplus(x Value) Value {
	if _, ok := SmiToInt(x); ok {
		return x
	} else if _, ok := x.(SuDnum); ok {
		return x
	}
	panic("can't convert to number")
}

func Uminus(x Value) Value {
	if xi, ok := SmiToInt(x); ok {
		return IntToValue(-xi)
	}
	return SuDnum{x.ToDnum().Neg()}
}

// IntToValue returns an SuInt if it fits, else a SuDnum
func IntToValue(n int) Value {
	if math.MinInt16 < n && n < math.MaxInt16 {
		return SuInt(n)
	}
	return SuDnum{dnum.FromInt(int64(n))}
}

func Cat(x Value, y Value) Value {
	const SMALL = 256

	xc, xcok := x.(SuConcat)
	yc, ycok := y.(SuConcat)
	if xcok && ycok {
		return xc.AddSuConcat(yc)
	} else if xcok {
		return xc.Add(y.ToStr())
	} else if ycok {
		return NewSuConcat().Add(x.ToStr()).AddSuConcat(yc)
	}
	xs := x.ToStr()
	ys := y.ToStr()
	if len(xs)+len(ys) < SMALL {
		return SuStr(xs + ys)
	}
	return NewSuConcat().Add(xs).Add(ys)
}

func BitNot(x Value) Value {
	return IntToValue(^x.ToInt())
}

func Cmp(x Value, y Value) int {
	xo := x.Order()
	yo := y.Order()
	if xo != yo {
		return ints.Compare(int(xo), int(yo))
	}
	return x.Cmp(y)
}

func Match(x Value, y regex.Pattern) SuBool {
	return SuBool(y.Matches(x.ToStr()))
}