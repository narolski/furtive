package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const serverAddress = ":9200"
const clientsMaxAmount = 3

func main() {
	caCert, err := ioutil.ReadFile("certs/ca.crt")
	if err != nil {
		panic(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}
	tlsConfig.BuildNameToCertificate()

	s := NewFurtiveServer(clientsMaxAmount)
	r := mux.NewRouter()
	r.HandleFunc("/ws", s.connectionHandler)

	server := &http.Server{
		Handler:   r,
		Addr:      serverAddress,
		TLSConfig: tlsConfig,
	}
	log.Info("Listening on: ", server.Addr)
	if err := server.ListenAndServeTLS("certs/server.crt", "certs/server.key"); err != nil {
		panic(err)
	}
}
