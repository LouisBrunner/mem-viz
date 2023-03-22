package fetch

import (
	"encoding/binary"
	"fmt"
	"strings"
	"unsafe"

	"github.com/LouisBrunner/dsc-viz/pkg/contracts"
	"github.com/lunixbochs/struc"
)

var unpackOptions = &struc.Options{
	Order: binary.LittleEndian,
}

type fromFile struct {
	fromFileCache
	subCaches []contracts.Cache
}

func (me *fromFile) Close() error {
	for _, subCache := range me.subCaches {
		subCache.Close()
	}
	return me.fromFileCache.Close()
}

func (me *fromFile) SubCaches() ([]contracts.Cache, error) {
	if _, v1 := me.header.V1(); v1 {
		return nil, nil
	}

	if len(me.subCaches) > 0 {
		return me.subCaches, nil
	}

	me.logger.Debugf("file-cache: loading sub caches")
	_, v2 := me.header.V2()

	me.subCaches = make([]contracts.Cache, 0, me.header.SubCacheArrayCount)
	for i := int64(0); i < int64(me.header.SubCacheArrayCount); i += 1 {
		suffix := fmt.Sprintf(".%d", i+1)
		if !v2 && me.header.SubCacheArrayOffset > 0 {
			cacheHeader := contracts.DYLDSubcacheEntryV2{}
			offset := int64(me.header.SubCacheArrayOffset) + i*int64(unsafe.Sizeof(cacheHeader))
			err := struc.UnpackWithOptions(me.ReaderAtOffset(offset), &cacheHeader, unpackOptions)
			if err != nil {
				return nil, err
			}
			suffix = strings.TrimRight(string(cacheHeader.FileSuffix[:]), "\x00")
		}
		cache, err := cacheFromPath(me.logger, fmt.Sprintf("%s%s", me.file.Name(), suffix))
		if err != nil {
			return nil, err
		}
		me.subCaches = append(me.subCaches, cache)
	}

	return me.subCaches, nil
}

func (me *fromFile) String() string {
	return fmt.Sprintf("Caches{main: %v, subs: %+v}", me.fromFileCache.String(), me.subCaches)
}
