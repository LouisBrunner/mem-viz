package fetch

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/unix"
)

func sharedRegionCheckNP() (uintptr, error) {
	dsc := uintptr(0)
	//lint:ignore SA1019 the deprecation is for the Go devs, not the users, in any case they don't have a wrapper for this one
	res, _, errn := unix.Syscall(unix.SYS_SHARED_REGION_CHECK_NP, uintptr(unsafe.Pointer(&dsc)), 0, 0)
	if errn != 0 {
		return 0, errn
	}
	if res != 0 {
		return 0, fmt.Errorf("shared_region_check_np returned %d", res)
	}
	return dsc, nil
}
