package api

import (
	"encoding/json"
	"log"
	"server/internal/db/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

var conn *websocket.Conn

func UIInit() {
	app := fiber.New()

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		log.Println(" UI connected via WebSocket")
		conn = c
		defer ConnectionStop()

		// Hold connection until UI disconnects
		for {
			if _, _, err := c.ReadMessage(); err != nil {
				break
			}
		}
	}))

	log.Println(" WebSocket server running at ws://localhost:8080/ws")
	log.Fatal(app.Listen(":8080"))
}

func SendLog(entry models.Consumer_msg) {
	if conn == nil {
		log.Println(" No active WebSocket connection")
		return
	}

	data, err := json.Marshal(entry)
	if err != nil {
		log.Println(" Failed to marshal log:", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Println(" Failed to send WebSocket message:", err)
	}
}

func ConnectionStop() {
	log.Println(" UI disconnected")
	if conn != nil {
		conn.Close()
		conn = nil
	}
}
