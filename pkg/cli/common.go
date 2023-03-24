package cli

import (
	"fmt"
	"os"
	"regexp"
	"runtime"

	"github.com/LouisBrunner/mem-viz/pkg/commons"
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	"github.com/LouisBrunner/mem-viz/pkg/viz"
	"github.com/sirupsen/logrus"
)

var funcMatcher = regexp.MustCompile(`^.*/([^/]*)$`)
var filenameMatcher = regexp.MustCompile(`^.*/?((cmd|pkg)/.+)$`)

func getLogger(params Args) *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(params.LoggingLevel)
	logger.SetOutput(os.Stderr)
	if os.Getenv("DEBUG") != "" {
		logger.SetLevel(logrus.DebugLevel)
		logger.SetReportCaller(true)
	}
	logger.Formatter = &logrus.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			cleanFunc := f.Function
			matches := funcMatcher.FindStringSubmatch(f.Function)
			if len(matches) > 1 {
				cleanFunc = matches[1]
			}
			matches = filenameMatcher.FindStringSubmatch(f.File)
			cleanFile := f.File
			if len(matches) > 1 {
				cleanFile = matches[1]
			}
			return fmt.Sprintf("{%s}", cleanFunc), fmt.Sprintf(" %s:%d", cleanFile, f.Line)
		},
	}
	return logger
}

func fetchJSON(logger *logrus.Logger, params Args) (*contracts.MemoryBlock, error) {
	var mb *contracts.MemoryBlock
	var err error

	if params.FromJSONFile != "" {
		mb, err = commons.FromJSONFile(logger, params.FromJSONFile)
	} else if params.FromJSONText != "" {
		mb, err = commons.FromJSONText(logger, params.FromJSONText)
	}
	if err != nil {
		return nil, err
	}

	return mb, nil
}

func getOutput(logger *logrus.Logger, outputFormat, outputFile string) (func(mb contracts.MemoryBlock) error, func(), error) {
	var outputFn func(mb contracts.MemoryBlock) error
	cleanupFn := func() {}
	var err error

	w := os.Stdout
	if outputFile != "" {
		w, err = os.Create(outputFile)
		if err != nil {
			return nil, nil, err
		}
		cleanupFn = func() {
			w.Close()
		}
	}

	outputter := viz.New(logger, w)

	switch outputFormat {
	case OutputFormatGraphviz:
		outputFn = outputter.Graphviz
	case OutputFormatLaTeX:
		outputFn = outputter.LaTeX
	case OutputFormatMarkdown:
		outputFn = outputter.Markdown
	case OutputFormatText:
		outputFn = outputter.Text
	case OutputFormatASCII:
		outputFn = outputter.ASCII
	case OutputFormatJSON:
		outputFn = outputter.JSON
	default:
		err = fmt.Errorf("unknown output format: %s", outputFormat)
	}
	if err != nil {
		cleanupFn()
		return nil, nil, err
	}
	return outputFn, cleanupFn, nil
}
