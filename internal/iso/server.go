package iso

import (
	"bufio"
	"context"
	"encoding/binary"
	"log"
	"net"
	"strings"
	"time"

	"github.com/karma-234/gateway-ps/internal/fineract"
	"github.com/moov-io/iso8583"
	"github.com/moov-io/iso8583/examples"
)

type Server struct {
	listener       net.Listener
	fineractClient *fineract.Client
}

func NewServer(addr string) (*Server, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	s := &Server{
		listener:       ln,
		fineractClient: fineract.NewClient("http://localhost:8081/api/v1", "mifos", "password"),
	}
	go s.acceptLoop()
	return s, nil
}

func (s *Server) acceptLoop() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("Accepted connection from %s", conn.RemoteAddr())
	reader := bufio.NewReader(conn)
	for {
		header := make([]byte, 2)
		_, err := reader.Read(header)
		if err != nil {
			log.Printf("Error reading header: %v", err)
			return
		}
		length := int(binary.BigEndian.Uint16(header))
		// Read the rest of the message based on the length
		data := make([]byte, length)
		_, err = reader.Read(data)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			return
		}
		// Process the message
		msg := iso8583.NewMessage(examples.Spec)
		err = msg.Unpack(data)
		if err != nil {
			log.Printf("Error unpacking message: %v", err)
			return
		}
		log.Printf("Received message: %v", msg)
		// Here you would handle the message and potentially send a response

		mti, _ := msg.GetMTI()
		log.Printf("Message MTI: %s", mti)

		s.processMessage(conn, msg)
	}
}

func (s *Server) processMessage(conn net.Conn, req *iso8583.Message) {
	// Here you would implement your business logic to process the message
	var finRq FinancialRequest
	err := req.Unmarshal(&finRq)
	if err != nil {
		log.Printf("Error unmarshaling message to struct: %v", err)
		return
	}
	log.Printf("Unmarshaled FinancialRequest: %+v", finRq)
	log.Printf("Masked PAN: %s Amount: %d, STAN: %s", maskPAN(finRq.PAN), finRq.Amount, finRq.STAN)

	// For demonstration, we will just send a simple response back
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = s.fineractClient.CreateSavingsTransaction(ctx, 12345, float64(finRq.Amount), finRq.RRN)
	responseCode := "00"
	if err != nil {
		log.Printf("Error creating savings transaction in Fineract: %v", err)
		responseCode = "05" // GENERIC ERROR RESPONSE CODE
	}

	resp := iso8583.NewMessage(Spec8353)
	resp.MTI("0210")
	resp.Field(39, responseCode)
	resp.Field(37, finRq.RRN) // Response code for success
	respDataPacked, err := resp.Pack()
	if err != nil {
		log.Printf("Error packing response: %v", err)
		return
	}
	header := make([]byte, 2)
	binary.BigEndian.PutUint16(header, uint16(len(respDataPacked)))
	_, err = conn.Write(append(header, respDataPacked...))
	if err != nil {
		log.Printf("Error writing response: %v", err)
		return
	}
	log.Printf("Sent response: %v", resp)

}

func (s *Server) Close() error {
	return s.listener.Close()
}

func maskPAN(pan string) string {
	if len(pan) <= 10 {
		return pan
	}
	var res strings.Builder
	res.WriteString(pan[:6])
	res.WriteString("******")
	res.WriteString(pan[len(pan)-4:])
	return res.String()
}
