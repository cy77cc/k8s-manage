package utils

import (
	"fmt"
	"strings"
	"time"
)

func SlicesToString[T any](s []T, sep string) string {
	if len(s) == 0 {
		return ""
	}
	var b strings.Builder
	for i, v := range s {
		if i > 0 {
			b.WriteString(sep)
			b.WriteString(" ")
		}
		fmt.Fprintf(&b, "%v", v)
	}
	return b.String()
}

func GetTimestamp() int64 {
	return time.Now().Unix()
}
