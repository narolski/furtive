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
}

func NewParticipant(g, q, p *big.Int, index int) *Participant {
	participant := new(Participant)
	participant.index = index
	participant.x = getRandom(q)
	participant.g = g
	participant.q = q
	participant.p = p
	return participant
}

func (participant *Participant) BroadcastGXi() *big.Int {
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

func (participant *Participant) VoteVeto() *big.Int {
	c := getRandom(participant.q)
	for c.Cmp(participant.x) == 0 {
		c = getRandom(participant.q)
	}
	return c.Exp(participant.GYi, c, participant.p)
}

func (participant *Participant) VoteNoVeto() *big.Int {
	return big.NewInt(0).Exp(participant.GYi, participant.x, participant.p)
}