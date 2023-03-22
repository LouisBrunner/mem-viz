//go:build !darwin
// +build !darwin

package fetch

import (
	"fmt"
)

func FromMemory(logger *logrus.Logger) (Fetcher, error) {
	return nil, fmt.Errorf("not available on this platform")
}
