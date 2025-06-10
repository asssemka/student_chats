package main

import (
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"

	"dorm-chat-api/config"
	"dorm-chat-api/routes"
)

func main() {
	// Загружаем переменные из .env (работает локально, на Render можно пропустить ошибку)
	_ = godotenv.Load()

	// Подключение к базе
	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("❌ Ошибка подключения к базе данных: %v", err)
	}

	// Миграции (если нужны)
	if err := config.AutoMigrate(db); err != nil {
		log.Fatalf("❌ Ошибка миграции: %v", err)
	}

	// Создаём Fiber-приложение
	app := fiber.New()

	// Включаем логгер
	app.Use(logger.New())

	// Читаем разрешённые origins из переменных окружения
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:3000"
	}
	// Fiber требует строки без пробелов!
	allowedOrigins = strings.ReplaceAll(allowedOrigins, " ", "")

	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	// Основные роуты
	routes.SetupRoutes(app, db)

	// Render задаёт порт через переменную PORT
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Go API сервер стартует на :%s", port)
	log.Fatal(app.Listen(":" + port))
}

// package main

// import (
// 	"log"
// 	"os"
// 	"strings"

// 	"github.com/gofiber/fiber/v2"
// 	"github.com/gofiber/fiber/v2/middleware/cors"
// 	"github.com/gofiber/fiber/v2/middleware/logger"
// 	"github.com/joho/godotenv"

// 	"dorm-chat-api/config"
// 	"dorm-chat-api/routes"
// )

// func main() {
// 	// Загрузка переменных окружения из .env
// 	if err := godotenv.Load(); err != nil {
// 		log.Println("⚠️  .env файл не найден, используются системные переменные")
// 	}

// 	// Подключение к БД
// 	db, err := config.InitDB()
// 	if err != nil {
// 		log.Fatalf("❌ Ошибка подключения к базе данных: %v", err)
// 	}

// 	// Миграции
// 	if err := config.AutoMigrate(db); err != nil {
// 		log.Fatalf("❌ Ошибка миграции: %v", err)
// 	}

// 	// Инициализация Fiber
// 	app := fiber.New()

// 	// Логирование
// 	app.Use(logger.New())

// 	// CORS для Flutter Web и React
// 	app.Use(cors.New(cors.Config{
// 		AllowOriginsFunc: func(origin string) bool {
// 			// Пропускаем локальные фронты
// 			allowed := []string{
// 				"http://localhost:3000", // React
// 				"http://127.0.0.1:3000",
// 				"http://localhost:54259", // Flutter Web (примерный порт)
// 				"http://127.0.0.1:5500",
// 			}
// 			for _, o := range allowed {
// 				if origin == o {
// 					return true
// 				}
// 			}
// 			// Допускаем все локальные origins
// 			return strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "http://127.0.0.1")
// 		},
// 		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
// 		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
// 		AllowCredentials: true,
// 	}))

// 	// Роуты
// 	routes.SetupRoutes(app, db)

// 	// Порт сервера
// 	port := os.Getenv("PORT")
// 	if port == "" {
// 		port = "8080"
// 	}

// 	log.Printf("🚀 Сервер Go API работает на http://localhost:%s", port)
// 	log.Fatal(app.Listen(":" + port))
// }
