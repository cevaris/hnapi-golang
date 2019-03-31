package logging

import (
	"context"
	"os"
	"strings"

	"github.com/op/go-logging"
	glog "google.golang.org/appengine/log"
)

const (
	// Debug logs debug mode
	Debug = iota
	// Info logs info mode
	Info = iota
	// Error logs error mode
	Error = iota
)

// Level is the global logging level
const Level int = Info

// Logger is a generic logger
type Logger interface {
	Info(context.Context, ...interface{})
	Error(context.Context, ...interface{})
	Debug(context.Context, ...interface{})
}

// NewLogger returns new logger
func NewLogger(name string) *logging.Logger {
	logger := logging.MustGetLogger(name)
	format := logging.MustStringFormatter(
		`%{color}%{time:2006-01-02T15:04:05.999Z-07:00.000} %{level:.3s} %{longfunc}%{color:reset} %{message}`,
	)

	logBackend := logging.NewLogBackend(os.Stdout, "", 0)
	logFormatter := logging.NewBackendFormatter(logBackend, format)
	logging.SetBackend(logFormatter)

	// logging.SetLevel(logging.INFO, "")
	switch Level {
	case Info:
		logging.SetLevel(logging.INFO, "")
		break
	case Error:
		logging.SetLevel(logging.ERROR, "")
		break
	default:
		// defaulting to debug
		logging.SetLevel(logging.DEBUG, "")
	}

	return logger
}

type googleLoggger struct {
}

// NewGoogleLogger is an google app engine backed logger
func NewGoogleLogger() Logger {
	return &googleLoggger{}
}

func (l *googleLoggger) Info(ctx context.Context, m ...interface{}) {
	if Level == Info || Level == Error {
		glog.Infof(ctx, strings.Repeat("%v ", len(m)), m...)
	}
}

func (l *googleLoggger) Error(ctx context.Context, m ...interface{}) {
	// always log errors
	glog.Errorf(ctx, strings.Repeat("%v ", len(m)), m...)
}

func (l *googleLoggger) Debug(ctx context.Context, m ...interface{}) {
	if Level == Debug {
		glog.Debugf(ctx, strings.Repeat("%v ", len(m)), m...)
	}
}

type opLogger struct {
	logger *logging.Logger
}

// NewOpLogger is an op-logging backed logger
func NewOpLogger(name string) Logger {
	logger := logging.MustGetLogger(name)
	format := logging.MustStringFormatter(
		`%{color}%{time:2006-01-02T15:04:05.999Z-07:00.000} %{level:.3s} %{longfunc}%{color:reset} %{message}`,
	)

	logBackend := logging.NewLogBackend(os.Stdout, "", 0)
	logFormatter := logging.NewBackendFormatter(logBackend, format)
	logging.SetBackend(logFormatter)

	// logging.SetLevel(logging.INFO, "")
	logging.SetLevel(logging.DEBUG, "")
	return &opLogger{logger: logger}
}

func (l *opLogger) Info(ctx context.Context, m ...interface{}) {
	l.logger.Info(m)
}

func (l *opLogger) Error(ctx context.Context, m ...interface{}) {
	l.logger.Error(m)
}
func (l *opLogger) Debug(ctx context.Context, m ...interface{}) {
	l.logger.Debug(m)
}
