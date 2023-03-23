package commons

import "strings"

func FromCString(b []byte) string {
	return strings.TrimRight(string(b[:]), "\x00")
}
