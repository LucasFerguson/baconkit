package main

import (
	"fmt"
)

func errhandle(e error) bool {
	if e != nil {
		fmt.Println("ERR", e)
		// os.Exit(0)
		return true
	}
	return false
}

func psbrute_scan() map[int]map[string]string {
	// Call processes.go's loadProcesses to get listed /proc processes
	proc_files := loadProcesses()

}
