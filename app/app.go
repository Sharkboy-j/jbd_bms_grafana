package app

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	ctx      context.Context
	once     sync.Once
	Canceled bool
)

// SigTermIntCtx returns context which cancels by SIGTERM or SIGINT
func SigTermIntCtx() context.Context {
	once.Do(func() {
		var cancel context.CancelFunc

		ctx, cancel = context.WithCancel(context.Background())

		var shutdown = make(chan os.Signal)

		signal.Notify(shutdown, syscall.SIGTERM)
		signal.Notify(shutdown, syscall.SIGINT)

		go func() {
			<-shutdown
			println("TERMINATE")
			Canceled = true
			cancel()
		}()
	})

	return ctx
}
