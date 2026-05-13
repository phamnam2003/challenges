package resource

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

func NewAppContext() (*errgroup.Group, context.Context, context.CancelFunc) {
	sig, cn := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	gr, ctx := errgroup.WithContext(sig)
	return gr, ctx, cn
}
