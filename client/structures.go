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
