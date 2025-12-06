package logger

import "go.uber.org/zap"

type ZapLogger struct {
	l *zap.Logger
}

func NewZapLogger(l *zap.Logger) *ZapLogger {
	return &ZapLogger{l: l}
}

func (z *ZapLogger) Debug(msg string, fields ...Field) {
	z.l.Debug(msg, z.toZapFields(fields)...)
}

func (z *ZapLogger) Info(msg string, fields ...Field) {
	z.l.Info(msg, z.toZapFields(fields)...)
}

func (z *ZapLogger) Warn(msg string, fields ...Field) {
	z.l.Warn(msg, z.toZapFields(fields)...)
}

func (z *ZapLogger) Error(msg string, fields ...Field) {
	z.l.Error(msg, z.toZapFields(fields)...)
}

func (z *ZapLogger) With(fields ...Field) Logger {
	return &ZapLogger{l: z.l.With(z.toZapFields(fields)...)}
}

func (z *ZapLogger) toZapFields(args []Field) []zap.Field {
	res := make([]zap.Field, 0, len(args))
	for _, arg := range args {
		res = append(res, zap.Any(arg.Key, arg.Value))
	}
	return res
}
