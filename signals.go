package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func withSignals(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-signals
		fmt.Printf("\n==> Received %s, stopping\n", s)
		fmt.Printf("    Send %s again for instant shutdown\n", s)
		cancel()

		go func() {
			s := <-signals
			fmt.Printf("\n==> Received %s, exiting aggressively\n", s)
			os.Exit(1)
		}()
	}()

	return ctx
}
