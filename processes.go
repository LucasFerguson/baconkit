package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/table"
)

func loadProcesses() []table.Row {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}

	users := loadUsers()

	var rows []table.Row
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, err := strconv.Atoi(entry.Name()); err != nil {
			continue
		}

		pid := entry.Name()
		content, err := os.ReadFile(filepath.Join("/proc", pid, "status"))
		if err != nil {
			continue
		}

		name, uid, state := parseStatus(string(content))
		user := uid
		if u, ok := users[uid]; ok {
			user = u
		}

		rows = append(rows, table.Row{pid, name, user, state})
	}
	return rows
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
