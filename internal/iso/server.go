package iso

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/karma-234/gateway-ps/internal/fineract"
	"github.com/karma-234/gateway-ps/internal/pkg"
	"github.com/moov-io/iso8583"
)

type Server struct {
	listener       net.Listener
	fineractClient *fineract.Client
}

func NewServer(addr string) (*Server, error) {
	cert, err := tls.LoadX509KeyPair("/certs/server.crt", "/certs/server.key")
	if err != nil {
		return nil, err
	}

	caCert, err := os.ReadFile("/certs/server.crt")
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	clientCACert, err := os.ReadFile("/certs/client.crt")
	if err != nil {
		return nil, err
	}
	clientCAPool := x509.NewCertPool()
	clientCAPool.AppendCertsFromPEM(clientCACert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    clientCAPool,
		MinVersion:   tls.VersionTLS12,
	}

	ln, err := tls.Listen("tcp", addr, tlsConfig)
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener:       ln,
		fineractClient: fineract.NewClient(),
	}

	go s.acceptLoop()
	log.Printf("🔒 ISO 8583 TLS + mTLS Server listening on %s", addr)
	return s, nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			return
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	if tlsConn, ok := conn.(*tls.Conn); ok {
		if err := tlsConn.Handshake(); err != nil {
			log.Printf("TLS handshake failed from %s: %v", conn.RemoteAddr(), err)
			return
		}
		state := tlsConn.ConnectionState()
		log.Printf("✅ TLS Connection from %s | Client certs: %d", conn.RemoteAddr(), len(state.PeerCertificates))
	}

	log.Printf("📥 New connection from %s", conn.RemoteAddr())

	reader := bufio.NewReader(conn)

	for {
		// Read 2-byte length header (Big Endian)
		header := make([]byte, 2)
		if _, err := io.ReadFull(reader, header); err != nil {
			if err != io.EOF {
				log.Printf("Error reading header from %s: %v", conn.RemoteAddr(), err)
			}
			return
		}

		length := int(binary.BigEndian.Uint16(header))
		if length <= 0 {
			continue
		}

		// Read message body
		data := make([]byte, length)
		if _, err := io.ReadFull(reader, data); err != nil {
			log.Printf("Error reading body from %s: %v", conn.RemoteAddr(), err)
			return
		}

		// Unpack ISO 8583 message
		msg := iso8583.NewMessage(Spec8353)
		if err := msg.Unpack(data); err != nil {
			log.Printf("Unpack error: %v", err)
			continue
		}

		mti, _ := msg.GetMTI()
		log.Printf("Received MTI: %s from %s", mti, conn.RemoteAddr())

		s.processMessage(conn, msg)
	}
}

func (s *Server) processMessage(conn net.Conn, req *iso8583.Message) {
	var finReq FinancialRequest
	if err := req.Unmarshal(&finReq); err != nil {
		log.Printf("Unmarshal error: %v", err)
		return
	}

	log.Printf("Processed: PAN=%s Amount=%d RRN=%s", pkg.MaskPAN(finReq.PAN), finReq.Amount, finReq.RRN)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := s.fineractClient.CreateSavingsTransaction(ctx, 1, float64(finReq.Amount), finReq.RRN)
	responseCode := "00"
	if err != nil {
		log.Printf("Fineract error: %v", err)
		responseCode = "05"
	}

	reqMTI, _ := req.GetMTI()

	resp := iso8583.NewMessage(Spec8353)
	resp.MTI(pkg.ResponseMTI(reqMTI))

	// Echo transactional fields with actual values (not spec metadata)
	if finReq.PAN != "" {
		resp.Field(2, finReq.PAN)
	}
	if finReq.ProcessingCode != "" {
		resp.Field(3, finReq.ProcessingCode)
	}
	resp.Field(4, fmt.Sprintf("%012d", finReq.Amount))
	if finReq.STAN != "" {
		resp.Field(11, finReq.STAN)
	}
	if finReq.RRN != "" {
		resp.Field(37, finReq.RRN)
	}
	resp.Field(39, responseCode)

	packed, err := resp.Pack()
	if err != nil {
		log.Printf("Pack response error: %v", err)
		return
	}

	header := make([]byte, 2)
	binary.BigEndian.PutUint16(header, uint16(len(packed)))
	if _, err = conn.Write(append(header, packed...)); err != nil {
		log.Printf("Write response error: %v", err)
		return
	}

	log.Printf("Sent response (MTI %s, Code %s)", pkg.ResponseMTI(reqMTI), responseCode)
}
func (s *Server) Close() error {
	return s.listener.Close()
}
