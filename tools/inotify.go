package tools

import (
	"baconkit/util"
	"fmt"
	"regexp"
	"strings"
)

func CheckInotify(pid int) map[string]string {
	exefile := getExe(pid)
	if exefile == "" {
		return nil
	}
	nm_out, nm_err, err := util.Cmd("nm", "-D", exefile)
	if util.ErrHandle(err) {
		return nil
	}
	nm := strings.TrimSpace(nm_out)
	if nm == "" {
		if strings.TrimSpace(nm_err) == "nm: "+exefile+": no symbols" {
			return map[string]string{"Inotifier": "Stripped"}
		}
		fmt.Println(nm_err)
		return nil
	}

	// re := regexp.MustCompile("(call|jmp) .*<(?P<Func>.*(inotify|fanotify).*)@plt")
	re := regexp.MustCompile(`(?P<Func>\S*(inotify|fanotify)\S*)`)
	matches := re.FindAllStringSubmatch(nm, -1)
	funcs := make([]string, 0, len(matches))
	for _, match := range matches {
		funcs = append(funcs, strings.TrimSpace(match[re.SubexpIndex("Func")]))
	}
	if len(funcs) == 0 {
		return map[string]string{"Inotifier": "No"}
		// return nil
	} else {
		return map[string]string{"Inotifier": funcs[0]}
		// return map[string]string{"Inotifier": strings.Join(funcs, " ")}
	}
}
