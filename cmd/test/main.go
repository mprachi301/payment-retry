package main

// import (
// 	"fmt"
// 	"log"
// 	"time"

// 	"github.com/mprachi301/payment-retry/config"
// 	"github.com/mprachi301/payment-retry/internal/models"
// 	"github.com/mprachi301/payment-retry/internal/repository"
// 	"github.com/mprachi301/payment-retry/internal/service"

// 	"gorm.io/driver/postgres"
// 	"gorm.io/gorm"
// )

// func main() {
// 	cnf := config.Load()

// 	db, err := gorm.Open(postgres.Open(cnf.DB_DSN), &gorm.Config{})
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	log.Println("Database connected", db != nil)

// 	repo := repository.NewJobRepository(db)

// 	//create testcase
// 	fmt.Println("---Test creating---")
// 	job := &models.RetryJob{
// 		PaymentID:   "pay_1",
// 		Amount:      99.99,
// 		Currency:    "INR",
// 		Status:      models.StatusPending,
// 		MaxRetries:  3,
// 		NextRetryAt: time.Now().Add(10 * time.Second),
// 	}

// 	err = repo.Create(job)
// 	if err != nil {
// 		log.Fatal("Create Failed", err)
// 	}
// 	fmt.Println("Job created with ID: ", job.ID)

// 	//test GetById
// 	fmt.Println("---Testing GetIdBy---")
// 	fetchedJob, err := repo.GetById(job.ID)
// 	if err != nil {
// 		log.Fatal("Fetching failed", err)
// 	}
// 	fmt.Println("fetched job: \n", fetchedJob)

// 	//test ClaimPendingJobs
// 	fmt.Println("---Testing ClaimPendingJobs---")
// 	db.Model(&models.RetryJob{}).Where("id=?", job.ID).Update("next_retry_at", time.Now().Add(-1*time.Second))
// 	jobs, err := repo.ClaimPendingJobs(10)
// 	if err != nil {
// 		log.Fatal("ClaimPendingJobs failed", err)
// 	}
// 	fmt.Println("Claimed job count: ", len(jobs))
// 	for _, i := range jobs {
// 		fmt.Printf("Job ID: %s, status: %s\n", i.ID, i.Status)
// 	}
// 	// verify what's actually in DB
// 	for _, j := range jobs {
// 		fetched, _ := repo.GetById(j.ID)
// 		fmt.Println("DB status:", fetched.Status) // will show processing
// 	}

// 	//test IncrementRetry
// 	fmt.Println("---Testing IncrementRetry---")
// 	err = repo.IncrementRetry(job.ID, "connection timeout", time.Now().Add(30*time.Second))
// 	if err != nil {
// 		log.Fatal("IncrementRetry Failed")
// 	}
// 	updated, _ := repo.GetById(job.ID)
// 	fmt.Printf("After IncrementRetry — status: %s, retry_count: %d, last_error: %s\n",
// 		updated.Status, updated.RetryCount, updated.LastError)

// 	// 5. Test UpdateStatus
// 	fmt.Println("--- Testing UpdateStatus ---")
// 	// first claim it again so status is processing
// 	db.Model(&models.RetryJob{}).Where("id = ?", job.ID).Updates(map[string]any{
// 		"status":        models.StatusProcessing,
// 		"next_retry_at": time.Now().Add(-1 * time.Second),
// 	})
// 	err = repo.UpdateStatus(job.ID, models.StatusSucceeded)
// 	if err != nil {
// 		log.Fatal("UpdateStatus failed: ", err)
// 	}
// 	fmt.Println("After UpdateStatus — status:", updated.Status)

// 	fmt.Println("--- All tests passed ---")

// 	fmt.Println("--- Testing MockPaymentProcessor ---")
// 	processor := service.NewMockPaymentProcessor(0.9)

// 	for i := 0; i < 5; i++ {
// 		err := processor.ProcessPayment(*fetchedJob)
// 		if err != nil {
// 			fmt.Printf("Payment attempt %d failed: %s\n", i+1, err)
// 		} else {
// 			fmt.Printf("Payment attempt %d succeeded\n", i+1)
// 		}
// 	}
// }
