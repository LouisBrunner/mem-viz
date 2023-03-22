package fetch

import (
	"fmt"
	"io"

	"github.com/LouisBrunner/dsc-viz/pkg/contracts"
	"github.com/sirupsen/logrus"
)

type fromMemoryCache struct {
	pointer uintptr
	header  contracts.DYLDCacheHeaderV3
}

func (me *fromMemoryCache) Close() error {
	return nil
}

func (me *fromMemoryCache) Header() contracts.DYLDCacheHeaderV3 {
	return me.header
}

func (me *fromMemoryCache) ReaderAtOffset(off int64) io.Reader {
	return &fromMemoryReader{pointer: me.pointer + uintptr(off)}
}

func (me *fromMemoryCache) String() string {
	return fmt.Sprintf("Memory{pointer: %#x, header: %+v}", me.pointer, me.header)
}

func cacheFromMemory(logger *logrus.Logger, pointer uintptr) (_ *fromMemoryCache, ferr error) {
	logger.Debugf("memory-cache: loading cache from %#x", pointer)
	defer func() {
		if ferr != nil {
			logger.Errorf("memory-cache: failed to load cache from %#x: %v", pointer, ferr)
		} else {
			logger.Debugf("memory-cache: loaded cache from %#x", pointer)
		}
	}()

	mem := &fromMemoryCache{pointer: pointer}
	err := unpack(mem.ReaderAtOffset(0), &mem.header)
	if err != nil {
		return nil, err
	}
	return mem, nil
}

type fromMemoryProcessor struct{}

func (me fromMemoryProcessor) CacheFromEntryV2(logger *logrus.Logger, main *fromMemoryCache, i int64, entry contracts.DYLDSubcacheEntryV2) (contracts.Cache, error) {
	return cacheFromMemory(logger, main.pointer+uintptr(entry.CacheVmOffset))
}

func (me fromMemoryProcessor) CacheFromEntryV1(logger *logrus.Logger, main *fromMemoryCache, i int64, entry contracts.DYLDSubcacheEntryV1) (contracts.Cache, error) {
	return cacheFromMemory(logger, main.pointer+uintptr(entry.CacheVmOffset))
}
