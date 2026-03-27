package tools

import (
	"baconkit/util"
	"bufio"
	"bytes"
	"fmt"
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
	// re := regexp.MustCompile("(?<Start>[0-9A-Fa-f]+)-(?<End>[0-9A-Fa-f]+) (?<Perm>[-r])")
	re := regexp.MustCompile("(?<Start>[0-9A-Fa-f]+)-(?<End>[0-9A-Fa-f]+) (?<Perm>[-r][-w])")
	for _, line := range mapLines {
		match := re.FindStringSubmatch(line)
		perm := match[re.SubexpIndex("Perm")]
		// if perm != "r" || strings.HasSuffix(line, "[vvar]") || strings.HasSuffix(line, "[vdso]") {
		if perm != "rw" || strings.HasSuffix(line, "[vvar]") || strings.HasSuffix(line, "[vdso]") {
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
		_, err = memFile.Seek(start, 0)
		if util.ErrHandle(err) {
			fmt.Println("Here1")
			return nil
		}
		reader := bufio.NewReader(io.LimitReader(memFile, end-start))
		err = nil
		var byteChunk []byte
		counter := 0
		for err == nil {
			// fmt.Println("success")
			byteChunk, err = reader.ReadBytes(byte(0))
			counter += len(byteChunk)
			byteChunk = bytes.Trim(byteChunk, "\x00")
			scanner(byteChunk)
		}
		if err != io.EOF {
			fmt.Println("Here2", end-start, counter, perm, line)
			return err
		}
	}
	return nil
}

func AcceptIPMatch(chunk string, matchStartIdx int, matchEndIdx int) bool {
	// RULES: (assume no digits immediately after match)
	// Banned prefixes: digit followed by period
	// prefixBlacklstRe := regexp.MustCompile(`\d\.$`)
	prefixBlacklstRe := `\d\.$`
	// Banned suffixes: period followed by digit
	// suffixBlacklstRe := regexp.MustCompile(`^\.\d`)
	suffixBlacklstRe := `^\.\d`

	// Check beginning of IP
	prefix := chunk[:matchStartIdx]
	// fmt.Println("Prefix:", prefix)
	match, err := regexp.MatchString(prefixBlacklstRe, prefix)
	if util.ErrHandle(err) {
		return false
	}
	if match {
		return false
	}

	// Check end of IP
	suffix := chunk[matchEndIdx:]
	// fmt.Println("Suffix:", suffix)
	match, err = regexp.MatchString(suffixBlacklstRe, suffix)
	if util.ErrHandle(err) {
		return false
	}
	return !match
}

func MemoryDumpIP(pid int) map[string]string {
	ips := make([]string, 0, 3)
	re := regexp.MustCompile(`^((25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)\.){3}(25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)$`) // strict full line ip regex
	// re := regexp.MustCompile(`(?:^|[^\d])(((25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)\.){3}(25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d))(?:$|[^\d])`) // ip with no number before or after
	err := scanMemoryDump(pid, func(byteChunk []byte) {
		strChunk := strings.TrimSpace(string(byteChunk))
		matches := re.FindStringSubmatch(strChunk) // the beginning and end groups that block digits count as groups and so must be ignored
		if matches != nil {
			match := matches[0] // strict full line version
			// match := matches[1] // index 0 has digit-ignorer-prefix group match
			// fmt.Println("Match:", match)
			idxs := re.FindStringSubmatchIndex(strChunk)[:2] // strict full line version
			// idxs := re.FindStringSubmatchIndex(strChunk)[2:4] // idx 0,1 contain the beginning and end of the digit-ignorer-prefix group match

			if AcceptIPMatch(strChunk, idxs[0], idxs[1]) {
				// fmt.Println("Accepted ip")
				// fmt.Println(strChunk)
				ips = append(ips, match)
			}
		}
	})
	if util.ErrHandle(err) {
		return nil
	}
	if len(ips) == 0 {
		return map[string]string{"Mem IPs": "None"}
	} else {
		// Get unique list of IPs
		ipMap := make(map[string]bool)
		for _, ip := range ips {
			ipMap[ip] = true
		}
		uniqueIps := make([]string, 0, len(ipMap))
		for ip := range ipMap {
			uniqueIps = append(uniqueIps, ip)
		}
		return map[string]string{"Mem IPs": strings.Join(uniqueIps, ", ")}
	}
}
