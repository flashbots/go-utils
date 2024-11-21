package main

//
// This example demonstrates how to create a TLS certificate and key and serve it on a port.
//
// The certificate can be required by curl like this:
//
//   curl --cacert cert.pem https://localhost:4433
//

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	utils_tls "github.com/flashbots/go-utils/tls"
)

// Configuration
const listenAddr = ":4433"
const certPath = "cert.pem"

func main() {
	cert, key, err := utils_tls.GenerateTLS(time.Hour*24*265, []string{"localhost"})
	if err != nil {
		panic(err)
	}
	fmt.Println("Generated TLS certificate and key:")
	fmt.Println(string(cert))

	// write cert to file
	err = os.WriteFile(certPath, cert, 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("Wrote certificate to", certPath)

	certificate, err := tls.X509KeyPair(cert, key)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// write certificate to response
		_, _ = w.Write(cert)
	})

	srv := &http.Server{
		Addr:              listenAddr,
		Handler:           mux,
		ReadHeaderTimeout: time.Second,
		TLSConfig: &tls.Config{
			Certificates:             []tls.Certificate{certificate},
			MinVersion:               tls.VersionTLS13,
			PreferServerCipherSuites: true,
		},
	}

	fmt.Println("Starting HTTPS server", "addr", listenAddr)
	if err := srv.ListenAndServeTLS("", ""); err != nil {
		panic(err)
	}
}
