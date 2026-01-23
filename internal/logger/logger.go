package logger

import (
	"context"
)

type Logger interface {
	Debug(msg string, fields ...Field)
	Debugf(format string, a []any, fields ...Field)
	Info(msg string, fields ...Field)
	Infof(format string, a []any, fields ...Field)
	Warn(msg string, fields ...Field)
	Warnf(msg string, a []any, fields ...Field)
	Error(msg string, fields ...Field)
	Errorf(msg string, a []any, fields ...Field)
	With(fields ...Field) Logger
	WithContext(ctx context.Context) Logger
}

var std Logger

func Init(l Logger) {
	std = l
}

func L() Logger {
	return std
}
