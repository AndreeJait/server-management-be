package retry

import (
	"context"
	"math"
	"time"

	"github.com/AndreeJait/go-utility/v2/logw"
)

// Config holds retry settings.
type Config struct {
	MaxAttempts int           // Total number of attempts (including the first)
	Interval    time.Duration // Initial wait between attempts
	MaxInterval time.Duration // Maximum wait between attempts (for backoff cap)
}

// Do retries fn with exponential backoff until it succeeds or ctx is cancelled.
func Do(ctx context.Context, cfg Config, fn func() error) error {
	var err error
	interval := cfg.Interval

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}

		if attempt < cfg.MaxAttempts {
			logw.Warningf("attempt %d/%d failed: %v, retrying in %v...", attempt, cfg.MaxAttempts, err, interval)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(interval):
			}
			interval = time.Duration(math.Min(float64(interval*2), float64(cfg.MaxInterval)))
		}
	}
	return err
}