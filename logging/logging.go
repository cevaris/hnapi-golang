package logging

import (
	"os"

	"github.com/op/go-logging"
)

// NewLogger returns new logger
func NewLogger(name string) *logging.Logger {
	result := logging.MustGetLogger(name)
	logBackend := logging.NewLogBackend(os.Stdout, "", 0)
	var format = logging.MustStringFormatter(
		`%{color}%{time:2006-01-02T15:04:05.999Z-07:00.000} %{level:.3s} %{longfunc}%{color:reset} %{message}`,
	)
	logFormatter := logging.NewBackendFormatter(logBackend, format)
	logLeveled := logging.AddModuleLevel(logBackend)
	logLeveled.SetLevel(logging.INFO, name)
	logging.SetBackend(logLeveled, logFormatter)
	return result
}
