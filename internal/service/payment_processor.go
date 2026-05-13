package service

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/mprachi301/payment-retry/internal/models"
)

// interface — defines what any payment processor must do
// MockPaymentProcessor implements this today
// StripePaymentProcessor, RazorpayProcessor etc can implement it later

type PaymentProcess interface {
	ProcessPayment(job models.RetryJob) error
}

// MockPaymentProcessor — fake processor for development and testing
type MockPaymentProcessor struct {
	successRate float64 // 0.0 to 1.0 — e.g. 0.3 means 30% success rate
}

func NewMockPaymentProcessor(successRate float64) *MockPaymentProcessor {
	return &MockPaymentProcessor{
		successRate: successRate,
	}
}

func (p *MockPaymentProcessor) ProcessPayment(job models.RetryJob) error {
	// simulate network latency — random between 50ms and 200ms
	latency := time.Duration(50+rand.Intn(150)) * time.Millisecond
	time.Sleep(latency)

	// randomly succeed or fail based on success rate
	if rand.Float64() < p.successRate {
		return nil // success
	}
	// failure — return an error with details
	return fmt.Errorf("payment failed for job %s: amount %.2f %s — gateway timeout", job.ID, job.Amount, job.Currency)
}
