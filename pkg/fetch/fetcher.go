package fetch

import (
	"fmt"
	"io"
	"unsafe"

	"github.com/LouisBrunner/dsc-viz/pkg/contracts"
	"github.com/sirupsen/logrus"
)

type fetchProcessor[T contracts.Cache] interface {
	CacheFromEntryV2(logger *logrus.Logger, main T, i int64, entry contracts.DYLDSubcacheEntryV2) (contracts.Cache, error)
	CacheFromEntryV1(logger *logrus.Logger, main T, i int64, entry contracts.DYLDSubcacheEntryV1) (contracts.Cache, error)
}

type fetcher[T contracts.Cache, F fetchProcessor[T]] struct {
	logger    *logrus.Logger
	processor F
	main      T
	subs      []contracts.Cache
}

func (me *fetcher[T, F]) Close() error {
	for _, sub := range me.subs {
		sub.Close()
	}
	return me.main.Close()
}

func (me *fetcher[T, F]) SubCaches() ([]contracts.Cache, error) {
	if len(me.subs) > 0 {
		return me.subs, nil
	}

	header := me.main.Header()
	if _, v1 := header.V1(); v1 {
		return nil, nil
	}

	me.logger.Debugf("file-cache: loading sub caches")
	_, v2 := header.V2()

	me.subs = make([]contracts.Cache, 0, header.SubCacheArrayCount)
	for i := int64(0); i < int64(header.SubCacheArrayCount); i += 1 {
		var cache contracts.Cache
		var err error

		if !v2 {
			cacheHeader := contracts.DYLDSubcacheEntryV2{}
			offset := int64(header.SubCacheArrayOffset) + i*int64(unsafe.Sizeof(cacheHeader))
			err = unpack(me.main.ReaderAtOffset(offset), &cacheHeader)
			if err != nil {
				return nil, err
			}
			cache, err = me.processor.CacheFromEntryV2(me.logger, me.main, i, cacheHeader)
		} else {
			cacheHeader := contracts.DYLDSubcacheEntryV1{}
			offset := int64(header.SubCacheArrayOffset) + i*int64(unsafe.Sizeof(cacheHeader))
			err = unpack(me.main.ReaderAtOffset(offset), &cacheHeader)
			if err != nil {
				return nil, err
			}
			cache, err = me.processor.CacheFromEntryV1(me.logger, me.main, i, cacheHeader)
		}

		if err != nil {
			return nil, err
		}
		me.subs = append(me.subs, cache)
	}

	return me.subs, nil
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

func newFetcher[T contracts.Cache, F fetchProcessor[T]](logger *logrus.Logger, main T, processor F) (_ contracts.Fetcher, ferr error) {
	fetcher := &fetcher[T, F]{logger: logger, main: main, processor: processor}
	defer func() {
		if ferr != nil {
			fetcher.Close()
		}
	}()
	_, err := fetcher.SubCaches()
	if err != nil {
		return nil, err
	}
	return fetcher, nil
}
