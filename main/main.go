package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	certFile, keyFile, port string
)

func main() {
	flag.StringVar(&certFile, "certFile", "server.pem", "File containing the x509 Certificate for HTTPS")
	flag.StringVar(&keyFile, "keyFile", "server-key.pem", "File containing the x509 private key for the given certificate")
	flag.StringVar(&port, "port", "8443", "Port to listen")

	flag.Parse()

	certs, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("Error loading key pair: %v", err)
	}

	server := &http.Server{
		Addr: fmt.Sprintf(":%v", port),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{certs},
		},
	}

	// Define server  handler
	handler := AdmissionHandler{}
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", handler.handler)
	server.Handler = mux

	go func() {
		log.Printf("Listening on port %v", port)
		if err := server.ListenAndServeTLS("", ""); err != nil {
			log.Fatalf("Failed to listen and serve webhook server: %v", err)
		}
	}()

	// Listen to the shutdown signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Printf("Shutting down webserver")
	server.Shutdown(context.Background())
}
