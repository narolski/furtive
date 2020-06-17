package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	log "github.com/sirupsen/logrus"
)

var questions = []string{"been run by a truck", "shoplifted", "lied to a police officer", "found true love"}

type questionResponse struct {
	Question string `json:"question"`
}

func QuestionHandler(w http.ResponseWriter, r *http.Request) {
	resp := &questionResponse{
		Question: fmt.Sprintf("Have you ever %s?", questions[rand.Intn(len(questions))]),
	}
	json.NewEncoder(w).Encode(resp)
	log.Info("Returned question: ", resp.Question)
}
