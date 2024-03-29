package cli

import (
	"fmt"
	"os"

	"github.com/LouisBrunner/mem-viz/pkg/checker"
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

type Worker[T any] struct {
	AddFlags       func(params *T)
	CheckExtraFrom func(params T) ([]bool, []string)
	GetMemory      func(logger *logrus.Logger, params T) (*contracts.MemoryBlock, error)
}

func Main[T any](name string, userParams T, worker Worker[T]) {
	err := work(worker, userParams)
	if err != nil {
		if err == pflag.ErrHelp {
			os.Exit(2)
		} else {
			fmt.Fprintf(os.Stderr, "%s: %s\n", name, err.Error())
			os.Exit(1)
		}
	}
}

func work[T any](worker Worker[T], userParams T) error {
	params := GetDefaultArgs()
	err := ParseArgs(&params, &userParams, worker.AddFlags, worker.CheckExtraFrom)
	if err != nil {
		return err
	}

	logger := getLogger(params)

	mb, err := fetchJSON(logger, params)
	if err != nil {
		return err
	}
	mbInternal := false

	// If we loaded the MemoryBlocks directly from JSON, then no need to parse anything
	if mb == nil {
		mbInternal = true
		mb, err = worker.GetMemory(logger, userParams)
		if err != nil {
			return err
		}
	} else {
		logger.Debug("skipping the parsing, we got JSON")
	}

	err = checker.Check(logger, mb)
	// FIXME: gross? genius? both?
	if os.Getenv("DEBUG_OVERRIDE_CHECKER") != "" && err != nil {
		logger.WithError(err).Warn("DEBUG_OVERRIDE_CHECKER is set, ignoring errors")
		err = nil
	}
	if err != nil {
		if mbInternal {
			message := "internal error, the front-end has generated an invalid memory map"
			logger.WithError(err).Error(message)
			return fmt.Errorf("%s (see log for details)", message)
		}
		return err
	}

	outputFn, cleanupFn, err := getOutput(logger, params.OutputFormat, params.OutputFile)
	if err != nil {
		return err
	}
	defer cleanupFn()

	return outputFn(*mb)
}
