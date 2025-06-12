package main

import (
	"log"
	"os"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"

	"dorm-chat-api/config"
	"dorm-chat-api/routes"
	ws "dorm-chat-api/websocket" // WebSocket-пакет
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := config.AutoMigrate(db); err != nil {
		log.Fatalf("DB migration failed: %v", err)
	}

	app := fiber.New()

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     os.Getenv("ALLOWED_ORIGINS"),
		AllowMethods:     "GET,POST,PUT,DELETE",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	// 🌐 Создаём WebSocket хаб
	hub := ws.NewHub()
	go hub.Run()

	// 📌 Подключаем WebSocket маршрут
	app.Get("/ws/:chat_id", websocket.New(func(c *websocket.Conn) {
		ws.ServeWS(hub, c)
	}))

	// 📦 Остальные HTTP-маршруты
	routes.SetupRoutes(app, db)

	// 🚀 Запуск
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Server running on port " + port)
	log.Fatal(app.Listen(":" + port))
}
