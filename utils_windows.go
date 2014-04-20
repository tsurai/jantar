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

type small_rect struct {
	left   int16
	top    int16
	right  int16
	bottom int16
}

type console_screen_buffer_info struct {
	size                coord
	cursor_position     coord
	attributes          uint16
	window              small_rect
	maximum_window_size coord
}

var (
	kernel32                            = syscall.NewLazyDLL("kernel32.dll")
	proc_get_console_screen_buffer_info = kernel32.NewProc("GetConsoleScreenBufferInfo")
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
	var info console_screen_buffer_info
	rc, _, _ := syscall.Syscall(proc_get_console_screen_buffer_info.Addr(), 2, uintptr(hOut), uintptr(unsafe.Pointer(&info)), 0)
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
