package web

import (
	"context"
	"net/http"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

func Serve(ctx context.Context, addr string, handler http.Handler) {
	srv := &http.Server{
		Addr:           addr,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        handler,
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

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Still try to start, application might still do useful work
		otelzap.Ctx(ctx).DPanic("http server failed to start", zap.Error(err))
		return
	}
	otelzap.Ctx(ctx).Info("http server started", zap.String("addr", addr))

	<-idleConnectionsClosed
}
