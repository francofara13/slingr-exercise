package utils

import (
	"fmt"
	"io"
	"strings"
)

func strip(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		b := s[i]
		if ('a' <= b && b <= 'z') ||
			('A' <= b && b <= 'Z') ||
			('0' <= b && b <= '9') ||
			b == ' ' {
			result.WriteByte(b)
		}
	}
	return result.String()
}

func GetStringResult(reader io.Reader) string {
	buf := new(strings.Builder)
	_, _ = io.Copy(buf, reader)
	result := strip(buf.String())
	fmt.Print(result + "\n")

	return result
}
