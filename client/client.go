package main

import (
	"bufio"
    "fmt"
    "os"
	"log"
	"math/big"
	"encoding/json"

	"github.com/gorilla/websocket"
)

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

type FurtiveClient struct {
	ws *websocket.Conn
	Participant *Participant
}

func NewFurtiveClient(ws *websocket.Conn) *FurtiveClient {
	return &FurtiveClient{
		ws: ws,
	}
}

func (fc *FurtiveClient) ReadMessages() {
	for {
		var contents json.RawMessage
		msg := &Message{
			Contents: &contents,
		}
		if err := fc.ws.ReadJSON(&msg); err != nil {
			fc.ws.Close()
			log.Fatalln("Error when reading JSON:", err)
			return
		}

		switch msg.Type {
		case "votingData":
			var votingData *VotingData
			if err := json.Unmarshal(contents, &votingData); err != nil {
				log.Fatalln("Error when reading JSON:", err)
				return
			}
			fc.RoundOne(votingData)
		case "roundOne":
			var values *Values
			if err := json.Unmarshal(contents, &values); err != nil {
				log.Fatalln("Error when reading JSON:", err)
				return
			}
			fc.AfterRoundOne(values)
			fc.RoundTwo()
		case "roundTwo":
			var values *Values
			if err := json.Unmarshal(contents, &values); err != nil {
				log.Fatalln("Error when reading JSON:", err)
				return
			}
			fc.CheckResult(values)
		default:
			log.Fatalf("unknown message type: %q", msg.Type)
		}
	}
}

func (fc *FurtiveClient) RoundOne(votingData *VotingData) {
	fmt.Println("Question:", votingData.Question)
	participant := NewParticipant(votingData.Generator, votingData.BigPrimary, votingData.Divisor, votingData.Id)
	gXi := participant.BroadcastGXi()
	value := &Value{gXi}
	msg := &Message{"roundOne", value}
	fc.SendMessage(msg)
	fc.Participant = participant
}

func (fc *FurtiveClient) AfterRoundOne(values *Values) {
	fc.Participant.ComputeGYi(values.Numbers, values.Length)
}

func (fc *FurtiveClient) RoundTwo() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Do you want to veto? (Y/N): ")
	text, _ := reader.ReadString('\n')
	var vote *big.Int
	switch text {
	case "Y":
		vote = fc.Participant.VoteVeto()
	case "N":
		vote = fc.Participant.VoteNoVeto()
	default:
		vote = fc.Participant.VoteNoVeto()
	}
	value := &Value{vote}
	msg := &Message{"roundTwo", value}
	fc.SendMessage(msg)
}

func (fc *FurtiveClient) CheckResult(values *Values) {
	if isVeto(values.Numbers, values.Length, fc.Participant.p) {
		fmt.Println("Veto!")
	} else {
		fmt.Println("No veto")
	}
}

func (fc *FurtiveClient) SendMessage(message *Message) {
	if err := fc.ws.WriteJSON(&message); err != nil {
		log.Println("Error when sending message:", err, fc.ws)
		fc.ws.Close()
	}
}
