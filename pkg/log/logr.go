package log

import (
	"github.com/go-logr/logr"
	"go.uber.org/zap"
)

type logrZapBridge struct {
	*zap.Logger
}

func NewLogrZapBridge(log *zap.Logger) logr.Logger {
	sink := logrZapBridge{log}
	logger := logr.New(sink)
	return logger
}

func (o logrZapBridge) Init(info logr.RuntimeInfo) {
	// do nothing
}

func (o logrZapBridge) Enabled(level int) bool {
	return true
}

func (o logrZapBridge) Info(level int, msg string, data ...interface{}) {
	log := o.Logger.WithOptions(zap.AddCallerSkip(3))
	fields := logrZapFields(data)
	fields = append(fields, zap.Int("logr.level", level))

	if level == 0 {
		log.Info(msg, fields...)
	} else if level > 0 {
		log.Debug(msg, fields...)
	}
}

func (o logrZapBridge) Error(err error, msg string, data ...interface{}) {
	log := o.Logger.WithOptions(zap.AddCallerSkip(3))
	fields := logrZapFields(data)
	fields = append(fields, zap.Error(err))

	log.Error(msg, fields...)
}

func (o logrZapBridge) WithValues(data ...interface{}) logr.LogSink {
	fields := logrZapFields(data)
	return logrZapBridge{o.Logger.With(fields...)}
}

func (o logrZapBridge) WithName(name string) logr.LogSink {
	return logrZapBridge{o.Logger.Named(name)}
}

func logrZapFields(data []interface{}) (fields []zap.Field) {
	for i := 0; i < len(data); i += 2 {
		key := data[i]
		value := data[i+1]
		if _, ok := key.(string); !ok {
			continue
		}
		fields = append(fields, zap.Any(key.(string), value))
	}
	return fields
}
