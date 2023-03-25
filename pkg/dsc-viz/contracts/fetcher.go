package contracts

import (
	"fmt"
	"io"
)

type Cache interface {
	BaseAddress() uintptr
	Header() DYLDCacheHeaderV3
	ReaderAbsolute(abs uint64) io.Reader
	ReaderAtOffset(off int64) io.Reader
	io.Closer
	fmt.Stringer
}

type SubCache interface {
	Cache
	SubCacheHeader() (*DYLDSubcacheEntryV2, *DYLDSubcacheEntryV1)
}

type Fetcher interface {
	Cache
	SubCaches() []SubCache
	io.Closer
	fmt.Stringer
}
