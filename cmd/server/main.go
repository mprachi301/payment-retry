package main

import (
	"log"

	"github.com/mprachi301/payment-retry/config"

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

}
