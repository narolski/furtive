package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var questions = []string{"been run by a truck", "shoplifted", "lied to a police officer", "found true love"}

type clientVal struct{}

type Message struct {
	Contents string `json:"contents"`
}

type FurtiveServer struct {
	playersAmount int
	question      string
	clients       map[*websocket.Conn]clientVal
	// addClientChan    chan *websocket.Conn
	// removeClientChan chan *websocket.Conn
	broadcastChan chan *Message
	mu            sync.Mutex
}

func NewFurtiveServer(playersAmount int) *FurtiveServer {
	return &FurtiveServer{
		playersAmount: playersAmount,
		question:      fmt.Sprintf("Have you ever %s?", questions[rand.Intn(len(questions))]),
		clients:       make(map[*websocket.Conn]clientVal),

		// registerClientChan: make(chan *websocket.Conn),
		// removeClientChan: make(chan *websocket.Conn),
		broadcastChan: make(chan *Message),
	}
}

func (fs *FurtiveServer) connectionHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	upgrader := &websocket.Upgrader{}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	// Register client and send initial message to client
	fs.registerClient(ws)
	fs.sendMessageToClient(&Message{fs.question}, ws)

	// Pass messages from clients to other clients
	for {
		var msg *Message
		if err := ws.ReadJSON(&msg); err != nil {
			log.Error("Error when reading message from client:", err)
			fs.removeClient(ws)
			ws.Close()
			return
		}
		fs.broadcastChan <- msg
		log.Info("Received message from client: ", msg.Contents)
	}
}

func (fs *FurtiveServer) registerClient(ws *websocket.Conn) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.clients[ws] = clientVal{}
}

func (fs *FurtiveServer) removeClient(ws *websocket.Conn) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	delete(fs.clients, ws)
}

func (fs *FurtiveServer) sendMessageToClient(message *Message, ws *websocket.Conn) {
	if err := ws.WriteJSON(&message); err != nil {
		log.Error("Error when sending message:", err, ws)
		fs.removeClient(ws)
		ws.Close()
	}
}

func (fs *FurtiveServer) broadcastToClients() {
	msg := <-fs.broadcastChan
	log.Info("Sent message to client: ", msg.Contents)
	for ws := range fs.clients {
		fs.sendMessageToClient(msg, ws)
	}
}
