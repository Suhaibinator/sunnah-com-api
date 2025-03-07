package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Suhaibinator/api/go_api"
	"github.com/Suhaibinator/api/go_persistence"
	"github.com/Suhaibinator/api/go_service"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load the environment variables
	godotenv.Load(".env.prod_connect")

	// Database connection parameters
	username := os.Getenv("MYSQL_USER")
	password := os.Getenv("MYSQL_PASSWORD")
	hostname := os.Getenv("MYSQL_HOST")
	port := os.Getenv("MYSQL_PORT")
	dbname := os.Getenv("MYSQL_DATABASE")

	// Build the DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username, password, hostname, port, dbname)

	// Configure GORM logger
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level (Silent, Error, Warn, Info)
			IgnoreRecordNotFoundError: false,       // Log record not found error
			Colorful:                  true,        // Enable color
		},
	)

	// Initialize the database connection with logger
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}

	appPersistence := go_persistence.NewApplicationPersistence(db)
	appService := go_service.NewApplicationService(appPersistence)
	appRouter := go_api.NewApplicationRouter(appService)

	appRouter.RegisterRoutes()
	appRouter.Run(8084)

}
