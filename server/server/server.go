package server

import (
	"fmt"
	"github.com/aaaaayushh/ot_editor/server/client"
	"github.com/aaaaayushh/ot_editor/server/operation"
	"github.com/aaaaayushh/ot_editor/server/ot"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

/*
	Document state : The true state of each document created.
	Operation queue : All the pending operations that the server is yet to transform and acknowledge.
	Client sessions : A list of all clients that are connected to the server, and the document they are
					working on.
	Client cursors : TBD
	Conflict resolution metadata : TBD
*/

// TODO: figure out multiple documents, maybe using a DB
var sharedDocument string //central copy of document stored on server

var clients = make(map[string]*client.Client)

type Server struct{}

func NewServer() *Server {
	return &Server{}
}

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
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {

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

	currClient := &client.Client{
		Conn:          conn,
		Id:            clientId,
		HistoryBuffer: make([]operation.Operation, 0),
		SequenceNum:   0,
		//cursor: 0,
	}

	//clientsLock.Lock()
	clients[clientId] = currClient
	//clientsLock.Unlock()

	for {
		// Read message from client
		_, message, err := currClient.Conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from client %s: %s", currClient.Id, err)
			break
		}

		op, err := parseOperation(message, clientId)
		if err != nil {
			log.Println(err)
			continue
		}

		transformedOp := ot.TransformOperation(op, currClient)
		applyOperation(&sharedDocument, transformedOp)
		broadcastOperation(clientId, transformedOp)

		currClient.Mu.Lock()
		currClient.HistoryBuffer = append(currClient.HistoryBuffer, *transformedOp)
		currClient.SequenceNum++
		currClient.Mu.Unlock()
	}
}

// function that takes in the message sent by a client, and outputs the corresponding operation.
// input is expected in this format ->  "operationType:position:content"
func parseOperation(message []byte, clientID string) (*operation.Operation, error) {
	parts := strings.Split(string(message), ":")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid message format")
	}

	opType, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid operation type: %s", err)
	}

	position, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid position: %s", err)
	}

	content := parts[2]

	op := &operation.Operation{
		Type:     operation.OperationType(opType),
		Position: position,
		Content:  content,
		ID: &operation.OperationID{
			ClientID:    clientID,
			SequenceNum: getNextSequenceNum(clientID),
		},
	}

	return op, nil
}

func getNextSequenceNum(clientID string) int {
	currClient, ok := clients[clientID]
	if !ok {
		return 0
	}

	currClient.Mu.Lock()
	defer currClient.Mu.Unlock()

	sequenceNum := currClient.SequenceNum
	currClient.SequenceNum++
	return sequenceNum
}

func applyOperation(document *string, op *operation.Operation) {
	if op.Type == operation.Insert {
		*document = (*document)[:op.Position] + op.Content + (*document)[op.Position:]
	} else {
		*document = (*document)[:op.Position] + (*document)[op.Position+len(op.Content):]
	}
}

func broadcastOperation(senderID string, op *operation.Operation) {
	for id, currClient := range clients {
		if id != senderID {
		}
		currClient.Mu.Lock()
		transformedOp := ot.TransformOperation(op, currClient)
		err := currClient.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%d:%d:%s", transformedOp.Type, transformedOp.Position, transformedOp.Content)))
		if err != nil {
			log.Println(err)
			currClient.Conn.Close()
			delete(clients, id)
		}
		currClient.HistoryBuffer = append(currClient.HistoryBuffer, *transformedOp)
		currClient.Mu.Unlock()
	}
}
