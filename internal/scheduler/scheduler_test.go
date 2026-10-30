package scheduler

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunExecutesTaskImmediatelyAndRepeats(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var calls atomic.Int32
	done := make(chan error, 1)
	go func() {
		done <- Run(ctx, 20*time.Millisecond, func() error {
			if calls.Add(1) == 2 {
				cancel()
			}
			return nil
		})
	}()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Run error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run did not stop after context cancellation")
	}

	if got := calls.Load(); got != 2 {
		t.Fatalf("task called %d times, want 2", got)
	}
}

func TestRunReturnsTaskError(t *testing.T) {
	wantErr := errors.New("task failed")
	var calls int
	err := Run(context.Background(), time.Hour, func() error {
		calls++
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Run error = %v, want %v", err, wantErr)
	}
	if calls != 1 {
		t.Fatalf("task called %d times, want 1", calls)
	}
}

func TestRunStopsWhenContextAlreadyCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	calls := 0
	err := Run(ctx, time.Millisecond, func() error {
		calls++
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run error = %v, want context.Canceled", err)
	}
	if calls != 1 {
		t.Fatalf("task called %d times, want immediate initial call", calls)
	}
}
