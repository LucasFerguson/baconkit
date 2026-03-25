package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
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

func ps_scan() map[int]map[string]string {
	// Run ps
	ps_out, ps_err, err := cmd("ps", "-eo", "user,pid,exe,stat")
	if errhandle(err) {
		return nil
	}
	fmt.Println(ps_err)
	ps_out = strings.TrimSpace(ps_out)

	procmap := make(map[int]map[string]string)

	// Parse ps
	processes := strings.Split(ps_out, "\n")[1:]
	for _, proc := range processes {
		attrs := strings.Fields(proc)
		pid, err := strconv.Atoi(attrs[1])
		if errhandle(err) {
			continue
		}
		procmap[pid] = map[string]string{
			"Name":  attrs[2],
			"User":  attrs[0],
			"State": attrs[3],
		}
	}
	// Return ps output
	fmt.Println(procmap)
	return procmap
}
