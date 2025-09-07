package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

func connectToOpenAI() (*websocket.Conn, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	var (
		OpenAIKey = os.Getenv("OPENAI_KEY")
		OpenAIURL = os.Getenv("OPENAI_URL")
	)
	header := http.Header{
		"Authorization": []string{"Bearer " + OpenAIKey},
		"OpenAI-Beta":   []string{"realtime=v1"},
	}
	conn, _, err := websocket.DefaultDialer.Dial(OpenAIURL, header)
	return conn, err
}

func setSessionconfig(conn *websocket.Conn) error {
	message := map[string]interface{}{
		"type": "session.update",
		"session": map[string]interface{}{
			"modalities":   []string{"text"},
			"instructions": "Be extra nice today!",
		},
	}
	return conn.WriteJSON(message)

}

func sendMessage(conn *websocket.Conn, content string) error {
	message := map[string]interface{}{
		"type": "conversation.item.create",
		"item": map[string]interface{}{
			"type": "message",
			"role": "user",
			"content": []map[string]interface{}{
				{"type": "input_text", "text": content}},
		},
	}
	if err := conn.WriteJSON(message); err != nil {
		return err
	}
	responseMsg := map[string]interface{}{
		"type": "response.create",
	}
	return conn.WriteJSON(responseMsg)
}

func listenForResponseMessages(conn *websocket.Conn) {
	for {
		var message map[string]interface{}
		err := conn.ReadJSON(&message)
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}
		eventType := message["type"].(string)
		switch eventType {
		case "response.text.delta":
			delta := message["delta"].(string)
			fmt.Print(delta)
		case "response.text.done":
			fmt.Println()
		}
	}
}
func main() {
	conn, err := connectToOpenAI()
	if err != nil {
		log.Fatal("Error connecting to OpenAI:", err)
	}
	defer conn.Close()
	setSessionconfig(conn)
	log.Println("Connected to OpenAI WebSocket")
	sendMessage(conn, "Hi.")
	listenForResponseMessages(conn)
}
