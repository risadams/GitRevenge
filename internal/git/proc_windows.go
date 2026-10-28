package git

import (
	"os/exec"
	"syscall"
)

// createNoWindow is CREATE_NO_WINDOW from the Windows API.
const createNoWindow = 0x08000000

// hideWindow stops a console window from flashing up for every git call, since the
// app itself is a GUI process with no console to inherit.
func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
}
