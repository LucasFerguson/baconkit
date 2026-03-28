package tools

import (
	"baconkit/util"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func GetSyscalls(pid int) map[string]string {
	// TODO: Allow support for syscall@plt
	exefile := getExe(pid)
	if exefile == "" {
		return nil
	}
	// obj, obj_err, err := util.Cmd("objdump", "-d", exefile)

	// Perform objdump -d exefile | grep -B 5 syscall
	objCmd := exec.Command("objdump", "-d", exefile)
	grepCmd := exec.Command("grep", "-B", "5", "-E", "syscall$")
	pipeReader, pipeWriter := io.Pipe()
	objCmd.Stdout = pipeWriter
	grepCmd.Stdin = pipeReader

	var cmdOut bytes.Buffer
	grepCmd.Stdout = &cmdOut
	var obj_err bytes.Buffer
	objCmd.Stderr = &obj_err
	var grepErr bytes.Buffer
	grepCmd.Stderr = &grepErr

	err := objCmd.Start()
	if util.ErrHandle(err) {
		fmt.Println("yeet")
		return nil
	}
	err = grepCmd.Start()
	if util.ErrHandle(err) {
		fmt.Println("yeet1")
		return nil
	}
	err = objCmd.Wait()
	if util.ErrHandle(err) {
		fmt.Println("yeet2")
		return nil
	}
	pipeWriter.Close()

	err = grepCmd.Wait()
	if err != nil {
		if err.Error() == "exit status 1" {
			return map[string]string{"Syscalls": "None"}
		} else {
			util.ErrHandle(err)
			return nil
		}
	}

	obj := strings.TrimSpace(cmdOut.String())
	if obj == "" {
		fmt.Println(obj_err)
		return nil
	}
	// Match syscall or anything involving the rax (syscall no) register
	re := regexp.MustCompile(`syscall|\$0x\w+,(%eax|%rax|%ax|%al)`)
	matches := re.FindAllString(obj, -1)
	// fmt.Println(matches)

	// Take found syscalls and rax loading and pair them!
	syscallNoMap := make(map[int]bool)

	for idx, match := range matches {
		// Check that the match isn't a syscall, that it isn't the last match, and that the next match is a syscall
		if match != "syscall" && idx != len(matches)-1 && matches[idx+1] == "syscall" {
			// Current format: $0x<hex>,%<reg>
			split := strings.Split(match, ",")
			if len(split) != 2 {
				return nil
			}
			syscallNo, err := strconv.ParseInt(split[0][3:], 16, 64)
			if util.ErrHandle(err) {
				return nil
			}
			syscallNoMap[int(syscallNo)] = true
		}
	}
	syscallNos := make([]string, 0, len(syscallNoMap))
	for no := range syscallNoMap {
		syscallNos = append(syscallNos, strconv.Itoa(no))
	}

	return map[string]string{"Syscalls": strings.Join(syscallNos, ", ")}
}

func GetLibcCalls(pid int) map[string]string {
	return nil
}
