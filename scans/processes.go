package scans

import (
	"os"
	"strconv"
)

// func loadProcesses() []table.Row {
// 	entries, err := os.ReadDir("/proc")
// 	if err != nil {
// 		return nil
// 	}

// 	users := loadUsers()

// 	var rows []table.Row
// 	for _, entry := range entries {
// 		if !entry.IsDir() {
// 			continue
// 		}
// 		if _, err := strconv.Atoi(entry.Name()); err != nil {
// 			continue
// 		}

// 		pid := entry.Name()
// 		content, err := os.ReadFile(filepath.Join("/proc", pid, "status"))
// 		if err != nil {
// 			continue
// 		}

// 		name, uid, state := parseStatus(string(content))
// 		user := uid
// 		if u, ok := users[uid]; ok {
// 			user = u
// 		}

// 		rows = append(rows, table.Row{pid, name, user, state})
// 	}
// 	return rows
// }

func LoadProcesses() []int {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}

	pidlst := make([]int, 0, 30)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		pidlst = append(pidlst, pid)
	}
	return pidlst
}
