package w32

import (
	"fmt"
	"log"
	"strconv"
	"syscall"
	"unsafe"
)

const (
	GCLP_HBRBACKGROUND int32 = -10
)

func ExtendFrameIntoClientArea(hwnd uintptr, extend bool) {
	// -1: Adds the default frame styling (aero shadow and e.g. rounded corners on Windows 11)
	//     Also shows the caption buttons if transparent ant translucent but they don't work.
	//  0: Adds the default frame styling but no aero shadow, does not show the caption buttons.
	//  1: Adds the default frame styling (aero shadow and e.g. rounded corners on Windows 11) but no caption buttons
	//     are shown if transparent ant translucent.
	var margins MARGINS
	if extend {
		margins = MARGINS{1, 1, 1, 1} // Only extend 1 pixel to have the default frame styling but no caption buttons
	}
	if err := dwmExtendFrameIntoClientArea(hwnd, &margins); err != nil {
		log.Fatal(fmt.Errorf("DwmExtendFrameIntoClientArea failed: %s", err))
	}
}

func IsVisible(hwnd uintptr) bool {
	ret, _, _ := procIsWindowVisible.Call(hwnd)
	return ret != 0
}

func IsWindowFullScreen(hwnd uintptr) bool {
	wRect := GetWindowRect(hwnd)
	m := MonitorFromWindow(hwnd, MONITOR_DEFAULTTOPRIMARY)
	var mi MONITORINFO
	mi.CbSize = uint32(unsafe.Sizeof(mi))
	if !GetMonitorInfo(m, &mi) {
		return false
	}
	return wRect.Left == mi.RcMonitor.Left &&
		wRect.Top == mi.RcMonitor.Top &&
		wRect.Right == mi.RcMonitor.Right &&
		wRect.Bottom == mi.RcMonitor.Bottom
}

func IsWindowMaximised(hwnd uintptr) bool {
	style := uint32(getWindowLong(hwnd, GWL_STYLE))
	return style&WS_MAXIMIZE != 0
}
func IsWindowMinimised(hwnd uintptr) bool {
	style := uint32(getWindowLong(hwnd, GWL_STYLE))
	return style&WS_MINIMIZE != 0
}

func RestoreWindow(hwnd uintptr) {
	showWindow(hwnd, SW_RESTORE)
}

func ShowWindowMaximised(hwnd uintptr) {
	showWindow(hwnd, SW_MAXIMIZE)
}
func ShowWindowMinimised(hwnd uintptr) {
	showWindow(hwnd, SW_MINIMIZE)
}

func SetBackgroundColour(hwnd uintptr, r, g, b uint8) {
	col := uint32(r) | uint32(g)<<8 | uint32(b)<<16
	hbrush, _, _ := procCreateSolidBrush.Call(uintptr(col))
	setClassLongPtr(hwnd, GCLP_HBRBACKGROUND, hbrush)
}

func IsWindowNormal(hwnd uintptr) bool {
	return !IsWindowMaximised(hwnd) && !IsWindowMinimised(hwnd) && !IsWindowFullScreen(hwnd)
}

func setClassLongPtr(hwnd uintptr, param int32, val uintptr) bool {
	proc := procSetClassLongPtr
	if strconv.IntSize == 32 {
		/*
			https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-setclasslongptrw
			Note: 	To write code that is compatible with both 32-bit and 64-bit Windows, use SetClassLongPtr.
					When compiling for 32-bit Windows, SetClassLongPtr is defined as a call to the SetClassLong function

			=> We have to do this dynamically when directly calling the DLL procedures
		*/
		proc = procSetClassLong
	}

	ret, _, _ := proc.Call(
		hwnd,
		uintptr(param),
		val,
	)
	return ret != 0
}

func getWindowLong(hwnd uintptr, index int) int32 {
	ret, _, _ := procGetWindowLong.Call(
		hwnd,
		uintptr(index))

	return int32(ret)
}

func showWindow(hwnd uintptr, cmdshow int) bool {
	ret, _, _ := procShowWindow.Call(
		hwnd,
		uintptr(cmdshow))
	return ret != 0
}

func MustStringToUTF16Ptr(input string) *uint16 {
	ret, err := syscall.UTF16PtrFromString(input)
	if err != nil {
		panic(err)
	}
	return ret
}

func MustStringToUTF16uintptr(input string) uintptr {
	ret, err := syscall.UTF16PtrFromString(input)
	if err != nil {
		panic(err)
	}
	return uintptr(unsafe.Pointer(ret))
}

func MustUTF16FromString(input string) []uint16 {
	ret, err := syscall.UTF16FromString(input)
	if err != nil {
		panic(err)
	}
	return ret
}