package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"math/big"
)

type questionResponse struct {
	Question string `json:"question"`
}

func main() {
	caCert, _ := ioutil.ReadFile("certs/ca.crt")
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cert, _ := tls.LoadX509KeyPair("certs/client1/client.crt", "certs/client1/client.key")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{cert},
			},
		},
	}

	r, err := client.Get("https://127.0.0.1:9443/question")
	if err != nil {
		panic("Cannot get question")
	}
	defer r.Body.Close()

	var resp questionResponse
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	bodyString := string(buf)
	fmt.Println("Body:", bodyString)

	err = json.Unmarshal(buf, &resp)
	if err != nil {
		panic(err)
	}

	fmt.Println(resp)

	// quadratic residue 
	// 7727 = 2 * 3863 + 1
	// p = 2 * q + 1
	// 3 - generator (g)
	g, q, p := big.NewInt(3), big.NewInt(3863), big.NewInt(7727)
	p0 := NewParticipant(g, q, p, 0)
	x0 := p0.BroadcastGXi()

	p1 := NewParticipant(g, q, p, 1)
	x1 := p1.BroadcastGXi()

	x := []*big.Int{x0, x1}

	p0.ComputeGYi(x, 2)
	// v0 := p0.VoteNoVeto()
	v0 := p0.VoteVeto()

	p1.ComputeGYi(x, 2)
	v1 := p1.VoteNoVeto()
	// v1 := p1.VoteVeto()

	fmt.Println(isVeto([]*big.Int{v0, v1}, 2, p))
}
