package tools

import (
	"baconkit/util"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func CheckService(pid int) map[string]string {
	sysout, syserr, err := util.Cmd("systemctl", "status", strconv.Itoa(pid))
	if util.ErrHandle(err) {
		return nil
	}
	sysout = strings.TrimSpace(sysout)
	if sysout == "" {
		fmt.Println(syserr)
		return nil
	}
	serviceLine := strings.TrimSpace(strings.Split(sysout, "\n")[0])
	re := regexp.MustCompile(`(?P<ServFile>\S*) - (?P<ServName>.*)`)
	matches := re.FindStringSubmatch(serviceLine)
	// Note!! Description of service can be gotten by matches[re.SubexpIndex("ServName")]
	return map[string]string{"Service": matches[re.SubexpIndex("ServFile")]}
}
