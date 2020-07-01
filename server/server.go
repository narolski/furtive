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
			log.Fatalf("Error when reading JSON:", err)
			return
		}

		switch msg.Type {
		case firstRoundMessageID:
			A.Set(value.Number)
			fs.handleMessageFromClient(fs.firstRoundChan, value.Number, clientID)
		case startFirstProofMessageID:
			V.Set(value.Number)
			Cx := fs.sendZeroKnowledgeProofChallenge(value.Number, ws, firstProofMessageID)
			C.Set(Cx)
		case continueFirstProofMessageID:
			fs.validateZeroKnowledgeProofResponse(value.Number, V, A, V, C, group.Generator, ws)
		case generatorForVoteMessageID:
			gYi.Set(value.Number)
			// TODO
			// receive gYi eg.
			// &Message{
			// 	Type: startSecondProofMessageID,
			// 	Contents: &Value{
			// 		Number: gYi,
			// 	}}
		case secondRoundMessageID:
			// receive Y
			Y.Set(value.Number)
			fs.handleMessageFromClient(fs.secondRoundChan, value.Number, clientID)
		case startSecondProofMessageID:
			V2.Set(value.Number)
			Cx := fs.sendZeroKnowledgeProofChallenge(value.Number, ws, secondProofMessageID)
			C2.Set(Cx)
			// TODO
			// receive V eg.
			// &Message{
			// 	Type: startSecondProofMessageID,
			// 	Contents: &Value{
			// 		Number: V,
			// 	}}
			// then get random big.Int number c from [0, 2^t-1] (say t=160)
			// and send it to client eg.
			// &Message{
			// 	Type: secondProofMessageID,
			// 	Contents: &Value{
			// 		Number: c,
			// 	}}
		case continueSecondProofMessageID:

			fmt.Println("Resx:", V2, Y, C2, gYi)

			fs.validateZeroKnowledgeProofResponse2(value.Number, V2, Y, V2, C2, gYi, ws)
			// TODO
			// receive r eg.
			// &Message{
			// 	Type: continueSecondProofMessageID,
			// 	Contents: &Value{
			// 		Number: r,
			// 	}}
			// then check
			// 1) Y is a valid public key
			//    use isValueFromRoundCorrect (end of file), where Y is from startSecondProofMessageID message
			//    and divisor, bigPrimary - from Group
			// 2) V = gYi^r * Y^c mod p
			//    use isValueFromProofCorrect (end of file), where
			//    A = Y is from secondRoundMessageID message
			//    V, C - startSecondProofMessageID message
			//    r - this message
			//    divisor, bigPrimary - from Group
			//    generator = gYi from generatorForVoteMessageID message
			// if sth is incorrect/false - stop game
		default:
			log.Errorf("Invalid message type '%s': %+v", msg.Type, msg)
		}
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

func (fs *FurtiveServer) validateZeroKnowledgeProofResponse(r, startFirstProofMsgVal, firstRoundMessageVal, startFirstProofMessageV, startFirstProofMessageC, generator *big.Int, ws *websocket.Conn) {
	if ok := fs.isValueFromRoundCorrect(startFirstProofMsgVal, group.Divisor, group.BigPrimary); !ok {
		panic("ZKP Error: Value from the first round is not correct")
	}
	if ok := fs.isValueFromProofCorrect(r, firstRoundMessageVal, startFirstProofMessageV, startFirstProofMessageC, generator, group.Divisor, group.BigPrimary); !ok {
		panic("ZKP Error: Proof failed")
	}
}

func (fs *FurtiveServer) validateZeroKnowledgeProofResponse2(r, startSecondProofMsgVal, secondRoundMsgVal, startSecondProofMessageV, startSecondProofMessageC, generator *big.Int, ws *websocket.Conn) {
	if ok := fs.isValueFromRoundCorrect(startSecondProofMsgVal, group.Divisor, group.BigPrimary); !ok {
		panic("ZKP Error: Value from the first round is not correct")
	}
	if ok := fs.isValueFromProofCorrect(r, secondRoundMsgVal, startSecondProofMessageV, startSecondProofMessageC, generator, group.Divisor, group.BigPrimary); !ok {
		panic("ZKP Error: Proof failed")
	}
}

// TODO
// receive r eg.
// &Message{
// 	Type: continueSecondProofMessageID,
// 	Contents: &Value{
// 		Number: r,
// 	}}
// then check
// 1) Y is a valid public key
//    use isValueFromRoundCorrect (end of file), where Y is from startSecondProofMessageID message
//    and divisor, bigPrimary - from Group
// 2) V = gYi^r * Y^c mod p
//    use isValueFromProofCorrect (end of file), where
//    A = Y is from secondRoundMessageID message
//    V, C - startSecondProofMessageID message
//    r - this message
//    divisor, bigPrimary - from Group
//    generator = gYi from generatorForVoteMessageID message
// if sth is incorrect/false - stop game

func (fs *FurtiveServer) isValueFromRoundCorrect(A *big.Int, divisor *big.Int, bigPrimary *big.Int) bool {
	if big.NewInt(0).Cmp(A) != -1 || A.Cmp(divisor) != -1 || big.NewInt(1).Cmp(big.NewInt(1).Exp(A, bigPrimary, divisor)) != 0 {
		return false
	}
	return true
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
