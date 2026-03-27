package tools

import "strings"

func checkFile(file string) map[string]string {
	if file == "" {
		return nil
	}
	for _, baddir := range []string{"/tmp", "/dev/shm", "/run/shm"} {
		if strings.Contains(file, baddir) {
			return map[string]string{"World Writeable Dir": baddir}
		}
	}
	return nil
}

func CheckWorldWriteableDir(pid int) map[string]string {
	cwdCheck := checkFile(getCwd(pid))
	if cwdCheck != nil {
		return cwdCheck
	}
	exeCheck := checkFile(getExe(pid))
	if exeCheck != nil {
		return exeCheck
	}
	return map[string]string{"World Writeable Dir": "No"}
}
