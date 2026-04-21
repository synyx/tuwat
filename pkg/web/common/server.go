package common

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"
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
			slog.ErrorContext(ctx, "error shutting down http server", slog.Any("error", err))
		}
		close(idleConnectionsClosed)
	}()

	slog.InfoContext(ctx, "Starting http server", slog.String("addr", "http://"+addr))

	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		// Still try to start, application might still do useful work
		slog.ErrorContext(ctx, "http server failed to start", slog.Any("error", err))
		return
	}

	<-idleConnectionsClosed
}
