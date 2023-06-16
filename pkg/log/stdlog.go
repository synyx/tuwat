package log

import (
	"bytes"
	"log"

	"go.uber.org/zap"
)

type stdLogger struct {
	*zap.Logger
}

func newStdLoggerBridge(zapLogger *zap.Logger) stdLogger {
	l := stdLogger{zapLogger}
	return l
}

func (l stdLogger) Write(p []byte) (n int, err error) {
	lg := l.Logger.WithOptions(zap.AddCallerSkip(3))

	// XXX: Remove messages we can do nothing about.
	if bytes.Contains(p, []byte("http: superfluous response.WriteHeade")) {
		return len(p), nil
	}

	lg.Info(string(p))

	return len(p), nil
}

func (l stdLogger) ReplaceGlobals() func() {

	o := log.Writer()
	p := log.Prefix()
	f := log.Flags()

	log.SetOutput(l)
	log.SetPrefix("")
	log.SetFlags(0)

	return func() {
		log.SetOutput(o)
		log.SetPrefix(p)
		log.SetFlags(f)
	}
}
