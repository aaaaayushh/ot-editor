package main

// TODO go install github.com/dkorunic/betteralign/cmd/betteralign@latest
import (
	"bufio"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run client.go <client_id>")
		return
	}
	serverAddr := "ws://localhost:8080/ws"

	req, err := http.NewRequest("GET", serverAddr, nil)
	if err != nil {
		log.Fatal("Error creating web socket request:", err)
	}

	clientId := os.Args[1]
	req.Header.Set("clientId", clientId)
	conn, _, err := websocket.DefaultDialer.Dial(serverAddr, req.Header)
	if err != nil {
		log.Fatal("Error establishing websocket connection:", err)
	}
	defer conn.Close()

	go readMessages(conn)
	sendMessages(conn)
}

func readMessages(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			return
		}
		fmt.Printf("\nReceived: %s\n", message)
	}
}

func sendMessages(conn *websocket.Conn) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading message:", err)
			return
		}

		// Remove the newline character from the input
		message = strings.TrimSuffix(message, "\n")

		// Check if the user wants to disconnect
		if message == "disconnect" {
			err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Error sending disconnect message:", err)
				return
			}
			log.Println("Disconnected from server")
			return
		}

		err = conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Println("Error sending message:", err)
			return
		}
	}
}
