package goc

// #cgo CFLAGS: -DWINVER=0x601 -D_WIN32_WINNT=0x0601
// #cgo LDFLAGS: -lurlmon -lole32 -luuid -lwininet -static
// #include "cside.h"
import "C"

import (
	"log"
	"runtime"
	"strconv"
	"unsafe"

	"golang.org/x/sys/windows"
)

const Ncb2s = C.ncb2s
const Ncb3s = C.ncb3s
const Ncb4s = C.ncb4s

var uiThreadId uint32

func Init() {
	runtime.LockOSThread()
	uiThreadId = windows.GetCurrentThreadId()
	C.start()
}

func Run() {
	C.args[0] = C.msg_result
	C.signalAndWait()
	interact()
	log.Fatalln("!!! should not reach here !!!")
}

func CThreadId() uintptr {
	return uintptr(C.threadid)
}

func GetCallback(nargs, i int) uintptr {
	switch nargs {
	case 2:
		return uintptr(C.cb2s[i])
	case 3:
		return uintptr(C.cb3s[i])
	case 4:
		return uintptr(C.cb4s[i])
	}
	panic("GetCallback unsupported nargs " + strconv.Itoa(nargs))
}

// must be injected
var TimerId func(a uintptr)
var Callback2 func(i, a, b uintptr) uintptr
var Callback3 func(i, a, b, c uintptr) uintptr
var Callback4 func(i, a, b, c, d uintptr) uintptr
var UpdateUI func()
var SunAPP func(string) string

func interact() uintptr {
	//TODO use Suneido thread instead
	if uiThreadId != windows.GetCurrentThreadId() {
		panic("illegal UI call from background thread")
	}
	for {
		switch C.args[0] {
		case C.msg_syscall:
			break
		case C.msg_callback2:
			C.args[0] = C.msg_result
			C.args[1] = C.uintptr(Callback2(uintptr(C.args[1]),
				uintptr(C.args[2]), uintptr(C.args[3])))
		case C.msg_callback3:
			C.args[0] = C.msg_result
			C.args[1] = C.uintptr(Callback3(uintptr(C.args[1]),
				uintptr(C.args[2]), uintptr(C.args[3]), uintptr(C.args[4])))
		case C.msg_callback4:
			C.args[0] = C.msg_result
			C.args[1] = C.uintptr(Callback4(uintptr(C.args[1]),
				uintptr(C.args[2]), uintptr(C.args[3]), uintptr(C.args[4]),
				uintptr(C.args[5])))
		case C.msg_timerid:
			TimerId(uintptr(C.args[1]))
			C.args[0] = C.msg_result
		case C.msg_updateui:
			UpdateUI()
			C.args[0] = C.msg_result
		case C.msg_sunapp:
			s := SunAPP(C.GoString((*C.char)(unsafe.Pointer(uintptr(C.args[1])))))
			C.args[0] = C.msg_result
			C.args[1] = (C.uintptr)(uintptr(unsafe.Pointer(C.CString(s))))
			C.args[2] = (C.uintptr)(len(s))
		case C.msg_result:
			return uintptr(C.args[1])
		}
		C.signalAndWait()
	}
}

func MessageLoop(hdlg uintptr) {
	C.args[0] = C.msg_msgloop
	C.args[1] = C.uintptr(hdlg)
	interact()
}

func Syscall0(adr uintptr) uintptr {
	C.args[0] = C.msg_syscall
	C.args[1] = C.uintptr(adr)
	C.args[2] = 0
	return interact()
}
func Syscall1(adr, a uintptr) uintptr {
	C.args[0] = C.msg_syscall
	C.args[1] = C.uintptr(adr)
	C.args[2] = 1
	C.args[3] = C.uintptr(a)
	return interact()
}
func Syscall2(adr, a, b uintptr) uintptr {
	C.args[0] = C.msg_syscall
	C.args[1] = C.uintptr(adr)
	C.args[2] = 2
	C.args[3] = C.uintptr(a)
	C.args[4] = C.uintptr(b)
	return interact()
}
func Syscall3(adr, a, b, c uintptr) uintptr {
	C.args[0] = C.msg_syscall
	C.args[1] = C.uintptr(adr)
	C.args[2] = 3
	C.args[3] = C.uintptr(a)
	C.args[4] = C.uintptr(b)
	C.args[5] = C.uintptr(c)
	return interact()
}
func Syscall4(adr, a, b, c, d uintptr) uintptr {
	C.args[0] = C.msg_syscall
	C.args[1] = C.uintptr(adr)
	C.args[2] = 4
	C.args[3] = C.uintptr(a)
	C.args[4] = C.uintptr(b)
	C.args[5] = C.uintptr(c)
	C.args[6] = C.uintptr(d)
	return interact()
}
func Syscall5(adr, a, b, c, d, e uintptr) uintptr {
	C.args[0] = C.msg_syscall
	C.args[1] = C.uintptr(adr)
	C.args[2] = 5
	C.args[3] = C.uintptr(a)
	C.args[4] = C.uintptr(b)
	C.args[5] = C.uintptr(c)
	C.args[6] = C.uintptr(d)
	C.args[7] = C.uintptr(e)
	return interact()
}
func Syscall6(adr, a, b, c, d, e, f uintptr) uintptr {
	C.args[0] = C.msg_syscall
	C.args[1] = C.uintptr(adr)
	C.args[2] = 6
	C.args[3] = C.uintptr(a)
	C.args[4] = C.uintptr(b)
	C.args[5] = C.uintptr(c)
	C.args[6] = C.uintptr(d)
	C.args[7] = C.uintptr(e)
	C.args[8] = C.uintptr(f)
	return interact()
}
func Syscall7(adr, a, b, c, d, e, f, g uintptr) uintptr {
	C.args[0] = C.msg_syscall
	C.args[1] = C.uintptr(adr)
	C.args[2] = 7
	C.args[3] = C.uintptr(a)
	C.args[4] = C.uintptr(b)
	C.args[5] = C.uintptr(c)
	C.args[6] = C.uintptr(d)
	C.args[7] = C.uintptr(e)
	C.args[8] = C.uintptr(f)
	C.args[9] = C.uintptr(g)
	return interact()
}
func Syscall8(adr, a, b, c, d, e, f, g, h uintptr) uintptr {
	C.args[0] = C.msg_syscall
	C.args[1] = C.uintptr(adr)
	C.args[2] = 8
	C.args[3] = C.uintptr(a)
	C.args[4] = C.uintptr(b)
	C.args[5] = C.uintptr(c)
	C.args[6] = C.uintptr(d)
	C.args[7] = C.uintptr(e)
	C.args[8] = C.uintptr(f)
	C.args[9] = C.uintptr(g)
	C.args[10] = C.uintptr(h)
	return interact()
}
func Syscall9(adr, a, b, c, d, e, f, g, h, i uintptr) uintptr {
	C.args[0] = C.msg_syscall
	C.args[1] = C.uintptr(adr)
	C.args[2] = 9
	C.args[3] = C.uintptr(a)
	C.args[4] = C.uintptr(b)
	C.args[5] = C.uintptr(c)
	C.args[6] = C.uintptr(d)
	C.args[7] = C.uintptr(e)
	C.args[8] = C.uintptr(f)
	C.args[9] = C.uintptr(g)
	C.args[10] = C.uintptr(h)
	C.args[11] = C.uintptr(i)
	return interact()
}
func Syscall10(adr, a, b, c, d, e, f, g, h, i, j uintptr) uintptr {
	C.args[0] = C.msg_syscall
	C.args[1] = C.uintptr(adr)
	C.args[2] = 10
	C.args[3] = C.uintptr(a)
	C.args[4] = C.uintptr(b)
	C.args[5] = C.uintptr(c)
	C.args[6] = C.uintptr(d)
	C.args[7] = C.uintptr(e)
	C.args[8] = C.uintptr(f)
	C.args[9] = C.uintptr(g)
	C.args[10] = C.uintptr(h)
	C.args[11] = C.uintptr(i)
	C.args[12] = C.uintptr(j)
	return interact()
}
func Syscall11(adr, a, b, c, d, e, f, g, h, i, j, k uintptr) uintptr {
	C.args[0] = C.msg_syscall
	C.args[1] = C.uintptr(adr)
	C.args[2] = 11
	C.args[3] = C.uintptr(a)
	C.args[4] = C.uintptr(b)
	C.args[5] = C.uintptr(c)
	C.args[6] = C.uintptr(d)
	C.args[7] = C.uintptr(e)
	C.args[8] = C.uintptr(f)
	C.args[9] = C.uintptr(g)
	C.args[10] = C.uintptr(h)
	C.args[11] = C.uintptr(i)
	C.args[12] = C.uintptr(j)
	C.args[13] = C.uintptr(k)
	return interact()
}
func Syscall12(adr, a, b, c, d, e, f, g, h, i, j, k, l uintptr) uintptr {
	C.args[0] = C.msg_syscall
	C.args[1] = C.uintptr(adr)
	C.args[2] = 12
	C.args[3] = C.uintptr(a)
	C.args[4] = C.uintptr(b)
	C.args[5] = C.uintptr(c)
	C.args[6] = C.uintptr(d)
	C.args[7] = C.uintptr(e)
	C.args[8] = C.uintptr(f)
	C.args[9] = C.uintptr(g)
	C.args[10] = C.uintptr(h)
	C.args[11] = C.uintptr(i)
	C.args[12] = C.uintptr(j)
	C.args[13] = C.uintptr(k)
	C.args[14] = C.uintptr(l)
	return interact()
}
func Syscall14(adr, a, b, c, d, e, f, g, h, i, j, k, l, m, n uintptr) uintptr {
	C.args[0] = C.msg_syscall
	C.args[1] = C.uintptr(adr)
	C.args[2] = 14
	C.args[3] = C.uintptr(a)
	C.args[4] = C.uintptr(b)
	C.args[5] = C.uintptr(c)
	C.args[6] = C.uintptr(d)
	C.args[7] = C.uintptr(e)
	C.args[8] = C.uintptr(f)
	C.args[9] = C.uintptr(g)
	C.args[10] = C.uintptr(h)
	C.args[11] = C.uintptr(i)
	C.args[12] = C.uintptr(j)
	C.args[13] = C.uintptr(k)
	C.args[14] = C.uintptr(l)
	C.args[15] = C.uintptr(m)
	C.args[16] = C.uintptr(n)
	return interact()
}
