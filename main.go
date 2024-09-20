package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var senderConn *websocket.Conn
var receiverConn *websocket.Conn

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	clientType := r.URL.Query().Get("type")
	if clientType == "sender" {
		senderConn = conn
		log.Println("Sender connected")
	} else if clientType == "receiver" {
		receiverConn = conn
		log.Println("Receiver connected")
	}

	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		if clientType == "sender" {
			if receiverConn != nil {
				err = receiverConn.WriteMessage(messageType, msg)
				if err != nil {
					log.Println("Error sending image to receiver:", err)
					break
				}
			}
		} else if clientType == "receiver" {
			if senderConn != nil {
				err = senderConn.WriteMessage(messageType, msg)
				if err != nil {
					log.Println("Error sending response to sender:", err)
					break
				}
			}
		}
	}
}

func main() {
	http.HandleFunc("/ws", wsHandler)
	http.Handle("/", http.FileServer(http.Dir("./test-frontend"))) // Just for testing with mock-up frontend
	log.Fatal(http.ListenAndServe(":8080", nil))
}
