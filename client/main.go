package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

type model struct {
	err      error
	textArea textarea.Model
}
type errMsg error

// Init can return a Cmd that could perform some initial I/O
func (m model) Init() tea.Cmd {
	return nil
}

type msgMoveCursor int
type msgInsertChar rune

// Initialize the application state
func initialModel() model {
	ti := textarea.New()
	ti.Placeholder = "Type something.."
	ti.Focus()

	return model{
		textArea: ti,
		err:      nil,
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
	return fmt.Sprintf("Shuru kare?\n\n\n\n%s", m.textArea.View())
}

func main() {
	// Initialize the Bubble Tea program with the initial model, update function, and view function
	p := tea.NewProgram(initialModel())

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
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
