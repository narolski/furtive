package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"math/big"
	"encoding/json"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var questions = []string{"been run by a truck", "shoplifted", "lied to a police officer", "found true love"}

type clientVal struct{}

type Message struct {
	Type string `json:"type"`
	Contents interface{} `json:"contents"`
}

type VotingData struct {
	Id int `json:"id"`
	Question string `json:"question"`
	Generator *big.Int `json:"generator"`
	BigPrimary *big.Int `json:"bigPrimary"`
	Divisor *big.Int `json:"divisor"`
}

type Value struct {
	Number *big.Int `json:"number"`
}

type Values struct {
	Numbers []*big.Int `json:"numbers"`
	Length int `json:"length"`
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
	
	// TODO: Necessary to add id for clients (now 0) -> FROM 0, not 1
	data := &VotingData{0, fs.question, big.NewInt(3), big.NewInt(3863), big.NewInt(7727)}
	fs.sendMessageToClient(&Message{"votingData", data}, ws)

	// Pass messages from clients to other clients
	// TODO: Now only one client
	for {
		var contents json.RawMessage
		msg := &Message{
			Contents: &contents,
		}
		if err := ws.ReadJSON(&msg); err != nil {
			log.Error("Error when reading message from client:", err)
			fs.removeClient(ws)
			ws.Close()
			return
		}
		log.Info("New message type: ", msg.Type)

		switch msg.Type {
		case "roundOne":
			var value *Value
			if err := json.Unmarshal(contents, &value); err != nil {
				log.Fatalln("Error when reading JSON:", err)
				return
			}
			fs.sendMessageToClient(&Message{"roundOne", &Values{[]*big.Int{value.Number}, 1}}, ws)
		case "roundTwo":
			var value *Value
			if err := json.Unmarshal(contents, &value); err != nil {
				log.Fatalln("Error when reading JSON:", err)
				return
			}
			fs.sendMessageToClient(&Message{"roundTwo", &Values{[]*big.Int{value.Number}, 1}}, ws)
		default:
			log.Fatalf("unknown message type: %q", msg.Type)
		}
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
