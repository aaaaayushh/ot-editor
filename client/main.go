package main

import (
	"fmt"
	"github.com/aaaaayushh/ot_editor/server/operation"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
	"github.com/sergi/go-diff/diffmatchpatch"
	"log"
	"net/http"
	"os"
)

type model struct {
	err      error
	textArea textarea.Model
	oldText  string
}
type errMsg error

// Init can return a Cmd that could perform some initial I/O
func (m model) Init() tea.Cmd {
	return nil
}

type msgMoveCursor int
type msgInsertChar rune

func calculateDelta(oldText, newText string) (inserted string, deleted string) {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(oldText, newText, false)

	for _, diff := range diffs {
		if diff.Type == diffmatchpatch.DiffInsert {
			inserted += diff.Text
		} else if diff.Type == diffmatchpatch.DiffDelete {
			deleted += diff.Text
		}
	}

	return inserted, deleted
}

// Initialize the application state
func initialModel() model {
	ti := textarea.New()
	ti.Placeholder = "Type something.."
	ti.Focus()
	ti.Prompt = "| "
	ti.CharLimit = -1
	ti.SetWidth(50)
	ti.SetHeight(10)
	return model{
		textArea: ti,
		err:      nil,
		oldText:  "",
	}

}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if m.textArea.Focused() {
				m.textArea.Blur()
			}
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyCtrlS: // When Cmd+S is pressed
			// Get the current text
			currentText := m.textArea.Value()

			// Calculate the inserted and deleted text
			insertedText, deletedText := calculateDelta(m.oldText, currentText)

			// Create a new operation with the inserted text as the content
			op := operation.Operation{
				Type:     operation.Insert,
				Position: len(m.oldText),
				Content:  insertedText,
				ID: &operation.OperationID{
					ClientID:    "client1", // Replace with actual client ID
					SequenceNum: 0,         // Replace with actual sequence number
				},
			}

			// Send the operation to the server over the WebSocket connection
			sendOperation(op)

			// Create a new operation with the deleted text as the content
			op = operation.Operation{
				Type:     operation.Delete,
				Position: len(m.oldText),
				Content:  deletedText,
				ID: &operation.OperationID{
					ClientID:    "client1", // Replace with actual client ID
					SequenceNum: 0,         // Replace with actual sequence number
				},
			}

			// Send the operation to the server over the WebSocket connection
			sendOperation(op)

			// Update the old text
			m.oldText = currentText
		default:
			if !m.textArea.Focused() {
				cmd = m.textArea.Focus()
				cmds = append(cmds, cmd)
			}
		}
	case errMsg:
		m.err = msg
		return m, nil

	}

	m.textArea, cmd = m.textArea.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View function to render the UI
func (m model) View() string {
	return fmt.Sprintf("Shuru kare?\n%s", m.textArea.View())
}

var conn *websocket.Conn

func main() {
	// Initialize the Bubble Tea program with the initial model, update function, and view function
	p := tea.NewProgram(initialModel())
	clientId := os.Args[1]
	//check clientId is a valid string
	if clientId == "" {
		fmt.Println("Client ID cannot be empty")
		os.Exit(1)
	}
	// make connection to server

	connectToServer(clientId)
	defer conn.Close()

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

func sendOperation(op operation.Operation) {
	// Format the operation as a string
	opString := fmt.Sprintf("%d:%d:%s", op.Type, op.Position, op.Content)
	fmt.Println("Sending operation:", opString)
	err := conn.WriteMessage(websocket.TextMessage, []byte(opString))
	if err != nil {
		log.Printf("Error sending operation: %v", err)
	}
}

func connectToServer(clientId string) {
	serverAddr := "ws://localhost:8080/ws"
	req, err := http.NewRequest("GET", serverAddr, nil)
	if err != nil {
		fmt.Println("Error creating web socket request:", err)
	}
	req.Header.Set("clientId", clientId)
	conn, _, err = websocket.DefaultDialer.Dial(serverAddr, req.Header)
	if err != nil {
		fmt.Println("Error establishing websocket connection:", err)
	}

}

//func main() {
//	if len(os.Args) < 2 {
//		fmt.Println("Usage: go run client.go <client_id>")
//		return
//	}
//	serverAddr := "ws://localhost:8080/ws"
//
//	req, err := http.NewRequest("GET", serverAddr, nil)
//	if err != nil {
//		log.Fatal("Error creating web socket request:", err)
//	}
//
//	clientId := os.Args[1]
//	req.Header.Set("clientId", clientId)
//	conn, _, err := websocket.DefaultDialer.Dial(serverAddr, req.Header)
//	if err != nil {
//		log.Fatal("Error establishing websocket connection:", err)
//	}
//	defer conn.Close()
//
//	go readMessages(conn)
//	sendMessages(conn)
//}
//
//func readMessages(conn *websocket.Conn) {
//	for {
//		_, message, err := conn.ReadMessage()
//		if err != nil {
//			log.Println("Error reading message:", err)
//			return
//		}
//		fmt.Printf("\nReceived: %s\n", message)
//	}
//}
//
//func sendMessages(conn *websocket.Conn) {
//	reader := bufio.NewReader(os.Stdin)
//
//	for {
//		fmt.Print("> ")
//		message, err := reader.ReadString('\n')
//		if err != nil {
//			log.Println("Error reading message:", err)
//			return
//		}
//
//		// Remove the newline character from the input
//		message = strings.TrimSuffix(message, "\n")
//
//		// Check if the user wants to disconnect
//		if message == "disconnect" {
//			err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
//			if err != nil {
//				log.Println("Error sending disconnect message:", err)
//				return
//			}
//			log.Println("Disconnected from server")
//			return
//		}
//
//		err = conn.WriteMessage(websocket.TextMessage, []byte(message))
//		if err != nil {
//			log.Println("Error sending message:", err)
//			return
//		}
//	}
//}
