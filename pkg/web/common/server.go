package common

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

func Serve(ctx context.Context, addr string, handler http.Handler) {
	srv := &http.Server{
		Addr:           addr,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Minute,
		IdleTimeout:    10 * time.Minute,
		MaxHeaderBytes: 1 << 20,
		Handler:        handler,
		BaseContext:    func(net.Listener) context.Context { return ctx },
	}

	idleConnectionsClosed := make(chan struct{})
	go func() {
		<-ctx.Done()

		// We received an interrupt signal, shut down.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			// Error from closing listeners, or context timeout:
			otelzap.Ctx(ctx).Error("error shutting down http server", zap.Error(err))
		}
		close(idleConnectionsClosed)
	}()

	otelzap.Ctx(ctx).Info("Starting http server", zap.String("addr", addr))

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Still try to start, application might still do useful work
		otelzap.Ctx(ctx).DPanic("http server failed to start", zap.Error(err))
		return
	}

	<-idleConnectionsClosed
}
