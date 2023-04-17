package contracts

import "io"

type Address interface {
	AddBase(base uintptr) Address
	Calculate(slide uint64) uintptr
	GetReader(cache Cache, offset, slide uint64) io.Reader
	Invalid() bool
}

type RelativeAddress32 uint32

func (me RelativeAddress32) AddBase(base uintptr) Address {
	return RelativeAddress64(uint64(base) + uint64(me))
}

func (me RelativeAddress32) Calculate(slide uint64) uintptr {
	return uintptr(me)
}

func (me RelativeAddress32) GetReader(cache Cache, offset, slide uint64) io.Reader {
	return cache.ReaderAtOffset(int64(me) + int64(offset))
}

func (me RelativeAddress32) Invalid() bool {
	return me == 0
}

type RelativeAddress64 uint64

func (me RelativeAddress64) AddBase(base uintptr) Address {
	return RelativeAddress64(uint64(base) + uint64(me))
}

func (me RelativeAddress64) Calculate(slide uint64) uintptr {
	return uintptr(me)
}

func (me RelativeAddress64) GetReader(cache Cache, offset, slide uint64) io.Reader {
	return cache.ReaderAtOffset(int64(me) + int64(offset))
}

func (me RelativeAddress64) Invalid() bool {
	return me == 0
}

type ManualAddress uint64

func (me ManualAddress) AddBase(base uintptr) Address {
	return ManualAddress(uint64(base) + uint64(me))
}

func (me ManualAddress) Calculate(slide uint64) uintptr {
	return uintptr(me)
}

func (me ManualAddress) GetReader(cache Cache, offset, slide uint64) io.Reader {
	return cache.ReaderAtOffset(int64(me) + int64(offset))
}

func (me ManualAddress) Invalid() bool {
	return false
}

type UnslidAddress uint64

func (me UnslidAddress) AddBase(base uintptr) Address {
	return me
}

func (me UnslidAddress) Calculate(slide uint64) uintptr {
	return uintptr(me) + uintptr(slide)
}

func (me UnslidAddress) GetReader(cache Cache, offset, slide uint64) io.Reader {
	return cache.ReaderAbsolute(uint64(me) + offset + slide)
}

func (me UnslidAddress) Invalid() bool {
	return me == 0
}

type UnslidAddress32 uint32

func (me UnslidAddress32) AddBase(base uintptr) Address {
	return me
}

func (me UnslidAddress32) Calculate(slide uint64) uintptr {
	return uintptr(me) + uintptr(slide)
}

func (me UnslidAddress32) GetReader(cache Cache, offset, slide uint64) io.Reader {
	return cache.ReaderAbsolute(uint64(me) + offset + slide)
}

func (me UnslidAddress32) Invalid() bool {
	return me == 0
}

type LinkEditOffset uint32
