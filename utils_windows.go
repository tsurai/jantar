package jantar

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

type coord struct {
	x int16
	y int16
}

type smallRect struct {
	left   int16
	top    int16
	right  int16
	bottom int16
}

type consoleScreenBufferInfo struct {
	size 								coord
	cursorPosition     	coord
	attributes         	uint16
	window              smallRect
	maximumWindowSize 	coord
}

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
)

// NOTE: assuming that all non standard terminals on windows support ansi escape codes
func isTerminal(fd uintptr) bool {
	term := os.Getenv("TERM")
	if term == "" || term == "cygwin" {
		return false
	}
	return true
}

func getTerminalSize() (int16, int16) {
	hOut, err1 := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	if err1 != nil {
		return 0, 0
	}

	// try windows api call
	var info consoleScreenBufferInfo
	rc, _, _ := syscall.Syscall(procGetConsoleScreenBufferInfo.Addr(), 2, uintptr(hOut), uintptr(unsafe.Pointer(&info)), 0)
	if int(rc) != 0 {
		return info.window.right - info.window.left + 1, info.window.bottom - info.window.top + 1
	}

	// try stty command
	var width int16
	var height int16

	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()

	if err == nil {
		if n, err := fmt.Sscanf(string(out), "%d %d", &height, &width); err == nil && n == 2 {
			return width, height
		}
	}

	return 0, 0
}
