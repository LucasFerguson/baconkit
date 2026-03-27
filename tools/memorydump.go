package tools

import (
	"baconkit/util"
	"bufio"
	"bytes"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type MemScanner func([]byte)

// This algorithm was adjusted from the Python script posted at https://stackoverflow.com/a/23001686
func scanMemoryDump(pid int, scanner MemScanner) error {
	pid_str := strconv.Itoa(pid)
	memFile, err := os.Open("/proc/" + pid_str + "/mem")
	if err != nil {
		return err
	}
	defer memFile.Close()
	mapBytes, err := os.ReadFile("/proc/" + pid_str + "/maps")
	if err != nil {
		return err
	}
	mapLines := util.TrimSplitLines(string(mapBytes))
	// memoryDump := make([]byte, 0, 1000)
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
		if err != nil {
			return err
		}
		if start > 0xFFFFFFFFFFFF {
			continue
		}
		end, err := strconv.ParseInt(endHex, 16, 64)
		if err != nil {
			return err
		}
		memFile.Seek(start, 0)
		reader := bufio.NewReader(io.LimitReader(memFile, end-start))
		err = nil
		var byteChunk []byte
		for err == nil {
			byteChunk, err = reader.ReadBytes(byte(0))
			byteChunk = bytes.Trim(byteChunk, "\x00")
			scanner(byteChunk)
		}
		if err != io.EOF {
			return err
		}
	}
	return nil
}

func MemoryDumpIP(pid int) map[string]string {
	ips := make([]string, 0, 3)
	re := regexp.MustCompile(`^((25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)\.){3}(25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)$`)
	err := scanMemoryDump(pid, func(byteChunk []byte) {
		strChunk := string(byteChunk)
		matches := re.FindStringSubmatch(strChunk)
		if len(matches) > 0 {
			ips = append(ips, matches[0])
		}
	})
	if util.ErrHandle(err) {
		return nil
	}
	if len(ips) == 0 {
		return map[string]string{"Mem IPs": "None"}
	} else {
		return map[string]string{"Mem IPs": strings.Join(ips, ", ")}
	}
}
