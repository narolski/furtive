package main

import (
	"math/big"
	"crypto/rand"
)

func getRandom(q *big.Int) *big.Int {
	x, err := rand.Int(rand.Reader, q)
	if err != nil {
        panic(err)
    }
	return x
}

func isVeto(listOfVotes []*big.Int, sizeOfList int, p *big.Int) bool {
	result := big.NewInt(1)
	for i := 0; i < sizeOfList; i++ {
		result.Mod(result.Mul(result, listOfVotes[i]), p)
	}
	return result.Cmp(big.NewInt(1)) != 0
}