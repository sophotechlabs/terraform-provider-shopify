package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	StoreURL    string
	AccessToken string
	APIVersion  string
	Endpoint    string
	HTTPClient  *http.Client
}

type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []graphQLError  `json:"errors,omitempty"`
}

type graphQLError struct {
	Message string `json:"message"`
}

type userError struct {
	Field   []string `json:"field"`
	Message string   `json:"message"`
}

func NewClient(storeURL, accessToken, apiVersion string) *Client {
	return &Client{
		StoreURL:    storeURL,
		AccessToken: accessToken,
		APIVersion:  apiVersion,
		Endpoint:    fmt.Sprintf("https://%s/admin/api/%s/graphql.json", storeURL, apiVersion),
		HTTPClient:  &http.Client{},
	}
}

func (c *Client) Execute(query string, variables map[string]any) (*graphQLResponse, error) {
	reqBody := graphQLRequest{Query: query, Variables: variables}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", c.Endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shopify-Access-Token", c.AccessToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("shopify API returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var gqlResp graphQLResponse
	if err := json.Unmarshal(respBytes, &gqlResp); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("graphql errors: %s", gqlResp.Errors[0].Message)
	}

	return &gqlResp, nil
}
