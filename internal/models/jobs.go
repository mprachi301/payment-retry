package models

import (
	"time"

	"github.com/google/uuid"
)

type JobStatus string

const (
	StatusPending    JobStatus = "pending"
	StatusProcessing JobStatus = "processing"
	StatusSucceeded  JobStatus = "succeeded"
	StatusFailed     JobStatus = "failed"
	StatusDead       JobStatus = "dead"
)

type RetryJob struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PaymentID   string    `gorm:"not null"`
	Amount      float64   `gorm:"type:decimal(10,2);not null"`
	Currency    string    `gorm:"type:varchar(3);not null;default:INR"`
	Status      JobStatus `gorm:"type:job_status;not null;default:pending"`
	RetryCount  int       `gorm:"not null;default:0"`
	MaxRetries  int       `gorm:"not null;default:3"`
	NextRetryAt time.Time
	LastError   string `gorm:"type:text"`
	WebhookURL  string `gorm:"type:varchar(500)"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
