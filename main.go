package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"context"

	"github.com/gorilla/websocket"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
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
	} else if clientType == "receiver" {
		receiverConn = conn
	}
	log.Printf("%s connected\n", clientType)


	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		if clientType == "sender" {
			fmt.Printf("sender %d", messageType)
				fmt.Println(msg)
				err = receiverConn.WriteMessage(messageType, msg)
				if err != nil {
					log.Println("Error sending image to receiver:", err)
					break
				} else {
					if messageType == websocket.BinaryMessage {
						fileName := time.Now().Format("2006-01-02_15:04:05")
						if err := uploadImageToFirebase(msg, fileName); err != nil {
							log.Println("Error uploading image:", err)
							break
						}
					}
				}
			
		} else if clientType == "receiver" {
			fmt.Printf("reciever %d", messageType)
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

func uploadImageToFirebase(imageData []byte, fileName string) error {
	ctx := context.Background()

	opt := option.WithCredentialsFile("serviceAccountKey.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return fmt.Errorf("error initializing app: %v", err)
	}

	client, err := app.Storage(ctx)
	if err != nil {
		return fmt.Errorf("error getting Storage client: %v", err)
	}

	bucketName := "smart-doorbell-9bd31.appspot.com"
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return fmt.Errorf("error getting bucket: %v", err)
	}

	writer := bucket.Object(fileName).NewWriter(ctx)
	if _, err = writer.Write(imageData); err != nil {
		writer.Close()
		return fmt.Errorf("error writing to bucket: %v", err)
	}
	if err = writer.Close(); err != nil {
		return fmt.Errorf("error closing writer: %v", err)
	}

	fmt.Printf("Image uploaded to Firebase Storage: %s\n", fileName)
	return nil
}

func main() {

	http.HandleFunc("/ws", wsHandler)
	http.Handle("/", http.FileServer(http.Dir("./test-frontend"))) // Just for testing with mock-up frontend
	log.Fatal(http.ListenAndServe(":8080", nil))
}
