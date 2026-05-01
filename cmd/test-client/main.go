package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/karma-234/gateway-ps/internal/iso"
	"github.com/karma-234/gateway-ps/internal/pkg"
	"github.com/moov-io/iso8583"
)

func main() {
	// Load client certificate for mTLS
	cert, err := tls.LoadX509KeyPair("certs/client.crt", "certs/client.key")
	if err != nil {
		log.Fatalf("Failed to load client certificate: %v", err)
	}

	// Load CA certificate
	caCert, err := os.ReadFile("certs/server.crt")
	if err != nil {
		log.Fatalf("Failed to load CA certificate: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false,
	}

	// Connect to gateway
	conn, err := tls.Dial("tcp", "localhost:8082", tlsConfig)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	log.Println("✅ Connected to ISO 8583 gateway with TLS + mTLS")

	// Create a sample authorization message
	msg := iso8583.NewMessage(iso.Spec8353)

	// Fill required fields
	msg.MTI("0100")
	msg.Field(2, "4242424242424242") // PAN
	msg.Field(3, "000000")           // Processing Code (Purchase)
	msg.Field(4, "000000001000")     // Amount 10.00
	msg.Field(7, "0430123456")       // Transmission Date
	msg.Field(11, "123456")          // STAN
	msg.Field(14, "2512")            // Expiry Date
	msg.Field(41, "12345678")        // Terminal ID
	msg.Field(42, "000000000000001") // Merchant ID
	msg.Field(49, "840")             // Currency Code (USD)

	// Pack message
	packed, err := msg.Pack()
	if err != nil {
		log.Fatalf("Failed to pack message: %v", err)
	}
	header := make([]byte, 2)
	binary.BigEndian.PutUint16(header, uint16(len(packed)))
	fullMsg := append(header, packed...)
	// Send message
	_, err = conn.Write(fullMsg)
	if err != nil {
		log.Fatalf("Failed to send: %v", err)
	}

	fmt.Printf("📤 Sent ISO 8583 message (MTI 0100):\n%s\n", hex.Dump(fullMsg))

	// Read response transport header (2 bytes)
	respHeader := make([]byte, 2)
	if _, err := io.ReadFull(conn, respHeader); err != nil {
		log.Fatalf("Failed to read response header: %v", err)
	}

	respLen := int(binary.BigEndian.Uint16(respHeader))
	if respLen <= 0 {
		log.Fatalf("Invalid response length: %d", respLen)
	}

	// Read exact response body
	respBody := make([]byte, respLen)
	if _, err := io.ReadFull(conn, respBody); err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	fmt.Printf("Received response frame (header=% x body=%d):\n%s\n", respHeader, respLen, hex.Dump(respBody))

	// Unpack ONLY ISO body (not transport header)
	respMsg := iso8583.NewMessage(iso.Spec8353)
	if err := respMsg.Unpack(respBody); err != nil {
		log.Fatalf("Failed to unpack response: %v", err)
	}

	mti, _ := respMsg.GetMTI()
	log.Printf("Response MTI: %s", mti)
	if code, _ := respMsg.GetString(39); code != "" {
		log.Printf("Response Code (39): %s", code)
	}
	if pan, _ := respMsg.GetString(2); pan != "" {
		log.Printf("This is the masked PAN (field 2): %s", pkg.MaskPAN(pan))
	}

}
