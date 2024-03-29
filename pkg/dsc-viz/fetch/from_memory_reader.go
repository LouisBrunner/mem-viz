package fetch

import (
	"sync"
	"unsafe"
)

type fromMemoryReader struct {
	// FIXME: would be nice to have a size too?
	pointer uintptr
	off     int64
	mutex   sync.Mutex
}

func (me *fromMemoryReader) Read(p []byte) (int, error) {
	me.mutex.Lock()
	defer me.mutex.Unlock()

	n, err := me.ReadAt(p, me.off)
	me.off += int64(n)
	return n, err
}

// suppress go vet warning
func unsafeExternPointer(addr uintptr) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&addr))
}

func (me *fromMemoryReader) ReadAt(p []byte, off int64) (int, error) {
	for i := range p {
		p[i] = *(*byte)(unsafeExternPointer(me.pointer + uintptr(off) + uintptr(i)))
	}
	return len(p), nil
}
