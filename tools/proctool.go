package tools

import (
	"baconkit/util"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
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

func GetParentProcess(pid int) map[string]string {
	const PARENT_KEY = "Parent"
	statusMap := getStatusMap(pid)
	ppid, ppid_exists := statusMap["PPid"]
	if !ppid_exists {
		return map[string]string{PARENT_KEY: "None"}
	}
	ppid_int, err := strconv.Atoi(ppid)
	if util.ErrHandle(err) {
		return map[string]string{PARENT_KEY: "[" + ppid + ":]"}
	}
	return map[string]string{PARENT_KEY: "[" + ppid + ":" + getExe(ppid_int) + "]"}
}

func GetNetwork(pid int) map[string]string {
	netBytes, err := os.ReadFile(filepath.Join("/proc/net/tcp"))
	if util.ErrHandle(err) {
		return nil
	}
	nets := util.TrimSplitLines(string(netBytes))[1:]
	net_inode_map := make(map[int]string)
	// Find all TCP IPv4 sockets
	for _, net_entry := range nets {
		cols := strings.Fields(net_entry)
		local_addr := cols[1]
		// remote_addr := cols[2]
		inode := cols[9]
		localsplit := strings.Split(local_addr, ":")
		ip_hex := localsplit[0]
		port_hex := localsplit[1]
		ip_bytes_le, err := hex.DecodeString(ip_hex)
		if util.ErrHandle(err) {
			continue
		}

		ip_uintdigits := binary.LittleEndian.Uint32(ip_bytes_le)
		ip_bytes_be := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip_bytes_be, ip_uintdigits)
		ip := ip_bytes_be.String()

		port, err := strconv.ParseInt(port_hex, 16, 64)
		// fmt.Println(ip, port, inode)
		inode_int, err := strconv.Atoi(inode)
		if util.ErrHandle(err) {
			return nil
		}
		net_inode_map[inode_int] = ip + ":" + strconv.Itoa(int(port))
	}

	// List PID file descriptor to see if it contains a fd linked to socket:[inode]
	FD_DIR := "/proc/" + strconv.Itoa(pid) + "/fd"
	fds, err := os.ReadDir(FD_DIR)
	if err != nil {
		return nil
	}
	matching_conns := make([]string, 0, 1)
	for _, fd := range fds {
		if fd.IsDir() {
			continue
		}
		fd_link, err := os.Readlink(FD_DIR + "/" + fd.Name())
		if util.ErrHandle(err) {
			return nil
		}
		// fmt.Println(pid, fd.Name(), fd_link)
		if strings.HasPrefix(fd_link, "socket:[") && strings.HasSuffix(fd_link, "]") {
			var socket_inode int
			_, err = fmt.Sscanf(fd_link, "socket:[%d]", &socket_inode)
			if util.ErrHandle(err) {
				return nil
			}
			ip_port, exists := net_inode_map[socket_inode]
			if exists {
				// fmt.Println("Found at pid", pid, "->", ip_port)
				matching_conns = append(matching_conns, ip_port)
			}
		}
	}
	return map[string]string{"Network": strings.Join(matching_conns, ", ")}
}
