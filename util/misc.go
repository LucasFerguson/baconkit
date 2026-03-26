package util

import (
	"fmt"
	"strings"
)

func ErrHandle(e error) bool {
	if e != nil {
		fmt.Println("ERR", e)
		// os.Exit(0)
		return true
	}
	return false
}

func TrimSplit(str string, delim string) []string {
	return strings.Split(strings.TrimSpace(str), delim)
}

func TrimSplitLines(str string) []string {
	return TrimSplit(str, "\n")
}
