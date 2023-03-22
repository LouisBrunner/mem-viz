package contracts

import (
	"fmt"
	"io"
)

type Cache interface {
	Header() DYLDCacheHeaderV3
	ReaderAtOffset(off int64) io.Reader
	io.Closer
	fmt.Stringer
}

type Fetcher interface {
	Cache
	SubCaches() ([]Cache, error)
	io.Closer
	fmt.Stringer
}
