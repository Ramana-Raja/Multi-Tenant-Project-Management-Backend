package models

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func getEnv(key, def string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return def
}

func ConnnectDatabase() {
	host := getEnv("DB_HOST", "localhost")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "password")
	dbname := getEnv("DB_NAME", "gindb")
	port := getEnv("DB_PORT", "5432")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port,
	)

	for i := 0; i < 10; i++ {
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			DB = db
			break
		}

		log.Println("database not ready yet... retrying in 3s")
		time.Sleep(3 * time.Second)
	}

	if DB == nil {
		log.Fatal("could not connect to database after retries")
	}

	if err := DB.AutoMigrate(
		&User{},
		&Workspace{},
		&WorkspaceMember{},
		&Project{},
		&Task{},
		&AuditLog{},
	); err != nil {
		log.Fatal("migration failed:", err)
	}
	fmt.Println(" database connected & migrations applied successfully")
}
