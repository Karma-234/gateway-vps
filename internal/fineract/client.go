package fineract

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/karma-234/gateway-ps/internal/pkg"
	"github.com/sony/gobreaker"
)

type Client struct {
	baseURL    string
	username   string
	password   string
	cb         *gobreaker.CircuitBreaker
	httpClient *retryablehttp.Client
}

type SavingsTransactionRequest struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
	TypeID int     `json:"typeID"` // "1 = Cash, 2 = Transfer, 3 = Withdrawal, 4 = Interest Posting, 5 = Fee Deduction, 6 = Dividend Posting, 7 = Refund, 8 = Chargeback, 9 = Credit Balance Adjustment, 10 = Debit Balance Adjustment"
}

func NewClient() *Client {
	cbSettings := gobreaker.Settings{
		Name:        "FineractClient",
		MaxRequests: 5,
		Interval:    2 * time.Second,
		Timeout:     1 * time.Minute}
	cb := gobreaker.NewCircuitBreaker(cbSettings)
	httpClient := retryablehttp.NewClient()
	httpClient.RetryMax = 3
	httpClient.RetryWaitMin = 500 * time.Millisecond
	httpClient.RetryWaitMax = 2 * time.Second
	baseURL := pkg.GetEnv("FINERACT_BASE_URL", "http://localhost:8081/api/v1")
	username := pkg.GetEnv("FINERACT_USERNAME", "mifos")
	password := pkg.GetEnv("FINERACT_PASSWORD", "password")
	return &Client{
		baseURL:    baseURL,
		username:   username,
		password:   password,
		cb:         cb,
		httpClient: httpClient,
	}
}

func (c *Client) CreateSavingsTransaction(ctx context.Context, accountID int64, amount float64, rrn string) error {
	req := &SavingsTransactionRequest{
		Date:   time.Now().Format("2006-01-02"),
		Amount: amount,
		TypeID: 1, // Cash
	}
	body, _ := json.Marshal(req)
	url := fmt.Sprintf("%s/savingsaccounts/%d/transactions?command=deposit", c.baseURL, accountID)
	httpReq, err := retryablehttp.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.SetBasicAuth(c.username, c.password)
	httpReq.Header.Set("Content-Type", "application/json")

	_, err = c.cb.Execute(func() (any, error) {
		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil, nil
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	})

	if err != nil {
		return fmt.Errorf("failed to create savings transaction: %w", err)
	}
	log.Printf("✅ Fineract deposit recorded for account %d | Amount: %.2f | RRN: %s",
		accountID, amount/100, rrn)
	return nil
}
