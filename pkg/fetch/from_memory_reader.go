package fetch

import (
	"sync"
	"unsafe"
)

type fromMemoryReader struct {
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

func (me *fromMemoryReader) ReadAt(p []byte, off int64) (int, error) {
	for i := range p {
		p[i] = *(*byte)(unsafe.Pointer(me.pointer + uintptr(off) + uintptr(i)))
	}
	return len(p), nil
}
