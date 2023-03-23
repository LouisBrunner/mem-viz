package commons

import (
	"io"

	"github.com/lunixbochs/struc"
)

func Unpack(r io.Reader, v interface{}) error {
	return struc.Unpack(r, v)
}
