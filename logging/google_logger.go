package logging

import (
	"context"
	"strings"

	"google.golang.org/appengine/log"
)

type googleLoggger struct {
}

// NewGoogleLogger is an google app engine backed logger
func NewGoogleLogger() Logger {
	return &googleLoggger{}
}

func (l *googleLoggger) Info(ctx context.Context, m ...interface{}) {
	if Level == Info || Level == Error {
		log.Infof(ctx, strings.Repeat("%v ", len(m)), m...)
	}
}

func (l *googleLoggger) Error(ctx context.Context, m ...interface{}) {
	// always log errors
	log.Errorf(ctx, strings.Repeat("%v ", len(m)), m...)
}

func (l *googleLoggger) Debug(ctx context.Context, m ...interface{}) {
	if Level == Debug {
		log.Debugf(ctx, strings.Repeat("%v ", len(m)), m...)
	}
}
