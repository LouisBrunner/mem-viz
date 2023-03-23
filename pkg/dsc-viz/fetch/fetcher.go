package fetch

import (
	"fmt"
	"io"
	"unsafe"

	"github.com/LouisBrunner/mem-viz/pkg/commons"
	"github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
	"github.com/sirupsen/logrus"
)

type fetchProcessor[T contracts.Cache] interface {
	CacheFromEntryV2(logger *logrus.Logger, main T, i int64, entry contracts.DYLDSubcacheEntryV2) (contracts.Cache, error)
	CacheFromEntryV1(logger *logrus.Logger, main T, i int64, entry contracts.DYLDSubcacheEntryV1) (contracts.Cache, error)
}

type fetcherSubCache struct {
	contracts.Cache
	headerV2 *contracts.DYLDSubcacheEntryV2
	headerV1 *contracts.DYLDSubcacheEntryV1
}

func (me *fetcherSubCache) BaseAddress() uintptr {
	if me.headerV2 != nil {
		return uintptr(me.headerV2.CacheVmOffset)
	} else if me.headerV1 != nil {
		return uintptr(me.headerV1.CacheVmOffset)
	}
	// FIXME: technically an error
	return 0
}

func (me *fetcherSubCache) SubCacheHeader() (*contracts.DYLDSubcacheEntryV2, *contracts.DYLDSubcacheEntryV1) {
	return me.headerV2, me.headerV1
}

type fetcher[T contracts.Cache, F fetchProcessor[T]] struct {
	logger    *logrus.Logger
	processor F
	main      T
	subs      []contracts.SubCache
}

func (me *fetcher[T, F]) Close() error {
	for _, sub := range me.subs {
		sub.Close()
	}
	return me.main.Close()
}

func (me *fetcher[T, F]) SubCaches() []contracts.SubCache {
	return me.subs
}

func (me *fetcher[T, F]) fetchSubCaches() error {
	if len(me.subs) > 0 {
		return nil
	}

	header := me.main.Header()
	if _, v1 := header.V1(); v1 {
		return nil
	}

	me.logger.Debugf("file-cache: loading sub caches")
	_, v2 := header.V2()

	me.subs = make([]contracts.SubCache, 0, header.SubCacheArrayCount)
	for i := int64(0); i < int64(header.SubCacheArrayCount); i += 1 {
		var cache contracts.Cache
		var err error
		var cacheHeaderV2 *contracts.DYLDSubcacheEntryV2
		var cacheHeaderV1 *contracts.DYLDSubcacheEntryV1

		if !v2 {
			cacheHeaderV2 = &contracts.DYLDSubcacheEntryV2{}
			offset := int64(header.SubCacheArrayOffset) + i*int64(unsafe.Sizeof(*cacheHeaderV2))
			err = commons.Unpack(me.main.ReaderAtOffset(offset), cacheHeaderV2)
			if err != nil {
				return err
			}
			cache, err = me.processor.CacheFromEntryV2(me.logger, me.main, i, *cacheHeaderV2)
		} else {
			cacheHeaderV1 = &contracts.DYLDSubcacheEntryV1{}
			offset := int64(header.SubCacheArrayOffset) + i*int64(unsafe.Sizeof(*cacheHeaderV1))
			err = commons.Unpack(me.main.ReaderAtOffset(offset), &cacheHeaderV1)
			if err != nil {
				return err
			}
			cache, err = me.processor.CacheFromEntryV1(me.logger, me.main, i, *cacheHeaderV1)
		}

		if err != nil {
			return err
		}
		me.subs = append(me.subs, &fetcherSubCache{Cache: cache, headerV2: cacheHeaderV2, headerV1: cacheHeaderV1})
	}

	return nil
}

func (me *fetcher[T, F]) String() string {
	return fmt.Sprintf("Caches{main: %s, subs: %+v}", me.main.String(), me.subs)
}

// rewrapped

func (me *fetcher[T, F]) Header() contracts.DYLDCacheHeaderV3 {
	return me.main.Header()
}

func (me *fetcher[T, F]) ReaderAtOffset(offset int64) io.Reader {
	return me.main.ReaderAtOffset(offset)
}

func (me *fetcher[T, F]) BaseAddress() uintptr {
	return me.main.BaseAddress()
}

func newFetcher[T contracts.Cache, F fetchProcessor[T]](logger *logrus.Logger, main T, processor F) (_ contracts.Fetcher, ferr error) {
	fetcher := &fetcher[T, F]{logger: logger, main: main, processor: processor}
	defer func() {
		if ferr != nil {
			fetcher.Close()
		}
	}()
	err := fetcher.fetchSubCaches()
	if err != nil {
		return nil, err
	}
	return fetcher, nil
}
