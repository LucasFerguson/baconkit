package util

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func ErrHandle(e error) bool {
	if e != nil {
		fmt.Println("ERR", e)
		// os.Exit(0)
		return true
	}
	return false
}

func TrimSplit(str string, delim string) []string {
	return strings.Split(strings.TrimSpace(str), delim)
}

func TrimSplitLines(str string) []string {
	return TrimSplit(str, "\n")
}

func Cmd(command string, args ...string) (string, string, error) {
	cmd := exec.Command(command, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}
