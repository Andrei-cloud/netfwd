package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

func CSNQ(client *resty.Client, req *[]byte) (*[]byte, error) {
	var msg []byte

	request, err := RequestX2J(*req)
	if err != nil {
		return nil, fmt.Errorf("CSNQ: %w", err)
	}

	resp, err := client.R().
		SetBody(request).
		Post(DestURL.String())

	if err != nil {
		return nil, fmt.Errorf("CSNQ http request failed: %w", err)
	}

	if status := resp.StatusCode(); status == http.StatusOK {
		if response, err := ResponseJ2X(resp.Body()); err == nil {
			l := []byte(fmt.Sprintf("%05d", len(response)))
			msg = append(l, response...)
			return &msg, nil
		}
		if err != nil {
			return nil, fmt.Errorf("CSNQ: %w", err)
		}
	}

	errResponse := struct {
		Message string `json:"message"`
	}{}

	err = json.Unmarshal(resp.Body(), &errResponse)
	if err == nil {
		err = fmt.Errorf("CSNQ: %s", errResponse.Message)
	}

	return nil, fmt.Errorf("CSNQ: %w", err)
}
