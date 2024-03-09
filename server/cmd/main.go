package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// state to be stored on server
/*
	Document state : The true state of each document created.
	Operation queue : All the pending operations that the server is yet to transform and acknowledge.
	Client sessions : A list of all clients that are connected to the server, and the document they are
					working on.
	Client cursors : TBD
	Conflict resolution metadata : TBD
*/

type Client struct {
	conn     *websocket.Conn
	id       string //unique client id assigned by the server
	document string //name of the document client wants to work on
	cursor   int    //current cursor position of the client
}

var (
	clients     = make(map[string]*Client)
	clientsLock sync.Mutex //mutex for synchronizing access to clients map
)

/*
The http.ResponseWriter is an interface provided by Go's net/http package.
It is used to construct an HTTP response to be sent back to the client.
In the context of WebSocket handling, the ResponseWriter is used to perform the WebSocket handshake and upgrade the HTTP
connection to a WebSocket connection.

The http.Request represents an HTTP request received by the server.
It contains information about the HTTP request, including headers, URL, query parameters, and other request details.
In the context of WebSocket handling, the Request object is used to access information about the incoming HTTP request,
such as headers and query parameters.
*/
func handleWebSocket(w http.ResponseWriter, r *http.Request) {

	// Upgrade HTTP connection to web socket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to websocket:", err)
		return
	}
	defer conn.Close()

	clientId := r.Header.Get("clientId")
	if clientId == "" {
		log.Println("Client ID not provided")
	}

	client := &Client{
		conn:   conn,
		id:     clientId,
		cursor: 0,
	}

	clientsLock.Lock()
	clients[clientId] = client
	clientsLock.Unlock()

	//go handleClientMessages(client)
	for {
		// Read message from client
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from client %s: %s", client.id, err)
			break
		}

		// Log received message
		log.Printf("Received message from client %s: %s", client.id, message)

		// Broadcast message to other clients
		broadcastMessage(client.id, message)
	}
}

func handleClientMessages(client *Client) {
	for {
		// Read message from client
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from client %s: %s", client.id, err)
			break
		}

		// Log received message
		log.Printf("Received message from client %s: %s", client.id, message)

		// Broadcast message to other clients
		broadcastMessage(client.id, message)
	}

	// Remove client from clients map when connection is closed
	//clientsLock.Lock()
	//delete(clients, client.id)
	//clientsLock.Unlock()

	//log.Printf("Client disconnected: %s", client.id)
}

func broadcastMessage(senderId string, message []byte) {
	for id, client := range clients {
		if id != senderId {
			if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Println("Error broadcasting message:", err)
			}
		}
	}
}

func main() {
	// entry point for server
	http.HandleFunc("/ws", handleWebSocket)
	//http.HandleFunc("/doc/create", handleCreateDocument)
	log.Println("Starting WebSocket server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Error starting WebSocket server:", err)
	}
}
