package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type FurtiveServer struct {
	clientsMaxAmount int
	question         string
	clients          map[*websocket.Conn]int
	firstRoundChan   chan *ClientValue
	secondRoundChan  chan *ClientValue
	broadcastChan    chan *Message
	mu               sync.Mutex
}

func NewFurtiveServer(clientsMaxAmount int) *FurtiveServer {
	rand.Seed(time.Now().UnixNano())
	fs := &FurtiveServer{
		clientsMaxAmount: clientsMaxAmount,
		question:         fmt.Sprintf("Have you ever %s?", questions[rand.Intn(len(questions))]),
		clients:          make(map[*websocket.Conn]int),
		firstRoundChan:   make(chan *ClientValue),
		secondRoundChan:  make(chan *ClientValue),
		broadcastChan:    make(chan *Message),
	}
	var wg sync.WaitGroup
	wg.Add(3)
	go fs.broadcastMessageToClients(&wg)
	go fs.createGroupedResponse(&wg, fs.firstRoundChan, firstRoundMessageID)
	go fs.createGroupedResponse(&wg, fs.secondRoundChan, secondRoundMessageID)
	wg.Wait()
	return fs
}

func (fs *FurtiveServer) connectionHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := &websocket.Upgrader{}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	// Register client internally and send initial message with question
	clientID := fs.registerClient(ws)
	fs.sendMessageToClient(&Message{
		Type: "votingData",
		Contents: &VotingData{
			Id:         clientID,
			Question:   fs.question,
			Generator:  group.Generator,
			BigPrimary: group.BigPrimary,
			Divisor:    group.Divisor,
		},
	}, ws)

	// Zero-knowledge-proof variables
	A := &big.Int{}
	V := &big.Int{}
	C := &big.Int{}
	
	gYi := &big.Int{}
	Y := &big.Int{}
	V2 := &big.Int{}
	C2 := &big.Int{}

	// Pass messages from clients to other clients
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
		log.Infof("Received message of type '%s' from client '%d': %+v", msg.Type, clientID, msg)

		var value *Value
		if err := json.Unmarshal(contents, &value); err != nil {
			log.Fatalln("Error when reading JSON:", err)
			return
		}

		switch msg.Type {
		case firstRoundMessageID:
			A.Set(value.Number)
			fs.handleMessageFromClient(fs.firstRoundChan, value.Number, clientID)
		case startFirstProofMessageID:
			V.Set(value.Number)
			C.Set(fs.sendZeroKnowledgeProofChallenge(value.Number, ws, firstProofMessageID))
		case continueFirstProofMessageID:
			if ok := fs.isValueFromRoundCorrect(V, group.Divisor, group.BigPrimary); !ok {
				fs.handleZeroKnowledgeProofError(1, 1, clientID, ws)	
				return	
			}
			if ok := fs.isValueFromProofCorrect(value.Number, A, V, C, group.Generator, group.Divisor, group.BigPrimary); !ok {
				fs.handleZeroKnowledgeProofError(1, 2, clientID, ws)
				return
			}
		case generatorForVoteMessageID:
			gYi.Set(value.Number)
		case secondRoundMessageID:
			Y.Set(value.Number)
			fs.handleMessageFromClient(fs.secondRoundChan, value.Number, clientID)
		case startSecondProofMessageID:
			V2.Set(value.Number)
			C2.Set(fs.sendZeroKnowledgeProofChallenge(value.Number, ws, secondProofMessageID))
		case continueSecondProofMessageID:
			if ok := fs.isValueFromRoundCorrect(V2, group.Divisor, group.BigPrimary); !ok {
				fs.handleZeroKnowledgeProofError(2, 1, clientID, ws)
				return
			}
			if ok := fs.isValueFromProofCorrect(value.Number, Y, V2, C2, gYi, group.Divisor, group.BigPrimary); !ok {
				fs.handleZeroKnowledgeProofError(2, 2, clientID, ws)
				return
			}

		default:
			log.Errorf("Invalid message type '%s': %+v", msg.Type, msg)
		}
	}
}

func (fs *FurtiveServer) registerClient(ws *websocket.Conn) int {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	id := len(fs.clients)
	fs.clients[ws] = id
	return id
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

func (fs *FurtiveServer) broadcastMessageToClients(wg *sync.WaitGroup) {
	wg.Done()
	for {
		msg := <-fs.broadcastChan
		for ws, clientID := range fs.clients {
			fs.sendMessageToClient(msg, ws)
			log.Infof("Message %+v sent to client ID %d", msg, clientID)
		}
	}
}

func (fs *FurtiveServer) createGroupedResponse(wg *sync.WaitGroup, messages chan *ClientValue, messageType string) {
	wg.Done()
	log.Info("Started createGroupedResponse for message type: ", messageType)
	clientVals := make(map[int]*big.Int)
	for {
		if len(clientVals) == fs.clientsMaxAmount {
			break
		}
		clientVal := <-messages
		clientVals[clientVal.ClientID] = clientVal.Value
		log.Infof("Added message from client %d to message type %s queue", clientVal.ClientID, messageType)
	}
	values := make([]*big.Int, fs.clientsMaxAmount)
	for clientID, value := range clientVals {
		values[clientID] = value
	}
	fs.broadcastChan <- &Message{
		Type: messageType,
		Contents: &Values{
			Numbers: values,
			Length:  len(values),
		},
	}
}

func (fs *FurtiveServer) handleMessageFromClient(target chan *ClientValue, value *big.Int, clientID int) {
	target <- &ClientValue{
		ClientID: clientID,
		Value:    value,
	}
}

func (fs *FurtiveServer) sendZeroKnowledgeProofChallenge(v *big.Int, ws *websocket.Conn, messageType string) *big.Int {
	c := new(big.Int)
	c.Exp(big.NewInt(2), big.NewInt(t), nil).Sub(c, big.NewInt(1))
	fs.sendMessageToClient(&Message{
		Type: messageType,
		Contents: &Value{
			Number: c,
		},
	}, ws)
	return c
}

func (fs *FurtiveServer) handleZeroKnowledgeProofError(round, turn, clientID int, ws *websocket.Conn) {
	fs.sendMessageToClient(&Message{
		Type: disconnectedMesageID,
		Contents: fmt.Sprintf("Zero-knowledge-proof turn %d, round %d failed", turn, round),
	}, ws)
	fs.removeClient(ws)
	ws.Close()
	log.Errorf("ZKP%d Error: Value from the %d round is incorrect for client ID '%d'. Client disconnected.", turn, round, clientID)
}

func (fs *FurtiveServer) isValueFromProofCorrect(r *big.Int, A *big.Int, V *big.Int, c *big.Int, generator *big.Int, divisor *big.Int, bigPrimary *big.Int) bool {
	if V.Cmp(
		big.NewInt(1).Mod(
			big.NewInt(1).Mul(
				big.NewInt(1).Exp(generator, r, divisor),
				big.NewInt(1).Exp(A, c, divisor),
			),
			divisor)) != 0 {
		return false
	}
	return true
}

func (fs *FurtiveServer) isValueFromRoundCorrect(A *big.Int, divisor *big.Int, bigPrimary *big.Int) bool {
	if big.NewInt(0).Cmp(A) != -1 || A.Cmp(divisor) != -1 || big.NewInt(1).Cmp(big.NewInt(1).Exp(A, bigPrimary, divisor)) != 0 {
		return false
	}
	return true
}