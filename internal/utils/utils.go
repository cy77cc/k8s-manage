package utils

import (
	"fmt"
	"time"
)

func SlicesToString[T any](s []T, sep string) string {
	res := ""
	for i, v := range s {
		res += fmt.Sprintf("%v", v)
		if i != len(s) - 1 {
			res += sep
			res += " "
		}
	}
	return res
}

func GetTimestamp() int64 {
	return time.Now().Unix()
}
