package main

import (
	"fmt"
	"log"
	"syscall"
	"unsafe"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	pGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
)

func getModuleHandle() (syscall.Handle, error) {
	ret, _, err := pGetModuleHandleW.Call(uintptr(0))
	if ret == 0 {
		return 0, err
	}
	return syscall.Handle(ret), nil
}

var (
	user32 = syscall.NewLazyDLL("user32.dll")
	//coredll = syscall.NewLazyDLL("coredll.dll") // if someone points you to this, it probably means user32 on desktop.
	gdi32 = syscall.NewLazyDLL("gdi32.dll")

	pCreateWindowExW  = user32.NewProc("CreateWindowExW")
	pDefWindowProcW   = user32.NewProc("DefWindowProcW")
	pDestroyWindow    = user32.NewProc("DestroyWindow")
	pDispatchMessageW = user32.NewProc("DispatchMessageW")
	pGetMessageW      = user32.NewProc("GetMessageW")
	pLoadCursorW      = user32.NewProc("LoadCursorW")
	pPostQuitMessage  = user32.NewProc("PostQuitMessage")
	pRegisterClassExW = user32.NewProc("RegisterClassExW")
	pTranslateMessage = user32.NewProc("TranslateMessage")

	pBeginPaint       = user32.NewProc("BeginPaint")
	pEndPaint         = user32.NewProc("EndPaint")

	pGetClientRect     = user32.NewProc("GetClientRect")
	pGetDC = user32.NewProc("GetDC")
	pCreateCompatibleDC = gdi32.NewProc("CreateCompatibleDC")
	pCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	pDeleteObject = gdi32.NewProc("DeleteObject")
)

const (
	cSW_SHOW        = 5
	cSW_USE_DEFAULT = 0x80000000
)

const (
	cWS_MAXIMIZE_BOX = 0x00010000
	cWS_MINIMIZEBOX  = 0x00020000
	cWS_THICKFRAME   = 0x00040000
	cWS_SYSMENU      = 0x00080000
	cWS_CAPTION      = 0x00C00000
	cWS_VISIBLE      = 0x10000000

	cWS_OVERLAPPEDWINDOW = 0x00CF0000
)

//func createWindow(className, windowName string, style uint32, x, y, width, height int32, parent, menu, instance syscall.Handle) (syscall.Handle, error) {
func createWindow(className, windowName string, style uint32, x, y, width, height uint32, parent, menu, instance syscall.Handle) (syscall.Handle, error) {
	classNamePtr, _ :=syscall.UTF16PtrFromString(className)
	windowNamePtr, _ :=syscall.UTF16PtrFromString(windowName)

	ret, _, err := pCreateWindowExW.Call(
		uintptr(0),
		uintptr(unsafe.Pointer(classNamePtr)),
		uintptr(unsafe.Pointer(windowNamePtr)),
		uintptr(style),
		uintptr(x),
		uintptr(y),
		uintptr(width),
		uintptr(height),
		uintptr(parent),
		uintptr(menu),
		uintptr(instance),
		uintptr(0),
	)
	if ret == 0 {
		return 0, err
	}
	return syscall.Handle(ret), nil
}

const (
	// Handy reference: https://www.pinvoke.net/default.aspx/Constants.WM
	cWM_DESTROY = 0x0002
	cWM_CLOSE   = 0x0010
	cWM_PAINT	= 0x000F
)

func defWindowProc(hwnd syscall.Handle, msg uint32, wparam, lparam uintptr) uintptr {
	ret, _, _ := pDefWindowProcW.Call(
		uintptr(hwnd),
		uintptr(msg),
		uintptr(wparam),
		uintptr(lparam),
	)
	return uintptr(ret)
}

func destroyWindow(hwnd syscall.Handle) error {
	ret, _, err := pDestroyWindow.Call(uintptr(hwnd))
	if ret == 0 {
		return err
	}
	return nil
}

type tPOINT struct {
	x, y int32
}

type tMSG struct {
	hwnd    syscall.Handle
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      tPOINT
}

func dispatchMessage(msg *tMSG) {
	pDispatchMessageW.Call(uintptr(unsafe.Pointer(msg)))
}

func getMessage(msg *tMSG, hwnd syscall.Handle, msgFilterMin, msgFilterMax uint32) (bool, error) {
	ret, _, err := pGetMessageW.Call(
		uintptr(unsafe.Pointer(msg)),
		uintptr(hwnd),
		uintptr(msgFilterMin),
		uintptr(msgFilterMax),
	)
	if int32(ret) == -1 {
		return false, err
	}
	return int32(ret) != 0, nil
}

const (
	cIDC_ARROW = 32512
)

func loadCursorResource(cursorName uint32) (syscall.Handle, error) {
	ret, _, err := pLoadCursorW.Call(
		uintptr(0),
		uintptr(uint16(cursorName)),
	)
	if ret == 0 {
		return 0, err
	}
	return syscall.Handle(ret), nil
}

func postQuitMessage(exitCode int32) {
	pPostQuitMessage.Call(uintptr(exitCode))
}

const (
	cCOLOR_WINDOW = 5
)

type tWNDCLASSEXW struct {
	size       uint32
	style      uint32
	wndProc    uintptr
	clsExtra   int32
	wndExtra   int32
	instance   syscall.Handle
	icon       syscall.Handle
	cursor     syscall.Handle
	background syscall.Handle
	menuName   *uint16
	className  *uint16
	iconSm     syscall.Handle
}

type tRECT struct{
	left,top,right,bottom int32
}

type tPAINTSTRUCT struct {
	// To decode the Win32 type macros, see -- https://docs.microsoft.com/en-us/windows/win32/learnwin32/windows-coding-conventions
	hdc        uintptr // draw context: used to do actual output
	fErase     bool
	rcPaint    tRECT
	fRestore   bool
	fIncUpdate bool
	reserved   [32]byte
}

type tBITMAPINFO struct {
	bmiHeader tBITMAPINFOHEADER
	bmiColors [1]int32;
}

type tBITMAPINFOHEADER struct {
	biSize int32
	biWidth int32
	biHeight int32
	biPlanes int16
	biBitCount int16
	biCompression uint32//BitmapCompressionMode - BI_RGB = 0,BI_RLE8 = 1,BI_RLE4 = 2,BI_BITFIELDS = 3,BI_JPEG = 4,BI_PNG = 5
	biSizeImage int32
	biXPelsperMeter int32
	biYPelsPerMeter int32
	biClrUsed int32
	biClrImportant int32
}

func registerClassEx(wcx *tWNDCLASSEXW) (uint16, error) {
	ret, _, err := pRegisterClassExW.Call(
		uintptr(unsafe.Pointer(wcx)),
	)
	if ret == 0 {
		return 0, err
	}
	return uint16(ret), nil
}

func translateMessage(msg *tMSG) {
	pTranslateMessage.Call(uintptr(unsafe.Pointer(msg)))
}

func beginPaint(hwnd syscall.Handle, paintStruct *tPAINTSTRUCT){
	// Win32 calls always return an error >:-(
	_, _, _ = pBeginPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(paintStruct)))
}

func endPaint(hwnd syscall.Handle, paintStruct *tPAINTSTRUCT){
	_, _, _ = pEndPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(paintStruct)))
}

func getClientRect(hwnd syscall.Handle, lpRect *tRECT){
	_, _, _ = pGetClientRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(lpRect)))
}

func getDC(hwnd syscall.Handle) uintptr{
	r1, _, _ := pGetDC.Call(uintptr(hwnd))
	return r1
}

func createCompatibleDC(hdc uintptr) uintptr {
	r1, _, _ := pCreateCompatibleDC.Call(hdc)
	return r1
}

func createCompatibleBitmap(hdc uintptr, width int32, height int32) uintptr {
	r1, _, _ := pCreateCompatibleBitmap.Call(hdc, uintptr(width), uintptr(height))
	return r1
}

func deleteObject(obj uintptr){
	_, _, _ = pDeleteObject.Call(obj)
}

func main() {
	className := "testClass"

	instance, err := getModuleHandle()
	if err != nil {
		log.Println(err)
		return
	}

	cursor, err := loadCursorResource(cIDC_ARROW)
	if err != nil {
		log.Println(err)
		return
	}

	// --------------------------------------------
	// MAIN WIN32 EVENT LOOP
	// --------------------------------------------
	fn := func(hwnd syscall.Handle, msg uint32, wparam, lparam uintptr) uintptr {
		switch msg {
		case cWM_CLOSE:
			destroyWindow(hwnd)
		case cWM_DESTROY:
			postQuitMessage(0)
		case cWM_PAINT:
			paint:=tPAINTSTRUCT{}
			beginPaint(hwnd, &paint)
			defer endPaint(hwnd, &paint)

			DrawBitsIntoWindow(hwnd)

			//fmt.Printf("drawing %v (%v)\r\n", hwnd, paint.hdc)
			// TODO: figure out drawing a raw bitmap here
			// see also https://www.codeproject.com/articles/224754/guide-to-win32-memory-dc
			// HBITMAP memBM = CreateCompatibleBitmap ( hDC, nWidth, nHeight );  https://docs.microsoft.com/en-us/windows/win32/api/wingdi/nf-wingdi-createcompatiblebitmap
			// int SetDIBits(hDC, memBM, ...); https://docs.microsoft.com/en-us/windows/win32/api/wingdi/nf-wingdi-setdibits
			// BOOL DeleteObject(memBM); https://docs.microsoft.com/en-us/windows/win32/api/wingdi/nf-wingdi-deleteobject
		default:
			ret := defWindowProc(hwnd, msg, wparam, lparam)
			return ret
		}
		return 0
	}

	wcx := tWNDCLASSEXW{
		wndProc:    syscall.NewCallback(fn),
		instance:   instance,
		cursor:     cursor,
		background: cCOLOR_WINDOW + 1,
		className:  syscall.StringToUTF16Ptr(className),
	}
	wcx.size = uint32(unsafe.Sizeof(wcx))

	if _, err = registerClassEx(&wcx); err != nil {
		log.Println(err)
		return
	}

	_, err = createWindow(
		className,
		"Hello window",
		cWS_VISIBLE|cWS_OVERLAPPEDWINDOW,
		cSW_USE_DEFAULT,
		cSW_USE_DEFAULT,
		cSW_USE_DEFAULT,
		cSW_USE_DEFAULT,
		0,
		0,
		instance,
	)
	if err != nil {
		log.Println(err)
		return
	}

	for {
		msg := tMSG{}
		gotMessage, err := getMessage(&msg, 0, 0, 0)
		if err != nil {
			log.Println(err)
			return
		}

		if gotMessage {
			translateMessage(&msg)
			dispatchMessage(&msg)
		} else {
			break
		}
	}
}

func DrawBitsIntoWindow(hwnd syscall.Handle) {
	// https://stackoverflow.com/questions/35762636/raw-direct-acess-on-pixels-data-in-a-bitmapinfo-hbitmap
	rc:= tRECT{}
	getClientRect(hwnd, &rc)

	width := rc.right
	height := rc.bottom
	if width < 1 || height < 1 {
		fmt.Println("error: invalid rectangle size")
		return
	}

	hdc := getDC(hwnd)
	//hCaptureDC := createCompatibleDC(hdc)
	hBitmap := createCompatibleBitmap(hdc, width, height)
	defer deleteObject(hBitmap)
	//fmt.Println(hBitmap, hCaptureDC)

	myBMInfo := tBITMAPINFO{}
	myBMInfo.bmiHeader = tBITMAPINFOHEADER{
		biWidth : width,
		biHeight : height,
		biPlanes : 1,
		biBitCount : 32,
	}
	myBMInfo.bmiHeader.biSize = int32(unsafe.Sizeof(myBMInfo));

	size := ((width * int32(myBMInfo.bmiHeader.biBitCount) + 31) / 32) * 4 * height
	rawbytes := make([]byte, size)//[size]byte
	fmt.Println(len(rawbytes))
}
