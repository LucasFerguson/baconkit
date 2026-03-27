package tools

import (
	"baconkit/util"
	"bytes"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// This algorithm was taken from the Python script posted at https://stackoverflow.com/a/23001686
func getMemoryDump(pid int) []byte {
	pid_str := strconv.Itoa(pid)
	memFile, err := os.Open("/proc/" + pid_str + "/mem")
	if util.ErrHandle(err) {
		return nil
	}
	defer memFile.Close()
	mapBytes, err := os.ReadFile("/proc/" + pid_str + "/maps")
	if util.ErrHandle(err) {
		return nil
	}
	mapLines := util.TrimSplitLines(string(mapBytes))
	memoryDump := make([]byte, 0, 1000)
	re := regexp.MustCompile("(?<Start>[0-9A-Fa-f]+)-(?<End>[0-9A-Fa-f]+) (?<Perm>[-r][-w])")
	for _, line := range mapLines {
		match := re.FindStringSubmatch(line)
		perm := match[re.SubexpIndex("Perm")]
		if perm != "rw" {
			continue
		}
		startHex := match[re.SubexpIndex("Start")]
		endHex := match[re.SubexpIndex("End")]
		start, err := strconv.ParseInt(startHex, 16, 64)
		if util.ErrHandle(err) {
			return nil
		}
		if start > 0xFFFFFFFFFFFF {
			continue
		}
		end, err := strconv.ParseInt(endHex, 16, 64)
		if util.ErrHandle(err) {
			return nil
		}
		memFile.Seek(start, 0)
		memBytes := make([]byte, end-start)
		n, err := io.ReadFull(memFile, memBytes)
		if util.ErrHandle(err) || int64(n) < end-start {
			return nil
		}
		memoryDump = append(memoryDump, byte(0))
		memoryDump = append(memoryDump, memBytes...)
	}
	return memoryDump
}

func MemoryDumpIP(pid int) map[string]string {
	// Perform memory dump
	memoryDump := getMemoryDump(pid)
	if memoryDump == nil {
		return nil
	}
	// Search memory dump for IP address
	re := regexp.MustCompile(`^((25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)\.){3}(25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)$`)
	memoryLines := bytes.Split(memoryDump, []byte{0})
	ips := make([]string, 0, 3)
	for _, line := range memoryLines {
		lineStr := string(line)
		matches := re.FindStringSubmatch(lineStr)
		if len(matches) > 0 {
			ips = append(ips, matches[0])
		}
	}

	// matches := re.FindAllStringSubmatch(memoryDump, -1)

	if len(ips) == 0 {
		return map[string]string{"Mem IPs": "None"}
	} else {
		return map[string]string{"Mem IPs": strings.Join(ips, ", ")}
	}
}
