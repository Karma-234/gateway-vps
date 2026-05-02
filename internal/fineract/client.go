package fineract

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
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
	TransactionDate   string `json:"transactionDate"`
	TransactionAmount string `json:"transactionAmount"`
	Locale            string `json:"locale"`
	DateFormat        string `json:"dateFormat"`
	PaymentTypeID     int8   `json:"paymentTypeId,omitempty"`
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
	password := pkg.GetEnv("fineract_password", "password")
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
		TransactionDate:   time.Now().Format("02 January 2006"),
		TransactionAmount: fmt.Sprintf("%.2f", amount/100.0),
		Locale:            "en",
		DateFormat:        "dd MMMM yyyy",
		PaymentTypeID:     1,
	}
	body, _ := json.Marshal(req)
	url := fmt.Sprintf("%s/savingsaccounts/%d/transactions?command=deposit", c.baseURL, accountID)
	httpReq, err := retryablehttp.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))

	if err != nil {
		return err
	}
	httpReq.SetBasicAuth(c.username, c.password)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Fineract-Platform-TenantId", "default")

	_, err = c.cb.Execute(func() (any, error) {
		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil, nil
		}
		log.Printf("Fineract deposit failed for account %s | Amount: %.2f | RRN: %s | Status: %d",
			pkg.MaskPAN(strconv.FormatInt(accountID, 10)), amount/100, rrn, resp.StatusCode)
		b, _ := io.ReadAll(resp.Body)
		log.Printf("Fineract error response body: %s", string(b))
		return nil, fmt.Errorf("unexpected status code: %d body: %s", resp.StatusCode, string(b))
	})

	if err != nil {
		return fmt.Errorf("failed to create savings transaction: %w", err)
	}
	log.Printf("Fineract deposit recorded for account %s | Amount: %.2f | RRN: %s",
		pkg.MaskPAN(strconv.FormatInt(accountID, 10)), amount/100, rrn)
	return nil
}
