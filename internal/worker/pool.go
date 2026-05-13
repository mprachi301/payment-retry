package worker

import (
	"log"
	"sync"

	"github.com/mprachi301/payment-retry/internal/models"
	"github.com/mprachi301/payment-retry/internal/service"
)

type WorkerPool struct {
	workerCount int
	jobQueue    chan models.RetryJob
	quit        chan struct{}
	wg          sync.WaitGroup
	service     *service.JobService
	processor   service.PaymentProcess
}

func NewWorkerPool(n int, service *service.JobService, processor service.PaymentProcess) *WorkerPool {
	return &WorkerPool{
		workerCount: n,
		jobQueue:    make(chan models.RetryJob, n*2),
		quit:        make(chan struct{}),
		service:     service,
		processor:   processor,
	}
}

func (p *WorkerPool) Start() {
	for i := 1; i <= p.workerCount; i++ {
		p.wg.Add(1)
		go p.runWorker(i)
	}
	log.Printf("worker pool started with %d workers\n", p.workerCount)
}

func (p *WorkerPool) runWorker(id int) {
	defer p.wg.Done()

	defer func() {
		if r := recover(); r != nil {
			log.Printf("worker %d recovered from panic: %v\n", id, r)
		}
	}()

	log.Printf("worker %d started\n", id)

	for {
		select {
		case job, ok := <-p.jobQueue:
			if !ok {
				log.Printf("worker %d stopping — job queue closed\n", id)
				return
			}
			p.processJob(id, job)

		case <-p.quit:
			// stop signal received
			log.Printf("worker %d exiting — quit signal received\n", id)
			return
		}
	}
}

func (p *WorkerPool) processJob(workerID int, job models.RetryJob) {
	log.Printf("worker %d processing job %s (attempt %d/%d)\n", workerID, job.ID, job.RetryCount+1, job.MaxRetries)

	// call the payment processor
	err := p.processor.ProcessPayment(job)

	if err == nil {
		// success — mark as succeeded
		log.Printf("worker %d job %s succeeded\n", workerID, job.ID)
		if updateErr := p.service.MarkSucceeded(job.ID); updateErr != nil {
			log.Printf("worker %d failed to mark job %s succeeded: %v\n", workerID, job.ID, updateErr)
		}
		return
	}
	log.Printf("worker %d job %s failed: %v\n", workerID, job.ID, err)

	if job.RetryCount+1 >= job.MaxRetries {
		log.Printf("worker %d job %s exhausted all retries\n", workerID, job.ID)
		if updateErr := p.service.MarkDead(job.ID, err.Error()); updateErr != nil {
			log.Printf("worker %d failed to mark job %s dead: %v\n", workerID, job.ID, updateErr)
		}
		return
	}

	nextRetry := service.CalculateNextRetry(job.RetryCount + 1)
	log.Printf("worker %d job %s will retry at %v\n", workerID, job.ID, nextRetry)
	if updateErr := p.service.IncrementRetry(job.ID, err.Error(), nextRetry); updateErr != nil {
		log.Printf("worker %d failed to increment retry for job %s: %v\n", workerID, job.ID, updateErr)
	}
}

func (p *WorkerPool) Stop() {
	log.Println("stopping worker pool...")
	close(p.quit)
	p.wg.Wait() // wait for all workers to finish
	log.Println("worker pool stopped")
}

func (p *WorkerPool) Submit(job models.RetryJob) {
	select {
	case p.jobQueue <- job:
		log.Printf("job %s submitted to queue\n", job.ID)
	default:
		log.Printf("job %s dropped — queue is full\n", job.ID)
	}
}
