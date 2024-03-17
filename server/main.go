package main

import (
	"github.com/aaaaayushh/ot_editor/server/server"
	"log"
	"net/http"
)

func main() {
	server := server.NewServer()
	http.HandleFunc("/ws", server.HandleWebSocket)
	log.Println("Starting server on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
