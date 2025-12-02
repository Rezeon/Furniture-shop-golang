package main

import (
	"fmt"
	"log"
	"os"

	"go-be/database"
	"go-be/models"
	"go-be/route"
	"go-be/utils"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables.")
	}

	// Setup logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	logger.Info("Starting server...")

	// Connect Database
	database.ConnectDB()
	utils.InitRedis()
	defer utils.RedisClient.Close()
	err := database.DB.AutoMigrate(
		&models.User{},
		&models.Address{},
		&models.Cart{},
		&models.CartItem{},
		&models.Category{},
		&models.Product{},
		&models.Order{},
	)
	if err != nil {
		logger.Fatal("Failed to migrate database", zap.Error(err))
	}
	logger.Info("Database connected and migrated successfully",
		zap.Strings("tables", []string{
			"user", "address", "cart", "cartitem", "category", "product", "order",
		}),
	)

	// Setup Gin router
	r := route.SetupRoute()

	// Run server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	serverAddr := fmt.Sprintf("0.0.0.0:%s", port)
	logger.Info("Server running", zap.String("url", "http://localhost:"+port))
	if err := r.Run(serverAddr); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
