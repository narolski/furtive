package main

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type Message struct {
	Contents string `json:"contents"`
}

type FurtiveClient struct {
	ws *websocket.Conn
}

func NewFurtiveClient(ws *websocket.Conn) *FurtiveClient {
	return &FurtiveClient{
		ws: ws,
	}
}

func (fc *FurtiveClient) ReadMessages() {
	for {
		var msg *Message
		if err := fc.ws.ReadJSON(&msg); err != nil {
			fc.ws.Close()
			log.Fatalln("Error when reading JSON:", err)
			return
		}
		fmt.Println("New message:", msg.Contents)
		fc.SendMessage(&Message{fmt.Sprintf("READMSG:%s", msg.Contents)})
	}
}

func (fc *FurtiveClient) SendMessage(message *Message) {
	if err := fc.ws.WriteJSON(&message); err != nil {
		log.Println("Error when sending message:", err, fc.ws)
		fc.ws.Close()
	}
}
