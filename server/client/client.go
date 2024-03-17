package client

import (
	"github.com/aaaaayushh/ot_editor/server/operation"
	"github.com/gorilla/websocket"
	"sync"
)

const MAX_BUFFER_SIZE = 100

type Client struct {
	Conn *websocket.Conn
	Id   string //unique client id assigned by the server
	//document string //name of the document client wants to work on
	//cursor   int    //current cursor position of the client
	HistoryBuffer []operation.Operation
	SequenceNum   int
	Mu            sync.Mutex
}

func (c *Client) AddToHistoryBuffer(op *operation.Operation) {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	// Add the new operation to the history buffer
	c.HistoryBuffer = append(c.HistoryBuffer, *op)

	// If the history buffer size exceeds the limit, remove the oldest operation
	if len(c.HistoryBuffer) == MAX_BUFFER_SIZE {
		c.HistoryBuffer = c.HistoryBuffer[1:]
	}
	c.SequenceNum++
}
