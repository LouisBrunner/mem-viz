package parsingutils

import "io"

func ReadCString(r io.Reader) string {
	var b []byte
	one := make([]byte, 1)
	for {
		c, err := r.Read(one)
		if err != nil || c == 0 || one[0] == 0 {
			break
		}
		b = append(b, one[:c]...)
	}
	return string(b)
}
