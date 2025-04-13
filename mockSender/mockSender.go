package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"
)

// List of possible processing codes:
var processingCodes = []string{
	"CSNQ", // Original code
	"ACNQ", // Account Number Query
	"BRNQ", // Branch Query
	"CUST", // Customer Information
	"TRNQ", // Transaction Query
}

func main() {
	log.Println("starting sender mock")
	start := time.Now()

	// Initialize a new random number generator with a time-based seed
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		log.Println(err)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}()

	tick := time.NewTicker(1 * time.Millisecond)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(15*time.Second))
	quit := false

	requestStart := time.Now()
	counter := 0
	for !quit {
		select {
		case <-ctx.Done():
			log.Println(ctx.Err())
			tick.Stop()
			quit = true
		case <-tick.C:
			// Get a random processing code using the local random generator
			procCode := processingCodes[r.Intn(len(processingCodes))]
			log.Printf("Sending with ProcCode: %s", procCode)

			// Generate the request with the selected processing code
			request := generateRequest(procCode)

			err := Sender(ctx, conn, request)
			fmt.Println("received in: ", time.Since(requestStart))
			counter++
			if err != nil {
				log.Println(err)
				tick.Stop()
				quit = true
			}
		}
	}
	cancel()
	log.Println("quiting sender mock")
	log.Printf("Run time: %s processed: %d\n", time.Since(start), counter)
}

// generateRequest creates a request with the specified processing code.
func generateRequest(procCode string) []byte {
	xmlTemplate := `<XML><MessageType>0</MessageType><ProcCode>%s</ProcCode>` +
		`<REFNUM>0220000245250</REFNUM><STAN>0220000245250</STAN>` +
		`<LocalTxnDtTime>2203221157</LocalTxnDtTime><DeliveryChannelCtrlID>ATM</DeliveryChannelCtrlID>` +
		`<PName>ACCOUNTNUMBER</PName><PValue>157336</PValue></XML>`

	xml := fmt.Sprintf(xmlTemplate, procCode)

	// Calculate and prefix the message length (5 digits)
	lengthPrefix := fmt.Sprintf("%05d", len(xml))

	return []byte(lengthPrefix + xml)
}

// Sender sends a request and reads the response, now with logging of message contents.
func Sender(_ context.Context, conn net.Conn, request []byte) error {
	var l int

	// Log the request being sent
	log.Printf("SENDING REQUEST: %s", string(request))

	_, err := conn.Write(request)
	if err != nil {
		return err
	}

	netLen := make([]byte, 5)

	for {
		n, err := conn.Read(netLen)
		if err != nil {
			return err
		}
		if n == 5 {
			l, err = strconv.Atoi(string(netLen))
			if err != nil {
				return errors.New("invalid msg length received: " + err.Error())
			}

			break
		}
	}

	buf := make([]byte, l)
	_, err = conn.Read(buf)
	if err != nil {
		return err
	}

	// Log the response received
	log.Printf("RECEIVED RESPONSE: %s", string(buf))

	return nil
}
