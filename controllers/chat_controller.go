// controllers/chatController.go
package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"dorm-chat-api/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// DormData описывает один объект dorm из поля "results" в ответе Django.
type DormData struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	FloorsCount uint   `json:"floors_count"`
}

// DormListResponse описывает корневую структуру, возвращаемую Django.
// {
//   "count": 2,
//   "next": null,
//   "previous": null,
//   "results": [ { "id": 4, "floors_count": 1, … }, { … } ]
// }
type DormListResponse struct {
	Count    int        `json:"count"`
	Next     *string    `json:"next"`
	Previous *string    `json:"previous"`
	Results  []DormData `json:"results"`
}

// CreateAllChatsHandler автоматически создаёт чаты для каждого dorm и этажей,
// используя поле "floors_count" из ответа Django.
func CreateAllChatsHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1) Запрос к Django, чтобы получить список dorm-ов
		djangoURL := "http://127.0.0.1:8000/api/v1/dorms/"
		resp, err := http.Get(djangoURL)
		if err != nil {
			fmt.Println("ERROR: не удалось выполнить http.Get до Django:", err)
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error": fmt.Sprintf("не удалось запросить список общежитий у Django: %v", err),
			})
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			fmt.Printf("ERROR: Django вернул статус %d: %s\n", resp.StatusCode, string(bodyBytes))
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error": fmt.Sprintf("Django вернул статус %d: %s", resp.StatusCode, string(bodyBytes)),
			})
		}

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("ERROR: не удалось прочитать тело ответа Django:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "не удалось прочитать ответ Django",
			})
		}

		var dormList DormListResponse
		if err := json.Unmarshal(bodyBytes, &dormList); err != nil {
			fmt.Println("ERROR: JSON.Unmarshal:", err, "raw body:", string(bodyBytes))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("ошибка парсинга JSON от Django: %v", err),
			})
		}

		// 2) Автосоздание недостающих чатов
		for _, dorm := range dormList.Results {
			// 2.1) Чат для всего общежития
			dormChatID := fmt.Sprintf("dorm_%d", dorm.ID)
			dormName := fmt.Sprintf("Общежитие № %d", dorm.ID)

			var existingDorm models.Chat
			err := db.Where("chat_id = ?", dormChatID).First(&existingDorm).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				newDormChat := models.Chat{
					ChatID:    dormChatID,
					Type:      "dorm",
					DormID:    dorm.ID,
					Floor:     0,
					Name:      dormName,
					CreatedAt: time.Now(),
				}
				if errCreate := db.Create(&newDormChat).Error; errCreate != nil {
					fmt.Println("ERROR: не удалось создать dorm-чат в БД:", errCreate)
				}
			} else if err != nil {
				// При любом другом err просто логируем и продолжаем
				fmt.Println("ERROR: при попытке найти существующий dorm-чат:", err)
			}

			// 2.2) Чаты для каждого этажа (используем FloorsCount)
			for floor := uint(1); floor <= dorm.FloorsCount; floor++ {
				floorChatID := fmt.Sprintf("floor_%d_%d", dorm.ID, floor)
				floorName := fmt.Sprintf("Этаж %d, общежитие № %d", floor, dorm.ID)

				var existingFloor models.Chat
				errFloor := db.Where("chat_id = ?", floorChatID).First(&existingFloor).Error
				if errors.Is(errFloor, gorm.ErrRecordNotFound) {
					newFloorChat := models.Chat{
						ChatID:    floorChatID,
						Type:      "floor",
						DormID:    dorm.ID,
						Floor:     floor,
						Name:      floorName,
						CreatedAt: time.Now(),
					}
					if errCreate := db.Create(&newFloorChat).Error; errCreate != nil {
						fmt.Println("ERROR: не удалось создать floor-чат в БД:", errCreate)
					}
				} else if errFloor != nil {
					// При любом другом errFloor просто логируем и продолжаем
					fmt.Println("ERROR: при попытке найти существующий floor-чат:", errFloor)
				}
			}
		}

		// 3) Возвращаем полный список чатов из БД (после автосоздания)
		var chats []models.Chat
		if err := db.Find(&chats).Error; err != nil {
			fmt.Println("ERROR: не удалось выполнить db.Find(&chats):", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "не удалось получить список чатов из БД",
			})
		}
		return c.JSON(chats)
	}
}

// CleanupChatsHandler удаляет чаты, если dorm был удалён или этаж превысил floors_count в Django.
func CleanupChatsHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1) Запрос к Django
		djangoURL := "http://127.0.0.1:8000/api/v1/dorms/"
		resp, err := http.Get(djangoURL)
		if err != nil {
			fmt.Println("ERROR: не удалось запросить dorms у Django:", err)
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error": fmt.Sprintf("не удалось запросить dorms у Django: %v", err),
			})
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			fmt.Printf("ERROR: Django вернул %d: %s\n", resp.StatusCode, string(bodyBytes))
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error": fmt.Sprintf("Django вернул %d: %s", resp.StatusCode, string(bodyBytes)),
			})
		}

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("ERROR: не удалось прочитать тело ответа Django:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "не удалось прочитать ответ Django",
			})
		}

		var dormList DormListResponse
		if err := json.Unmarshal(bodyBytes, &dormList); err != nil {
			fmt.Println("ERROR: json.Unmarshal при cleanup:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("ошибка парсинга JSON: %v", err),
			})
		}

		// 2) Собираем карту dormID → floors_count
		dormMap := make(map[uint]uint, len(dormList.Results))
		for _, d := range dormList.Results {
			dormMap[d.ID] = d.FloorsCount
		}

		// 3) Берём все чаты из БД
		var allChats []models.Chat
		if err := db.Find(&allChats).Error; err != nil {
			fmt.Println("ERROR: db.Find(&allChats) при cleanup:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "не удалось получить чаты из БД",
			})
		}

		// 4) Удаляем устаревшие
		var deletedCount int
		for _, chat := range allChats {
			switch chat.Type {
			case "dorm":
				if _, exists := dormMap[chat.DormID]; !exists {
					if err := db.Delete(&chat).Error; err == nil {
						deletedCount++
					} else {
						fmt.Println("ERROR: удаление dorm-чата:", err)
					}
				}
			case "floor":
				if maxFloor, exists := dormMap[chat.DormID]; !exists || chat.Floor > maxFloor {
					if err := db.Delete(&chat).Error; err == nil {
						deletedCount++
					} else {
						fmt.Println("ERROR: удаление floor-чата:", err)
					}
				}
			default:
				if err := db.Delete(&chat).Error; err == nil {
					deletedCount++
				} else {
					fmt.Println("ERROR: удаление чата с неизвестным Type:", err)
				}
			}
		}

		return c.JSON(fiber.Map{
			"deleted_chats": deletedCount,
		})
	}
}

// GetChatMessages и SendMessage оставляем без изменений

func GetChatMessages(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		chatID := c.Params("chat_id")
		var messages []models.Message
		db.Where("chat_id = ?", chatID).Order("created_at asc").Find(&messages)
		return c.JSON(messages)
	}
}

func SendMessage(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		chatID := c.Params("chat_id")

		// 1) Считаем тело запроса включая sender_type
		var req struct {
			Content    string `json:"content"`
			SenderType string `json:"sender_type"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "bad request"})
		}

		// 2) Получаем userID из JWT (дальше можно использовать, если нужно)
		userID := c.Locals("userID")
		if userID == nil {
			return c.Status(401).JSON(fiber.Map{"error": "user_id not found in token"})
		}
		senderID := fmt.Sprintf("%v", userID)

		// 3) Теперь мы просто доверяем client-side полю req.SenderType
		//    (например, "admin" или "student").
		//    Если хотите дополнительную валидацию, можете проверить JWT + /usertype/,
		//    но минимум — нужно записать именно то, что пришло.
		msg := models.Message{
			ChatID:     chatID,
			SenderID:   senderID,
			SenderType: req.SenderType,
			Content:    req.Content,
			CreatedAt:  time.Now(),
		}
		if err := db.Create(&msg).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("не удалось сохранить сообщение: %v", err)})
		}

		return c.JSON(msg)
	}
}
