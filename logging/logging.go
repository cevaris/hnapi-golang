package logging

import (
	"os"

	"github.com/op/go-logging"
)

// NewLogger returns new logger
func NewLogger(name string) *logging.Logger {
	logger := logging.MustGetLogger(name)
	format := logging.MustStringFormatter(
		`%{color}%{time:2006-01-02T15:04:05.999Z-07:00.000} %{level:.3s} %{longfunc}%{color:reset} %{message}`,
	)

	logBackend := logging.NewLogBackend(os.Stdout, "", 0)
	logFormatter := logging.NewBackendFormatter(logBackend, format)
	logging.SetBackend(logFormatter)

	logging.SetLevel(logging.DEBUG, name)

	return logger
}
