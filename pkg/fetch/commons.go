package fetch

import (
	"encoding/binary"
	"io"

	"github.com/lunixbochs/struc"
)

var unpackOptions = &struc.Options{
	Order: binary.LittleEndian,
}

func unpack(r io.Reader, v interface{}) error {
	return struc.UnpackWithOptions(r, v, unpackOptions)
}
