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
