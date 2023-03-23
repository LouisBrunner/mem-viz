//go:build !darwin
// +build !darwin

package fetch

import (
	"fmt"
)

func sharedRegionCheckNP() (uintptr, error) {
	return 0, fmt.Errorf("not available on this platform")
}
