package tunnel

import (
	"context"
	"log/slog"
	"time"
)

const (
	minBackoff = time.Second
	maxBackoff = 30 * time.Second
	stableFor  = time.Minute
)

func Maintain(ctx context.Context, spec Spec, log *slog.Logger) {
	backoff := minBackoff
	for {
		if ctx.Err() != nil {
			return
		}
		started := time.Now()
		tun, err := start(spec)
		if err != nil {
			log.Warn("ssh start failed, retry",
				slog.Duration("wait", backoff),
				slog.Any("err", err),
			)
			if sleep(ctx, backoff) != nil {
				return
			}
			backoff = nextBackoff(backoff, 0)
			continue
		}
		log.Info("ssh started")
		err = waitOrStop(ctx, tun)
		if ctx.Err() != nil {
			return
		}
		lived := time.Since(started)
		log.Warn("ssh exited, restart",
			slog.Duration("up", lived),
			slog.Any("err", err),
		)
		backoff = nextBackoff(backoff, lived)
		if sleep(ctx, backoff) != nil {
			return
		}
	}
}

func waitOrStop(ctx context.Context, tun *tunnel) error {
	done := make(chan error, 1)
	go func() { done <- tun.wait() }()
	select {
	case <-ctx.Done():
		tun.kill()
		return <-done
	case err := <-done:
		return err
	}
}

func nextBackoff(cur, lived time.Duration) time.Duration {
	if lived >= stableFor {
		return minBackoff
	}
	next := cur * 2
	if next > maxBackoff {
		return maxBackoff
	}
	return next
}

func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
