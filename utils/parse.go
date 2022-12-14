package utils

import "strings"

func ParseAPI(s string) (addr, token string) {
	if s == "" {
		return
	}
	split := strings.Split(s, ":")
	if len(split) == 1 {
		addr = split[0]
		return
	}
	if len(split) == 2 {
		token = split[0]
		addr = split[1]
		return
	}
	return
}
