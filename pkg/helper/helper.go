package helper

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

func NewErrGroupAndContext() (*errgroup.Group, context.Context, context.CancelFunc) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	g, ctx := errgroup.WithContext(ctx)
	return g, ctx, cancel
}

func WaitForErrGroup(g *errgroup.Group) error {
	err := g.Wait()
	if err != nil {
		return fmt.Errorf("error groups error: %w", err)
	}

	return nil
}

func NewShutdownTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 2*time.Second)
}

func WaitForStop(stopChan chan os.Signal, ctx context.Context) string {
	select {
	case sig := <-stopChan:
		return fmt.Sprintf("os.Signal (%s)", sig)
	case <-ctx.Done():
		return "context"
	}
}

func NewStopChannel() chan os.Signal {
	stopChan := make(chan os.Signal, 2)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGPIPE)
	return stopChan
}
