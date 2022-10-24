package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func ApplicationContext() context.Context {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	return ctx
}
