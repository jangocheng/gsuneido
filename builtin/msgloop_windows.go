package builtin

import (
	"runtime"
	"syscall"
	"unsafe"

	heap "github.com/apmckinlay/gsuneido/builtin/heapstack"
	. "github.com/apmckinlay/gsuneido/runtime"
)

var _ = builtin("MessageLoop(hwnd = 0)", func(t *Thread, args []Value) Value {
	MessageLoop(ToInt(args[0]))
	return nil
})

type MSG struct {
	hwnd    HANDLE
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      POINT
	_       [4]byte // padding
}

const nMSG = unsafe.Sizeof(MSG{})

var uiThreadId uintptr

const WM_USER = 0x400

var retChan = make(chan uintptr, 1)

func Init() {
	runtime.LockOSThread()
	uiThreadId = GetCurrentThreadId()
	heap.UIthread = uiThreadId // for debug
	heap.GetCurrentThreadId = GetCurrentThreadId
}

const GA_ROOT = 2
const WM_QUIT = 0x12
const GMEM_FIXED = 0

func MessageLoop(hdlg int) {
	defer heap.FreeTo(heap.CurSize())
	msg := (*MSG)(heap.Alloc(nMSG))
	for {
		if -1 == GetMessage(msg, 0, 0, 0) {
			continue // ignore error ???
		}
		if msg.message == WM_QUIT {
			if hdlg != 0 {
				PostQuitMessage(msg.wParam)
				return
			}
			return //TODO shutdown(msg.wParam)
		}
		if msg.message == WM_USER && msg.hwnd == 0 {
			// from SetTimer in another thread
			rtn, _, _ := syscall.Syscall6(setTimer, 4,
				0, 0, msg.wParam, msg.lParam, 0, 0)
			retChan <- rtn
			continue
		}
		if window := GetAncestor(msg.hwnd, GA_ROOT); window != 0 {
			// if (HACCEL haccel = (HACCEL) GetWindowLong(window, GWL_USERDATA))
			// 	if (TranslateAccelerator(window, haccel, &msg))
			// 		continue;
			if IsDialogMessage(window, msg) {
				continue
			}
		}
		TranslateMessage(msg)
		DispatchMessage(msg)
	}
}

func GetCurrentThreadId() uintptr {
	ret, _, _ := syscall.Syscall(getCurrentThreadId, 0, 0, 0, 0)
	return ret
}

var getMessage = user32.MustFindProc("GetMessageA").Addr()

func GetMessage(msg *MSG, hwnd HANDLE, msgFilterMin, msgFilterMax uint32) int {
	ret, _, _ := syscall.Syscall6(getMessage, 4,
		uintptr(unsafe.Pointer(msg)),
		uintptr(hwnd),
		uintptr(msgFilterMin),
		uintptr(msgFilterMax),
		0, 0)
	return int(ret)
}

var translateMessage = user32.MustFindProc("TranslateMessage").Addr()

func TranslateMessage(msg *MSG) bool {
	ret, _, _ := syscall.Syscall(translateMessage, 1,
		uintptr(unsafe.Pointer(msg)),
		0, 0)
	return ret != 0
}

var dispatchMessage = user32.MustFindProc("DispatchMessageA").Addr()

func DispatchMessage(msg *MSG) uintptr {
	ret, _, _ := syscall.Syscall(dispatchMessage, 1,
		uintptr(unsafe.Pointer(msg)),
		0, 0)
	return ret
}

func GetAncestor(hwnd HANDLE, gaFlags uint32) uintptr {
	ret, _, _ := syscall.Syscall(getAncestor, 2,
		uintptr(hwnd),
		uintptr(gaFlags),
		0)
	return ret
}

func IsDialogMessage(hwnd HANDLE, msg *MSG) bool {
	ret, _, _ := syscall.Syscall(isDialogMessage, 2,
		uintptr(hwnd),
		uintptr(unsafe.Pointer(msg)),
		0)
	return ret != 0
}

func PostQuitMessage(exitCode uintptr) uintptr {
	ret, _, _ := syscall.Syscall(postQuitMessage, 2,
		uintptr(exitCode),
		0, 0)
	return ret
}

//-------------------------------------------------------------------

// since we don't have the functions it calls
var _ = builtin0("WorkSpaceStatus()", func() Value {
	return EmptyStr
})
