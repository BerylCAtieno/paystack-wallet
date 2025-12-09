package paystack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	secretKey string
	baseURL   string
}

func NewClient(secretKey string) *Client {
	return &Client{
		secretKey: secretKey,
		baseURL:   "https://api.paystack.co",
	}
}

type InitializeRequest struct {
	Email     string `json:"email"`
	Amount    int64  `json:"amount"` // in kobo
	Reference string `json:"reference"`
}

type InitializeResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

type VerifyResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID              int64  `json:"id"`
		Reference       string `json:"reference"`
		Amount          int64  `json:"amount"`
		Status          string `json:"status"`
		PaidAt          string `json:"paid_at"`
		Channel         string `json:"channel"`
		Currency        string `json:"currency"`
		IPAddress       string `json:"ip_address"`
		CreatedAt       string `json:"createdAt"`
		GatewayResponse string `json:"gateway_response"`
	} `json:"data"`
}

func (c *Client) InitializeTransaction(email string, amount int64, reference string) (*InitializeResponse, error) {
	reqBody := InitializeRequest{
		Email:     email,
		Amount:    amount,
		Reference: reference,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/transaction/initialize", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.secretKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var initResp InitializeResponse
	if err := json.Unmarshal(body, &initResp); err != nil {
		return nil, err
	}

	if !initResp.Status {
		return nil, fmt.Errorf("paystack error: %s", initResp.Message)
	}

	return &initResp, nil
}

func (c *Client) VerifyTransaction(reference string) (*VerifyResponse, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/transaction/verify/"+reference, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.secretKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var verifyResp VerifyResponse
	if err := json.Unmarshal(body, &verifyResp); err != nil {
		return nil, err
	}

	return &verifyResp, nil
}
