package scheduler

import (
	"context"
	"time"
)

func Run(
	ctx context.Context,
	interval time.Duration,
	task func() error,
) error {
	for {
		if err := task(); err != nil {
			return err
		}

		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}
