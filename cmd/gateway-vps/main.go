package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/karma-234/gateway-ps/internal/handler/health"
	"github.com/karma-234/gateway-ps/internal/iso"
	"github.com/karma-234/gateway-ps/internal/metrics"
)

func main() {
	fmt.Println("Hello, World!")
	go func() {

		srv, err := iso.NewServer(":8080")
		defer srv.Close()
		if err != nil {
			log.Fatalf("failed to create server: %v", err)
		}
		// if err := srv.Start(); err != nil {
		// 	log.Fatal(err)
		// }
		log.Println("ISO 8583 Gateway listening on :8080 (Simple TCP)")
	}()

	http.HandleFunc("/health", health.Handler)
	http.Handle("/metrics", metrics.Handler())

	server := &http.Server{Addr: ":8081", TLSConfig: &tls.Config{
		MinVersion: tls.VersionTLS12,
	}}

	log.Println("Health check at: http://localhost:8081/health")
	go func() {
		log.Println("HTTP server listening on :8081")
		if err := server.ListenAndServeTLS("/certs/server.crt", "/certs/server.key"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
	server.Close()
	log.Println("Server stopped")
}
