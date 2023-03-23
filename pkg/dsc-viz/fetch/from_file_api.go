package fetch

import (
	"fmt"
	"path"
	"runtime"

	"github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
	"github.com/sirupsen/logrus"
)

func FromFile(logger *logrus.Logger, path string) (_ contracts.Fetcher, ferr error) {
	cache, err := cacheFromPath(logger, path)
	if err != nil {
		return nil, err
	}

	return newFetcher(logger, cache, &fromFileProcessor{})
}

func ScanForFileWithArchitecture(logger *logrus.Logger, arch string) (contracts.Fetcher, error) {
	archSuffixes, found := archToSuffixes[arch]
	if !found {
		return nil, fmt.Errorf("unsupported architecture %s", arch)
	}
	osPath, found := osToPath[runtime.GOOS]
	if !found {
		return nil, fmt.Errorf("unsupported os %s", runtime.GOOS)
	}

	for _, prefixes := range append(contracts.CryptexPrefixes, "") {
		for _, suffix := range archSuffixes {
			try := path.Join(prefixes, osPath, fmt.Sprintf("%s%s", contracts.DYLD_SHARED_CACHE_BASE_NAME, suffix))
			logger.Debugf("scanning: trying to load %q", try)
			res, err := FromFile(logger, try)
			if err == nil {
				return res, nil
			}
			logger.Errorf("scanning: Could not load %q: %v", try, err)
		}
	}

	return nil, fmt.Errorf("could not find file for architecture %s", arch)
}

var archToSuffixes = map[string][]string{
	"amd64": {"x86_64h", "x86_64"},
	"arm64": {"arm64e", "arm64", "arm64_32"},
	"arm":   {"armv7", "armv6", "armv5"},
	"386":   {"i386"},
}

var osToPath = map[string]string{
	"ios":    contracts.IPHONE_DYLD_SHARED_CACHE_DIR,
	"darwin": contracts.MACOSX_DYLD_SHARED_CACHE_DIR,
}

func ScanForFileWithCurrentArchitecture(logger *logrus.Logger) (contracts.Fetcher, error) {
	return ScanForFileWithArchitecture(logger, runtime.GOARCH)
}
