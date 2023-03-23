package fetch

import (
	"fmt"
	"io"
	"math"
	"os"

	"github.com/LouisBrunner/mem-viz/pkg/commons"
	"github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
	"github.com/sirupsen/logrus"
)

type fromFileCache struct {
	logger *logrus.Logger
	file   *os.File
	header contracts.DYLDCacheHeaderV3
}

func (me *fromFileCache) Close() error {
	return me.file.Close()
}

func (me *fromFileCache) ReaderAtOffset(off int64) io.Reader {
	// FIXME: better bounds?
	return io.NewSectionReader(me.file, off, math.MaxInt64)
}

func (me *fromFileCache) Header() contracts.DYLDCacheHeaderV3 {
	return me.header
}

func (me *fromFileCache) BaseAddress() uintptr {
	return 0
}

func (me *fromFileCache) String() string {
	return fmt.Sprintf("File{path: %s, header: %+v}", me.file.Name(), me.header)
}

func cacheFromFile(logger *logrus.Logger, file *os.File) (*fromFileCache, error) {
	cache := &fromFileCache{logger: logger, file: file}
	err := commons.Unpack(file, &cache.header)
	if err != nil {
		return nil, err
	}
	if string(cache.header.Magic[:])[0:7] != "dyld_v1" {
		return nil, fmt.Errorf("invalid magic %s", cache.header.Magic)
	}
	return cache, nil
}

func cacheFromPath(logger *logrus.Logger, path string) (_ *fromFileCache, ferr error) {
	logger.Debugf("file-cache: loading cache from %q", path)
	defer func() {
		if ferr != nil {
			logger.Errorf("file-cache: failed to load cache from %q: %v", path, ferr)
		} else {
			logger.Debugf("file-cache: loaded cache from %q", path)
		}
	}()

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if ferr != nil {
			file.Close()
		}
	}()

	return cacheFromFile(logger, file)
}

type fromFileProcessor struct{}

func (me fromFileProcessor) CacheFromEntryV2(logger *logrus.Logger, main *fromFileCache, i int64, entry contracts.DYLDSubcacheEntryV2) (contracts.Cache, error) {
	suffix := commons.FromCString(entry.FileSuffix[:])
	return cacheFromPath(logger, fmt.Sprintf("%s%s", main.file.Name(), suffix))
}

func (me fromFileProcessor) CacheFromEntryV1(logger *logrus.Logger, main *fromFileCache, i int64, entry contracts.DYLDSubcacheEntryV1) (contracts.Cache, error) {
	return cacheFromPath(logger, fmt.Sprintf("%s.%d", main.file.Name(), i+1))
}
