package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	// "math/big"
	"net/url"
	// "fmt"

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
}
