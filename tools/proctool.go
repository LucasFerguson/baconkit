package tools

import (
	"baconkit/util"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func getExe(pid int) string {
	exelink := "/proc/" + strconv.Itoa(pid) + "/exe"
	exefile, err := filepath.EvalSymlinks(exelink)
	if util.ErrHandle(err) {
		return ""
	}
	return exefile
}

func getCwd(pid int) string {
	cwdlink := "/proc/" + strconv.Itoa(pid) + "/cwd"
	cwdfile, err := filepath.EvalSymlinks(cwdlink)
	if util.ErrHandle(err) {
		return ""
	}
	return cwdfile
}

func getStatusMap(pid int) map[string]string {
	statusBytes, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "status"))
	if util.ErrHandle(err) {
		return nil
	}
	statusLines := util.TrimSplitLines(string(statusBytes))
	statusMap := make(map[string]string)
	for _, line := range statusLines {
		parts := strings.SplitN(line, ":", 2)
		statusMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return statusMap
}

func GetBasicStatus(pid int) map[string]string {
	users := loadUsers()
	statusMap := getStatusMap(pid)
	if statusMap == nil {
		return nil
	}
	var user string
	if u, ok := users[strings.Fields(statusMap["Uid"])[0]]; ok {
		user = u
	}

	return map[string]string{
		"Name":  statusMap["Name"],
		"User":  user,
		"State": statusMap["State"],
	}
}

func loadUsers() map[string]string {
	users := make(map[string]string)
	content, err := os.ReadFile("/etc/passwd")
	if err != nil {
		return users
	}
	for _, line := range strings.Split(string(content), "\n") {
		parts := strings.Split(line, ":")
		if len(parts) >= 3 {
			users[parts[2]] = parts[0]
		}
	}
	return users
}

func IsPtraced(pid int) map[string]string {
	statusMap := getStatusMap(pid)
	if statusMap == nil {
		return nil
	}
	if statusMap["TracerPid"] == "0" {
		return map[string]string{"PTracer": "None"}
	} else {
		tracerPid, err := strconv.Atoi(statusMap["TracerPid"])
		if util.ErrHandle(err) {
			return map[string]string{"PTracer": statusMap["TracerPid"]}
		}
		tracerStatusMap := getStatusMap(tracerPid)
		if tracerStatusMap == nil {
			return map[string]string{"PTracer": statusMap["TracerPid"]}
		}
		return map[string]string{"PTracer": tracerStatusMap["Name"]}
	}
}
