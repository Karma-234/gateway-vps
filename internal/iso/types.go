package iso

type Acceptor struct {
	Name    string `iso8583:"1"` // Card Acceptor Name/Location
	City    string `iso8583:"2"` // Card Acceptor Name/Location
	Country string `iso8583:"3"` // Card Acceptor Name/Location
}

type FinancialRequest struct {
	MTI            string    `iso8583:"0"`  // Message Type Indicator
	ProcessingCode string    `iso8583:"3"`  // Processing Code
	PAN            string    `iso8583:"2"`  // Primary Account Number
	Amount         string    `iso8583:"4"`  // Transaction Amount
	TransmissionDT string    `iso8583:"7"`  // Transmission Date & Time
	STAN           string    `iso8583:"11"` // System Trace Audit Number (STAN)
	ExpDate        string    `iso8583:"14"` // Expiration Date
	RRN            string    `iso8583:"37"` // Retrieval Reference Number
	TerminalID     string    `iso8583:"41"` // Terminal ID
	MerchantID     string    `iso8583:"42"` // Merchant ID
	Acceptor       *Acceptor `iso8583:"43"` // Card Acceptor Name/Location (composite field)
	CurrencyCode   string    `iso8583:"49"` // Currency Code
}
