package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

}
