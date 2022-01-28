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
	must_rules, must_not_rules []Rule
)

func main() {
	var certFile, keyFile, port string
	var verbose bool
	var must_rules_filename, must_not_rules_filename string

	flag.StringVar(&certFile, "cert", "server.pem", "File containing the x509 Certificate for HTTPS")
	flag.StringVar(&keyFile, "key", "server-key.pem", "File containing the x509 private key for the given certificate")
	flag.StringVar(&port, "port", "8443", "Port to listen")
	flag.StringVar(&must_rules_filename, "must-rules", "must.rules", "YAML file containing the 'must' rules")
	flag.StringVar(&must_not_rules_filename, "must-not-rules", "must-not.rules", "YAML file containing the 'must-not' rules")
	flag.BoolVar(&verbose, "verbose", false, "Verbosity mode for debugging")

	flag.Parse()

	log_level := log.InfoLevel
	if verbose {
		log_level = log.DebugLevel
	}
	log.SetLevel(log_level)

	log.Debugf("Using cert file %s", certFile)
	log.Debugf("Using key file %s", keyFile)
	certs, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("Error loading key pair: %v", err)
	}

	must_rules = load_rules_file(must_rules_filename)
	must_not_rules = load_rules_file(must_not_rules_filename)

	server := &http.Server{
		Addr: fmt.Sprintf(":%v", port),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{certs},
		},
	}

	// Define server  handler
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", handleAdmissionRequest)
	mux.HandleFunc("/health/liveness", handleLiveness)
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
