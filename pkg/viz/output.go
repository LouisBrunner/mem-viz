package viz

import (
	"io"

	"github.com/sirupsen/logrus"
)

type outputter struct {
	logger *logrus.Logger
	w      io.Writer
}

func New(logger *logrus.Logger, w io.Writer) *outputter {
	return &outputter{
		w:      w,
		logger: logger,
	}
}
