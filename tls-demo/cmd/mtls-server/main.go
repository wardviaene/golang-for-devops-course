package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func index(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "it's working")
}

func showCommonName(w http.ResponseWriter, req *http.Request) {
	var commonName string
	if req.TLS != nil && len(req.TLS.VerifiedChains) > 0 && len(req.TLS.VerifiedChains[0]) > 0 {
		commonName = req.TLS.VerifiedChains[0][0].Subject.CommonName
	}
	fmt.Fprintf(w, "Your common name: %s", commonName)
}

func main() {
	caBytes, err := ioutil.ReadFile("ca.crt")
	if err != nil {
		log.Fatal(err)
	}
	ca := x509.NewCertPool()
	if !ca.AppendCertsFromPEM(caBytes) {
		log.Fatal("ca.cert not valid")
	}

	http.HandleFunc("/", index)
	http.HandleFunc("/common-name", showCommonName)
	server := http.Server{
		Addr: ":443",
		TLSConfig: &tls.Config{
			ClientAuth: tls.RequireAndVerifyClientCert,
			ClientCAs:  ca,
			MinVersion: tls.VersionTLS13,
		},
	}
	err = server.ListenAndServeTLS("server.crt", "server.key")
	if err != nil {
		log.Fatal("ListenAndServeTLS error: ", err)
	}
}
