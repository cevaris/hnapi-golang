package logging

import (
	"context"
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
