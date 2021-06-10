package helper

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type ServiceStarter interface {
	Start(ctx context.Context) error
}

type ServiceStopper interface {
	Stop(ctx context.Context) error
}

func StartService(ctx context.Context, g *errgroup.Group, s ServiceStarter) {
	g.Go(func() error {
		return s.Start(ctx)
	})
}

func StopService(ctx context.Context, g *errgroup.Group, s ServiceStopper) {
	g.Go(func() error {
		return s.Stop(ctx)
	})
}
