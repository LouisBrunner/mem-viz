package fetch

import (
	"fmt"
	"unsafe"

	"github.com/LouisBrunner/dsc-viz/pkg/contracts"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

func FromMemory(logger *logrus.Logger) (contracts.Fetcher, error) {
	dsc := uintptr(0)
	res, _, errn := unix.Syscall(unix.SYS_SHARED_REGION_CHECK_NP, uintptr(unsafe.Pointer(&dsc)), 0, 0)
	if errn != 0 {
		return nil, errn
	}
	if res != 0 {
		return nil, fmt.Errorf("shared_region_check_np returned %d", res)
	}
	logger.Debugf("shared_region_check_np returned: %#x", dsc)

	cache, err := cacheFromMemory(logger, dsc)
	if err != nil {
		return nil, err
	}

	return newFetcher(logger, cache, &fromMemoryProcessor{})
}
