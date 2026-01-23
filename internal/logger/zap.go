package logger

import (
	"fmt"
	"strings"

	"github.com/cy77cc/k8s-manage/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	l *zap.Logger
}

func NewZapLogger() Logger {
	level := zap.NewAtomicLevel()
	levelStr := strings.ToLower(config.CFG.Log.Level)
	if err := level.UnmarshalText([]byte(levelStr)); err != nil {
		return nil
	}

	cfg := zap.Config{
		Level:       level,
		Development: config.CFG.App.Debug,
		Encoding:    config.CFG.Log.Format, // 生产推荐 json
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:     "ts",
			LevelKey:    "level",
			MessageKey:  "msg",
			CallerKey:   "caller",
			EncodeTime:  zapcore.ISO8601TimeEncoder,
			EncodeLevel: zapcore.LowercaseLevelEncoder,
		},
		OutputPaths:      []string{"stdout", config.CFG.Log.File.Path},
		ErrorOutputPaths: []string{"stderr", config.CFG.Log.File.Path},
		
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil
	}

	return &zapLogger{l: logger}
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
