package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"math/big"
	"net/url"
	"fmt"

	"github.com/gorilla/websocket"
)

const URL = "ws://127.0.0.1:9200/ws"
const BUFFER_SIZE = 64

func main() {
	caCert, _ := ioutil.ReadFile("certs/ca.crt")
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cert, _ := tls.LoadX509KeyPair("certs/client1/client.crt", "certs/client1/client.key")

	u, err := url.Parse(URL)
	if err != nil {
		panic(err)
	}

	conn, err := tls.Dial("tcp", u.Host, &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
	})
	if err != nil {
		panic(err)
	}

	ws, _, err := websocket.NewClient(conn, u, http.Header{}, BUFFER_SIZE, BUFFER_SIZE)
	if err != nil {
		panic(err)
	}

	client := NewFurtiveClient(ws)
	client.ReadMessages()

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
