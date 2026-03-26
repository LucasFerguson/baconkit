package tools

import "strings"

func CheckWorldWriteableDir(pid int) map[string]string {
	cwdfile := getCwd(pid)
	if cwdfile == "" {
		return nil
	}
	for _, baddir := range []string{"/tmp", "/dev/shm", "/run/shm"} {
		if strings.Contains(cwdfile, baddir) {
			return map[string]string{"World Writeable Dir": baddir}
		}
	}
	return map[string]string{"World Writeable Dir": "No"}
}
