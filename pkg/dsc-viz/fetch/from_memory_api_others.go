//go:build !darwin
// +build !darwin

package fetch

import (
	"errors"
)

var errNotAvailable = errors.New("not available on this platform")

func sharedRegionCheckNP() (uintptr, error) {
	return 0, errNotAvailable
}
