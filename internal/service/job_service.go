package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/mprachi301/payment-retry/internal/models"
	"github.com/mprachi301/payment-retry/internal/repository"
)

type JobService struct {
	repo       repository.JobRepository
	maxRetries int
}

func NewJobService(repo repository.JobRepository, maxRetries int) *JobService {
	return &JobService{
		repo:       repo,
		maxRetries: maxRetries,
	}
}

func (s *JobService) Create(paymentID string, amount float64, currency string) (*models.RetryJob, error) {
	job := &models.RetryJob{
		PaymentID:   paymentID,
		Amount:      amount,
		Currency:    currency,
		Status:      models.StatusPending,
		MaxRetries:  s.maxRetries,
		NextRetryAt: time.Now().Add(30 * time.Second),
	}
	err := s.repo.Create(job)
	if err != nil {
		return nil, err
	}
	return job, err
}

func (s *JobService) GetJob(id uuid.UUID) (*models.RetryJob, error) {
	return s.repo.GetById(id)
}
