package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func main() {
	responder := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log incoming request
		requestBody, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
		}

		// Log request details
		log.Printf("--- INCOMING REQUEST ---")
		log.Printf("Time: %s", time.Now().Format(time.RFC3339))
		log.Printf("Method: %s", r.Method)
		log.Printf("URL: %s", r.URL.String())
		log.Printf("Headers: %v", r.Header)
		log.Printf("Body: %s", string(requestBody))
		log.Printf("------------------------")

		// Prepare response
		responseBody := []byte(
			`{"RequestInfo":{"requestId":"0220000245250","userId":"256557","basenumber":"157336",` +
				`"chanelId":"ATM","requestTime":"2203221157"},"CustomerDetails":[{"QID":"2734XXXXXXX",` +
				`"BASENO":"157336","CRNO":null,"PASSPORTNO":"XXXXXX","MOBILENO":"","EMAILID":"XXXXXX@example.com",` +
				`"GUID":"7c7f7a47-f236-ea11-9132-00505685b1c3","FirstName":"XXXXXX","LastName":"XXXXXX",` +
				`"IsBlacklisted":false,"IsQANationalityWithdrawn":false}]}`,
		)

		// Set headers
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Write response
		bytesWritten, err := w.Write(responseBody)
		if err != nil {
			log.Printf("Error writing response: %v", err)
		}

		// Log outgoing response
		log.Printf("--- OUTGOING RESPONSE ---")
		log.Printf("Time: %s", time.Now().Format(time.RFC3339))
		log.Printf("Status: %d", http.StatusOK)
		log.Printf("Headers: %v", w.Header())
		log.Printf("Body: %s", string(responseBody))
		log.Printf("Bytes written: %d", bytesWritten)
		log.Printf("-------------------------")
	})

	fmt.Println("Mock web server starting on port 3030...")
	log.Fatal(http.ListenAndServe(":3030", responder))
}
