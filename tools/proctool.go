package tools

import (
	"baconkit/util"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func GetBasicStatus(pid int) map[string]string {
	statusBytes, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "status"))
	if util.ErrHandle(err) {
		return nil
	}
	users := loadUsers()
	// statusVals := util.TrimSplitLines(string(statusBytes))
	name, uid, state := parseStatus(string(statusBytes))
	user := uid
	if u, ok := users[uid]; ok {
		user = u
	}

	return map[string]string{
		"Name":  name,
		"User":  user,
		"State": state,
	}
}

func parseStatus(content string) (name, uid, state string) {
	for _, line := range strings.Split(content, "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		val := strings.TrimSpace(parts[1])
		switch strings.TrimSpace(parts[0]) {
		case "Name":
			name = val
		case "State":
			state = val
		case "Uid":
			fields := strings.Fields(val)
			if len(fields) > 0 {
				uid = fields[0]
			}
		}
	}
	return
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
