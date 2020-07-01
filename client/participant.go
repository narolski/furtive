package main

import (
	"math/big"
)

type Participant struct {
	index int
	x *big.Int
	g *big.Int
	q *big.Int
	p *big.Int
	GYi *big.Int
	v *big.Int
}

func NewParticipant(g, q, p *big.Int, index int) *Participant {
	return &Participant{
	    index: index,
	    x: getRandom(q),
	    g: g,
	    q : q,
	    p : p,
	}
}

func (participant *Participant) GetGXi() *big.Int {
	return big.NewInt(0).Exp(participant.g, participant.x, participant.p)
}

func (participant *Participant) ComputeGYi(listOfGXi []*big.Int, sizeOfList int) {
	lowerIndexes, higherIndexes := big.NewInt(1), big.NewInt(1)
	for i := 0; i < participant.index; i++ {
		lowerIndexes.Mod(lowerIndexes.Mul(lowerIndexes, listOfGXi[i]), participant.p)
	}
	for i := participant.index + 1; i < sizeOfList; i++ {
		higherIndexes.Mod(higherIndexes.Mul(higherIndexes, listOfGXi[i]), participant.p)
	}
	participant.GYi = lowerIndexes.Mod(lowerIndexes.Mul(lowerIndexes, higherIndexes.ModInverse(higherIndexes, participant.p)), participant.p)
}

func (participant *Participant) GetVoteVeto() *big.Int {
	c := getRandom(participant.q)
	for c.Cmp(participant.x) == 0 {
		c = getRandom(participant.q)
	}
	participant.x = c
	return c.Exp(participant.GYi, c, participant.p)
}

func (participant *Participant) GetVoteNoVeto() *big.Int {
	return big.NewInt(0).Exp(participant.GYi, participant.x, participant.p)
}

func (participant *Participant) isVeto(listOfVotes []*big.Int, sizeOfList int) bool {
	result := big.NewInt(1)
	for i := 0; i < sizeOfList; i++ {
		result.Mod(result.Mul(result, listOfVotes[i]), participant.p)
	}
	return result.Cmp(big.NewInt(1)) != 0
}

func (participant *Participant) GetVToProofOne() *big.Int {
	participant.v = getRandom(participant.q)
	return big.NewInt(0).Exp(participant.g, participant.v, participant.p)
}

func (participant *Participant) GetRToProof(c *big.Int) *big.Int {
	result := big.NewInt(0);
	return result.Mod(result.Sub(participant.v, big.NewInt(1).Mul(participant.x, c)), participant.q)
}

func (participant *Participant) GetVToProofTwo() *big.Int {
	participant.v = getRandom(participant.q)
	return big.NewInt(0).Exp(participant.GYi, participant.v, participant.p)
}
