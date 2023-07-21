package utils

import "strings"

func ParseAPI(s string) (addr, token string) {
	// expect token:addr or addr
	if s == "" {
		return
	}

	if strings.HasPrefix(s, "http") || strings.HasPrefix(s, "ip4") || strings.HasPrefix(s, "ip6") || strings.HasPrefix(s, "unix") {
		return s, ""
	}

	before, after, exist := strings.Cut(s, ":")
	if !exist {
		return s, ""
	} else {
		return after, before
	}
}
