package service

import (
	"math"
	"math/rand"
	"time"
)

const baseDelay = 30 * time.Second

func CalculateNextRetry(attempt int) time.Time {
	// baseDelay * 2^attempt
	backoff := float64(baseDelay) * math.Pow(2, float64(attempt))

	// add random jitter between 0 and 30 seconds
	jitter := time.Duration(rand.Intn(30)) * time.Second

	return time.Now().Add(time.Duration(backoff) + jitter)
}
