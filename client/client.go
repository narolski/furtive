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

const firstRoundMessageID = "roundOne"
const secondRoundMessageID = "roundTwo"
const startFirstProofMessageID = "startProofOne"
const continueFirstProofMessageID = "continueProofOne"
const firstProofMessageID = "proofOne"
const startSecondProofMessageID = "startProofTwo"
const continueSecondProofMessageID = "continueProofTwo"
const secondProofMessageID = "proofTwo"
const generatorForVoteMessageID = "generator"

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
			fc.StartProofOne()
		case firstProofMessageID:
			var value *Value
			if err := json.Unmarshal(contents, &value); err != nil {
				log.Fatalln("Error when reading JSON:", err)
				return
			}
			fc.ContinueProofOne(value)
		case firstRoundMessageID:
			var values *Values
			if err := json.Unmarshal(contents, &values); err != nil {
				log.Fatalln("Error when reading JSON:", err)
				return
			}
			fc.AfterRoundOne(values)
			fc.RoundTwo()
			fc.StartProofTwo()
		case secondProofMessageID:
			var value *Value
			if err := json.Unmarshal(contents, &value); err != nil {
				log.Fatalln("Error when reading JSON:", err)
				return
			}
			fc.ContinueProofTwo(value)
		case secondRoundMessageID:
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
	fc.SendMessage(&Message{
		Type: firstRoundMessageID, 
		Contents: &Value{
			Number: participant.GetGXi(),
		}})
	fc.Participant = participant
}

func (fc *FurtiveClient) StartProofOne() {
	fc.SendMessage(&Message{
		Type: startFirstProofMessageID, 
		Contents: &Value{
			Number: fc.Participant.GetVToProofOne(),
		}})
}

func (fc *FurtiveClient) ContinueProofOne(value *Value) {
	fc.SendMessage(&Message{
		Type: continueFirstProofMessageID, 
		Contents: &Value{
			Number: fc.Participant.GetRToProof(value.Number),
		}})
}

func (fc *FurtiveClient) AfterRoundOne(values *Values) {
	fc.Participant.ComputeGYi(values.Numbers, values.Length)
	fc.SendMessage(&Message{
		Type: generatorForVoteMessageID, 
		Contents: &Value{
			Number: fc.Participant.GYi,
		}})
}

func (fc *FurtiveClient) RoundTwo() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Do you want to veto? (Y/N): ")
	text, _ := reader.ReadString('\n')
	var vote *big.Int
	switch text {
	case "Y\n":
		vote = fc.Participant.GetVoteVeto()
	case "N\n":
		vote = fc.Participant.GetVoteNoVeto()
	default:
		vote = fc.Participant.GetVoteNoVeto()
	}
	fc.SendMessage(&Message{
		Type: secondRoundMessageID, 
		Contents: &Value{
			Number: vote,
		}})
}

func (fc *FurtiveClient) StartProofTwo() {
	fc.SendMessage(&Message{
		Type: startSecondProofMessageID, 
		Contents: &Value{
			Number: fc.Participant.GetVToProofTwo(),
		}})
}

func (fc *FurtiveClient) ContinueProofTwo(value *Value) {
	fc.SendMessage(&Message{
		Type: continueSecondProofMessageID, 
		Contents: &Value{
			Number: fc.Participant.GetRToProof(value.Number),
		}})
}

func (fc *FurtiveClient) CheckResult(values *Values) {
	if fc.Participant.isVeto(values.Numbers, values.Length) {
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
