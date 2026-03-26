package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func errhandle(e error) bool {
	if e != nil {
		fmt.Println("ERR", e)
		// os.Exit(0)
		return true
	}
	return false
}

func cmd(command string, args ...string) (string, string, error) {
	cmd := exec.Command(command, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func checkpid(pid int) {
	exelink = "/proc/" + pid + "/exe"
	exefile = filepath.EvalSymlinks(exelink)
	dpkg_out, dpkg_err, err = cmd("dpkg", "-S", exefile)
	if errhandle(err) {
		return false
	}
	dpkg_out, dpkg_err = strings.TrimSpace(dpkg_out), strings.TrimSpace(dpkg_err)
	if dpkg_out == "" || strings.Contains(dpkg_err, "no path found matching pattern") {
		return true
	} else {
		return false
	}
}
