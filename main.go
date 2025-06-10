package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"

	"dorm-chat-api/config"
	"dorm-chat-api/routes"
)

func main() {
	// Загрузка переменных окружения из .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Инициализация базы данных
	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Автоматическая миграция моделей
	if err := config.AutoMigrate(db); err != nil {
		log.Fatalf("DB migration failed: %v", err)
	}

	// Создание Fiber-приложения
	app := fiber.New()

	// Логирование запросов
	app.Use(logger.New())

	// Настройка CORS
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "https://dormmate-mobile.onrender.com"
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	// Настройка маршрутов
	routes.SetupRoutes(app, db)

	// Получение порта из переменных окружения или использование порта по умолчанию
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Запуск сервера
	log.Fatal(app.Listen(":" + port))
}
