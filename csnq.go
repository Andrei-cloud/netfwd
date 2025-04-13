package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CSNQ handles the transformation of messages to HTTP API calls and back.
func CSNQ(client *http.Client, req *[]byte) (*[]byte, error) {
	// Transform XML request to JSON
	request, err := RequestX2J(*req)
	if err != nil {
		return nil, fmt.Errorf("failed to transform XML to JSON: %w", err)
	}

	// Create HTTP request with the JSON body
	httpReq, err := http.NewRequest(http.MethodPost, DestURL.String(), bytes.NewReader(request))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "go-frwd/0.0.1")

	// Add Basic Authentication
	auth := base64.StdEncoding.EncodeToString([]byte(*Username + ":" + *Password))
	httpReq.Header.Set("Authorization", "Basic "+auth)

	// Make the API call
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Process response based on status code
	if resp.StatusCode == http.StatusOK {
		response, err := ResponseJ2X(body)
		if err != nil {
			return nil, fmt.Errorf("failed to transform JSON response to XML: %w", err)
		}

		// Format response with length prefix
		l := []byte(fmt.Sprintf("%0*d", lengthSize, len(response)))
		msg := append(l, response...)
		return &msg, nil
	}

	// Handle error responses
	return nil, processErrorResponse(body)
}

// processErrorResponse extracts error information from the API response
func processErrorResponse(body []byte) error {
	errResponse := struct {
		Message string `json:"message"`
	}{}

	if err := json.Unmarshal(body, &errResponse); err != nil {
		// If we can't parse the error message, return the raw error
		return fmt.Errorf("API error (unparseable response)")
	}

	return fmt.Errorf("API error: %s", errResponse.Message)
}
