package tools

import (
	"baconkit/util"
	"os"
	"strconv"
	"strings"
)

func CheckRunningDeleted(pid int) map[string]string {
	// Check if file is deleted
	linkName, err := os.Readlink("/proc/" + strconv.Itoa(pid) + "/exe")
	if util.ErrHandle(err) {
		return nil
	}
	if strings.HasSuffix(linkName, " (deleted)") {
		exeFile := strings.TrimSuffix(linkName, " (deleted)")
		// Check if file is actually deleted
		_, err := os.Stat(exeFile)
		if os.IsNotExist(err) {
			return map[string]string{"Fileless Exe": "(Deleted)"}
		} else if util.ErrHandle(err) {
			return nil
		}
	}
	// Check if file is creating using memfd
	return nil
}
