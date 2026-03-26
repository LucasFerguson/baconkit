package scans

import (
	"fmt"
	"os"
	"slices"
	"strconv"
)

func errhandle(e error) bool {
	if e != nil {
		fmt.Println("ERR", e)
		// os.Exit(0)
		return true
	}
	return false
}

func PsBruteScan() []int {
	// Call processes.go's loadProcesses to get listed /proc processes
	proc_files := LoadProcesses()

	// Get max pid
	maxpid := 1
	for pid := range proc_files {
		if pid > maxpid {
			maxpid = pid
		}
	}

	pidlst := make([]int, 0, len(proc_files))

	for pid := 1; pid < maxpid; pid++ {
		// if _, exists := proc_files[pid]; exists {
		if slices.Contains(proc_files, pid) {
			continue
		}
		statfile := "/proc/" + strconv.Itoa(pid) + "/stat"
		// Try reading stat file
		_, err := os.ReadFile(statfile)
		if err == nil {
			// If works, append PID to PID list
			pidlst = append(pidlst, pid)
		}
	}
	return pidlst
}
