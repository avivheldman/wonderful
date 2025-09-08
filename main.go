package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var responseComplete = make(chan bool)

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
			"instructions": "You are a helpful assistant. When asked to multiply numbers, use the multiply function.",
			"tools": []map[string]interface{}{
				{
					"type":        "function",
					"name":        "multiply",
					"description": "Multiplies two numbers and returns the result",
					"parameters": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"a": map[string]interface{}{
								"type":        "number",
								"description": "First number to multiply",
							},
							"b": map[string]interface{}{
								"type":        "number",
								"description": "Second number to multiply",
							},
						},
						"required": []string{"a", "b"},
					},
				},
			},
			"tool_choice": "auto",
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

func listenForResponseMessages(conn *websocket.Conn, done chan bool) {
	for {
		select {
		case <-done:
			// Received shutdown signal
			log.Println("Shutting down message listener...")
			return
		default:
			// Check for new messages
			var message map[string]interface{}
			err := conn.ReadJSON(&message)
			if err != nil {
				log.Println("Error reading message:", err)
				return
			}
			eventType := message["type"].(string)
			switch eventType {
			case "response.text.delta":
				delta := message["delta"].(string)
				fmt.Print(delta)
			case "response.text.done":
				fmt.Println()
				responseComplete <- true
			case "response.function_call_arguments.done":
				handleFunctionCall(conn, message)
			}
		}
	}
}

func multiply(a float64, b float64) float64 {
	return a * b
}

func handleFunctionCall(conn *websocket.Conn, message map[string]interface{}) {
	callID := message["call_id"].(string)
	functionName := message["name"].(string)
	argumentsJSON := message["arguments"].(string)

	fmt.Printf("\n[Calling function: %s with args: %s]\n", functionName, argumentsJSON)

	var args map[string]float64
	if err := json.Unmarshal([]byte(argumentsJSON), &args); err != nil {
		fmt.Printf("Error parsing arguments: %v\n", err)
		return
	}

	var result float64
	if functionName == "multiply" {
		result = multiply(args["a"], args["b"])
		fmt.Printf("Result: %.2f\n", result)
	}

	functionOutput := map[string]interface{}{
		"type": "conversation.item.create",
		"item": map[string]interface{}{
			"type":    "function_call_output",
			"call_id": callID,
			"output":  fmt.Sprintf("%.2f", result),
		},
	}

	if err := conn.WriteJSON(functionOutput); err != nil {
		fmt.Printf("Error sending function result: %v\n", err)
		return
	}

	responseMsg := map[string]interface{}{
		"type": "response.create",
	}
	conn.WriteJSON(responseMsg)
}

func main() {
	conn, err := connectToOpenAI()
	if err != nil {
		log.Fatal("Error connecting to OpenAI:", err)
	}
	defer conn.Close()

	setSessionconfig(conn)
	log.Println("Connected to OpenAI WebSocket")
	done := make(chan bool)
	go listenForResponseMessages(conn, done)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\nYou: ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		if input == "exit" || input == "quit" {
			break
		}
		if input == "" {
			continue
		}

		fmt.Print("AI: ")
		sendMessage(conn, input)
		<-responseComplete
	}

	// Clean shutdown
	fmt.Println("\n👋 Goodbye!")
	done <- true // Signal goroutine to stop
}
