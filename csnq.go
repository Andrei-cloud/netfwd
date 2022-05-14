package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"
)

func CSNQ(req *[]byte) *[]byte {
	var msg []byte
	if request, err := RequestX2J(*req); err == nil {
		client := clientPool.Get().(*resty.Client)

		resp, err := client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).R().
			SetBody(request).
			Post(destURL.String())

		if err != nil {
			log.Printf("http request failed: %v\n", err)
		}

		clientPool.Put(client)

		if status := resp.StatusCode(); status == http.StatusOK {
			if response, err := ResponseJ2X(resp.Body()); err == nil {
				l := []byte(fmt.Sprintf("%05d", len(response)))
				msg = append(l, response...)
			}
		}
	}
	return &msg
}
