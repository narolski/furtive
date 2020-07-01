package main

import "math/big"

type Message struct {
	Type     string      `json:"type"`
	Contents interface{} `json:"contents"`
}

type VotingData struct {
	Id         int      `json:"id"`
	Question   string   `json:"question"`
	Generator  *big.Int `json:"generator"`
	BigPrimary *big.Int `json:"bigPrimary"`
	Divisor    *big.Int `json:"divisor"`
}

type Value struct {
	Number *big.Int `json:"number"`
}

type Values struct {
	Numbers []*big.Int `json:"numbers"`
	Length  int        `json:"length"`
}

type Group struct {
	Generator  *big.Int
	BigPrimary *big.Int
	Divisor    *big.Int
}

type ClientValue struct {
	ClientID int
	Value    *big.Int
}

var questions = []string{"been run by a truck", "shoplifted", "lied to a police officer", "voted for PiS", "sworn a  revenge for undisclosed course requirements", "owned a VW Golf IV 1.9 TDi", "been with a girl named Jessica", "lived in Brochów"}

var group = &Group{
	Generator:  big.NewInt(3),
	BigPrimary: big.NewInt(3863),
	Divisor:    big.NewInt(7727),
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
const disconnectedMesageID = "disconnected"

const t = 160
