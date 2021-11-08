//go:build windows
// +build windows

package executil

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func MakeCmd(args ...string) *exec.Cmd {
	return exec.Command("cmd", append([]string{"/C"}, args...)...)
}

func CheckFile(filename string) error {
	if _, err := os.Stat(filename); err != nil {
		if errors.Is(err, syscall.ERROR_FILE_NOT_FOUND) {
			return fmt.Errorf("файл %q не найден", filename)
		}
		return err
	}
	return nil
}

func CheckFileFs(fsystem fs.FS, filename string) error {
	if _, err := fs.Stat(fsystem, filename); err != nil {
		return err
	}
	return nil
}

func Executable(file fs.DirEntry) bool {
	return strings.HasSuffix(file.Name(), ".exe")
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
