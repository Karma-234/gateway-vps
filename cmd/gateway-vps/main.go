package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/karma-234/gateway-ps/internal/iso"
)

func main() {
	fmt.Println("Hello, World!")
	server, err := iso.NewServer(":8080")
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}
	log.Println("ISO 8583 Gateway listening on :8080 (Simple TCP)")
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
	server.Close()
	log.Println("Server stopped")
}
