package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/mprachi301/payment-retry/config"
	"github.com/mprachi301/payment-retry/internal/api"
	"github.com/mprachi301/payment-retry/internal/repository"
	"github.com/mprachi301/payment-retry/internal/service"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cnf := config.Load()

	db, err := gorm.Open(postgres.Open(cnf.DB_DSN), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Database connected", db != nil)

	repo := repository.NewJobRepository(db)
	jobService := service.NewJobService(repo, cnf.MAX_RETRIES)
	handler := api.NewJobHandler(jobService)

	r := gin.Default()
	api.SetupRoutes(r, handler)

	log.Println("server starting on port:", cnf.PORT)
	err = r.Run(":" + cnf.PORT)
	if err != nil {
		log.Fatal("failed to start server: ", err)
	}
}
