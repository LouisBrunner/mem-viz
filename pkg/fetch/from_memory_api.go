package fetch

import (
	"github.com/LouisBrunner/dsc-viz/pkg/contracts"
	"github.com/sirupsen/logrus"
)

func FromMemory(logger *logrus.Logger) (contracts.Fetcher, error) {
	dsc, err := sharedRegionCheckNP()
	if err != nil {
		return nil, err
	}
	logger.Debugf("shared_region_check_np returned: %#x", dsc)

	cache, err := cacheFromMemory(logger, dsc)
	if err != nil {
		return nil, err
	}

	return newFetcher(logger, cache, &fromMemoryProcessor{})
}
