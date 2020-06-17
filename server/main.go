package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func main() {
	caCert, err := ioutil.ReadFile("certs/ca.crt")
	if err != nil {
		panic("Cannot read CA")
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}
	tlsConfig.BuildNameToCertificate()

	r := mux.NewRouter()
	r.HandleFunc("/question", QuestionHandler)

	server := &http.Server{
		Handler:   r,
		Addr:      ":9443",
		TLSConfig: tlsConfig,
	}
	log.Info("Listening on: ", server.Addr)
	if err := server.ListenAndServeTLS("certs/server.crt", "certs/server.key"); err != nil {
		panic(err)
	}
}
