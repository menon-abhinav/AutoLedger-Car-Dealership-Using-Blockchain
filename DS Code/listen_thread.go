package main

import (
	"encoding/gob"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Relaxes the same-origin policy.
}

// Using a sync.Map to safely handle concurrent access to the peers map.
var peers sync.Map

func handleConnections(w http.ResponseWriter, r *http.Request) {
	bc := NewBlockchain()
	defer bc.db.Close()

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer ws.Close()

	// Register new peer
	peers.Store(ws, true)

	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				log.Printf("Error: Unexpected close error: %v", err)
			} else {
				log.Printf("Read error: %v", err)
			}
			break
		}

		// Process received message
		handleMessage(msg, ws, bc)
	}

	// Clean up after the loop ends
	peers.Delete(ws)
	log.Println("Disconnected:", ws.RemoteAddr())
}

func main() {
	gob.Register(&VehicleRegistration{})
	gob.Register(&VehicleSale{})
	gob.Register(&LoanContract{})
	gob.Register(&genesis{})
	gob.Register(&Block{})
	http.HandleFunc("/ws", handleConnections)
	log.Println("WebSocket server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe error:", err)
	}
}
