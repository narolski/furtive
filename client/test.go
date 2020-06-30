package main

import (
	"math/big"
	"fmt"
)

func main() {
	generator := big.NewInt(3)
	divisor := big.NewInt(7727)
	bigPrimary := big.NewInt(3863)
	participant := NewParticipant(generator, bigPrimary, divisor, 0)

	// round one with proof
	A := participant.GetGXi()
	V := participant.GetVToProofOne()
	c := big.NewInt(12345678)
	r := participant.GetRToProof(c)
	A1 := isValueFromRoundCorrect(A, divisor, bigPrimary)
	fmt.Println(A1)
	V1 := isValueFromProofCorrect(r, V, A, c, generator, divisor, bigPrimary)
	fmt.Println(V1)

	// round two with proof
	participant.ComputeGYi([]*big.Int{A}, 1)
	Y := participant.GetVoteVeto()
	// Y := participant.GetVoteNoVeto()
	V = participant.GetVToProofTwo()
	r = participant.GetRToProof(c)
	A2 := isValueFromRoundCorrect(Y, divisor, bigPrimary)
	fmt.Println(A2)
	V2 := isValueFromProofCorrect(r, V, Y, c, participant.GYi, divisor, bigPrimary)
	fmt.Println(V2)
}

func isValueFromRoundCorrect(A *big.Int, divisor *big.Int, bigPrimary *big.Int) bool {
	if big.NewInt(0).Cmp(A) != -1 || A.Cmp(divisor) != -1 || big.NewInt(1).Cmp(big.NewInt(1).Exp(A, bigPrimary, divisor)) != 0 {
		return false
	}
	return true
}

func isValueFromProofCorrect(r *big.Int, V *big.Int, A *big.Int, c *big.Int, generator *big.Int, divisor *big.Int, bigPrimary *big.Int) bool {
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
