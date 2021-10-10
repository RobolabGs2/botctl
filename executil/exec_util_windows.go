// +build windows

package executil

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

func MakeCmd(args ...string) *exec.Cmd {
	return exec.Command("cmd", append([]string{"/C"}, args...)...)
}

func CheckFile(filename string) error {
	if _, err := os.Stat(filename); err != nil {
		return err
	}
	return nil
}

func KillProcess(cmd *exec.Cmd) error {
	return exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid)).Run()
}

func OpenInBrowser(path string) error {
	logs, err := MakeCmd("start", path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w, logs: %q", err, logs)
	}
	return nil
}
