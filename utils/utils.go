package utils

import (
	"fmt"
	"io"
	"strings"
)

func Strip(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		b := s[i]
		if ('a' <= b && b <= 'z') ||
			('A' <= b && b <= 'Z') ||
			b == ' ' {
			result.WriteByte(b)
		}
	}
	return result.String()
}

func StripAlphanumerical(s string) string {
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

func GetStringResult(reader io.Reader, alphanumerical bool) string {
	buf := new(strings.Builder)
	_, _ = io.Copy(buf, reader)
	var result string
	if alphanumerical {
		result = StripAlphanumerical(buf.String())
	} else {
		result = Strip(buf.String())
	}

	fmt.Print(result + "\n")

	return result
}
