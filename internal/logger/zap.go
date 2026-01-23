package logger

import (
	"fmt"

	"go.uber.org/zap"
)

type zapLogger struct {
	l *zap.Logger
}

func NewZapLogger(l *zap.Logger) Logger {
	return &zapLogger{l: l}
}

func (z *zapLogger) Info(msg string, fields ...Field) {
	z.l.Info(msg, toZapFields(fields)...)
}

func (z *zapLogger) Infof(format string, a []any, fields ...Field) {
	z.l.Info(fmt.Sprintf(format, a...), toZapFields(fields)...)
}

func (z *zapLogger) Error(msg string, fields ...Field) {
	z.l.Error(msg, toZapFields(fields)...)
}

func (z *zapLogger) Errorf(format string, a []any, fields ...Field) {
	z.l.Info(fmt.Sprintf(format, a...), toZapFields(fields)...)
}

// Debug / Warn 同理

func (z *zapLogger) Debug(msg string, fields ...Field) {
	z.l.Error(msg, toZapFields(fields)...)
}

func (z *zapLogger) Debugf(format string, a []any, fields ...Field) {
	z.l.Info(fmt.Sprintf(format, a...), toZapFields(fields)...)
}

func (z *zapLogger) Warn(msg string, fields ...Field) {
	z.l.Error(msg, toZapFields(fields)...)
}

func (z *zapLogger) Warnf(format string, a []any, fields ...Field) {
	z.l.Info(fmt.Sprintf(format, a...), toZapFields(fields)...)
}

func (z *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{
		l: z.l.With(toZapFields(fields)...),
	}
}

func toZapFields(fields []Field) []zap.Field {
	zfs := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		zfs = append(zfs, zap.Any(f.Key, f.Value))
	}
	return zfs
}
