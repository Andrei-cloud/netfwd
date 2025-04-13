package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

// CSNQ handles the transformation of messages to HTTP API calls and back.
func CSNQ(client *resty.Client, req *[]byte) (*[]byte, error) {
	// Transform XML request to JSON
	request, err := RequestX2J(*req)
	if err != nil {
		return nil, fmt.Errorf("failed to transform XML to JSON: %w", err)
	}

	// Make API call
	resp, err := client.R().
		SetBody(request).
		Post(DestURL.String())
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	// Process response based on status code
	if status := resp.StatusCode(); status == http.StatusOK {
		response, err := ResponseJ2X(resp.Body())
		if err != nil {
			return nil, fmt.Errorf("failed to transform JSON response to XML: %w", err)
		}

		// Format response with length prefix
		l := []byte(fmt.Sprintf("%0*d", lengthSize, len(response)))
		msg := append(l, response...)
		return &msg, nil
	}

	// Handle error responses
	return nil, processErrorResponse(resp.Body())
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
