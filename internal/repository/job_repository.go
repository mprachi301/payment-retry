package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/mprachi301/payment-retry/internal/models"
	"gorm.io/gorm"
)

// interface are set of rules which every struct belonging to it should follow
// here JobRepository is an interface which says - every struct belonging to this interface must be able to create a jon and return an error.
// it is just stating the rule not telling how to implement.
type JobRepository interface {
	Create(job *models.RetryJob) error
	GetById(id uuid.UUID) (*models.RetryJob, error)
	ClaimPendingJobs(limit int) ([]models.RetryJob, error)
	//UpdateStatus(id uuid.UUID, status models.JobStatus) error
	IncrementRetry(id uuid.UUID, lastError string, nextRetryAt time.Time) error
	MarkSucceeded(id uuid.UUID) error
	MarkDead(id uuid.UUID, lastError string) error
}

// this is a struct which does not belong to JobRepository interface
// it is defining a db object which is an instance of gorm.db object, its just like insert into query. this gorm.db is like insert into db
type postgresJobRepo struct {
	db *gorm.DB
}

func (r *postgresJobRepo) Create(job *models.RetryJob) error {
	return r.db.Create(job).Error
}

func (r *postgresJobRepo) GetById(id uuid.UUID) (*models.RetryJob, error) {
	var job models.RetryJob
	err := r.db.Where("id=?", id).First(&job).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *postgresJobRepo) ClaimPendingJobs(limit int) ([]models.RetryJob, error) {
	var jobs []models.RetryJob
	err := r.db.Transaction(func(tx *gorm.DB) error {
		err := tx.
			Set("gorm:query_option", "FOR UPDATE SKIP LOCKED").
			Where("status=? AND next_retry_at <= ?", models.StatusPending, time.Now()).
			Order("next_retry_at ASC").
			Limit(limit).
			Find(&jobs).Error

		if err != nil {
			return err
		}

		if len(jobs) == 0 {
			return nil
		}

		//let's say there are 3 jobs, make will create an empty slice [empty, empty, empty]
		ids := make([]uuid.UUID, len(jobs))
		for i, job := range jobs {
			ids[i] = job.ID
		}

		return tx.Model(&models.RetryJob{}).Where("id in ?", ids).Update("status", models.StatusProcessing).Error
	})

	return jobs, err
}

// func (r *postgresJobRepo) UpdateStatus(id uuid.UUID, status models.JobStatus) error {
// 	if status != models.StatusSucceeded && status != models.StatusDead {
// 		return fmt.Errorf("UpdateStatus only accepts succeeded or dead, got: %s", status)
// 	}
// 	return r.db.Model(&models.RetryJob{}).Where("id = ?", id).Update("status", status).Error
// }

func (r *postgresJobRepo) IncrementRetry(id uuid.UUID, lastError string, nextRetryAt time.Time) error {
	return r.db.
		Model(&models.RetryJob{}).
		Where("id=? and status=?", id, models.StatusProcessing).
		Updates(map[string]interface{}{
			"retry_count":   gorm.Expr("retry_count + 1"),
			"last_error":    lastError,
			"next_retry_at": nextRetryAt,
			"status":        models.StatusPending,
		}).Error
}

func (r *postgresJobRepo) MarkSucceeded(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Model(&models.RetryJob{}).Where("id=? and status=?", id, models.StatusProcessing).Update("status", models.StatusSucceeded).Error
	})
}

func (r *postgresJobRepo) MarkDead(id uuid.UUID, lastError string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Model(&models.RetryJob{}).Where("id=? and status=?", id, models.StatusProcessing).Updates(map[string]interface{}{
			"status":     models.StatusDead,
			"last_error": lastError,
		}).Error
	})
}

// we always use this convention while creating a method:
// func [function name](parameters datatype/object) [return value] {}
func NewJobRepository(db *gorm.DB) JobRepository {
	return &postgresJobRepo{
		db: db,
	}
}

/* This whole code translates to this node js snippet:
class JobRepository {
	async create(job) {
		throw new Error("Method not implemented");
	}
}

class PostgresJobRepo extends JobRepository {
	constructor(db) {
		super();
		this.db = db;
	}

	async create(job) {
		return await this.db.create(job);
	}
}

function NewJobRepository(db) {
    return new PostgresJobRepo(db);
}
*/
