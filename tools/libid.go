package tools

import (
	"baconkit/util"
	"strings"
)

func DpkgPid(pid int) map[string]string {
	exefile := getExe(pid)
	if exefile == "" {
		return nil
	}
	dpkg_out, dpkg_err, err := util.Cmd("dpkg", "-S", exefile)
	if util.ErrHandle(err) {
		return nil
	}
	dpkg_out, dpkg_err = strings.TrimSpace(dpkg_out), strings.TrimSpace(dpkg_err)
	if dpkg_out == "" || strings.Contains(dpkg_err, "no path found matching pattern") {
		return map[string]string{"Package": "None"}
	} else {
		outsplit := strings.Split(dpkg_out, ":")
		if len(outsplit) <= 2 {
			return nil
		}
		return map[string]string{"Package": outsplit[0]}
	}
}

func RpmPid(pid int) map[string]string {
	exefile := getExe(pid)
	if exefile == "" {
		return nil
	}
	rpm_out, _, err := util.Cmd("rpm", "-qf", exefile)
	if util.ErrHandle(err) {
		return nil
	}
	rpm_out = strings.TrimSpace(rpm_out)
	if rpm_out == "" {
		return nil
	} else if strings.Contains(rpm_out, "is not owned by any package") {
		return map[string]string{"Package": "None"}
	} else {
		return map[string]string{"Package": rpm_out}
	}
}
