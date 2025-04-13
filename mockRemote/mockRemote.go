package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

func main() {
	l, err := net.Listen("tcp", ":9002")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = l.Close() // Ignoring close error
	}()
	log.Println("Listening on " + l.Addr().String())

	for {
		// Listen for an incoming connection
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Fatal error: %v", err)
			return
		}

		log.Printf(
			"[%s] New connection established from %s",
			time.Now().Format(time.RFC3339),
			conn.RemoteAddr().String(),
		)

		go func(c net.Conn) {
			defer func() {
				log.Printf(
					"[%s] Closing connection from %s",
					time.Now().Format(time.RFC3339),
					c.RemoteAddr().String(),
				)
				_ = c.Close() // Ignoring close error
			}()
			buf := make([]byte, 4096)

			for {
				n, err := c.Read(buf)
				if err != nil {
					if err == io.EOF {
						log.Printf(
							"[%s] Remote connection from %s is closed",
							time.Now().Format(time.RFC3339),
							c.RemoteAddr().String(),
						)

						return
					}
					log.Printf(
						"[%s] Error reading from %s: %s",
						time.Now().Format(time.RFC3339),
						c.RemoteAddr().String(),
						err,
					)

					return
				}
				if n > 0 {
					// Log incoming data
					logData("RECV", c.RemoteAddr().String(), buf[:n])

					// Echo back
					l, err := c.Write(buf[:n])
					if err != nil {
						log.Printf(
							"[%s] Error writing to %s: %s",
							time.Now().Format(time.RFC3339),
							c.RemoteAddr().String(),
							err,
						)

						return
					}

					// Log outgoing data
					logData("SENT", c.RemoteAddr().String(), buf[:l])
				}
			}
		}(conn)
	}
}

// logData logs detailed information about the data being sent or received.
func logData(direction, addr string, data []byte) {
	log.Printf(
		"[%s] %s %d bytes to/from %s",
		time.Now().Format(time.RFC3339),
		direction,
		len(data),
		addr,
	)

	// Parse the message format: [5 bytes length prefix][message body]
	if len(data) > 5 {
		// Check if the first 5 bytes could be a length prefix
		lenPrefix := string(data[:5])
		msgLength := 0
		lenPrefixValid := true

		// Validate the length prefix
		for i := 0; i < 5; i++ {
			if data[i] < '0' || data[i] > '9' {
				lenPrefixValid = false
				break
			}
		}

		if lenPrefixValid {
			var err error
			msgLength, err = strconv.Atoi(lenPrefix)
			if err == nil && msgLength > 0 && len(data) >= 5+msgLength {
				// We have a valid length-prefixed message
				msgBody := data[5 : 5+msgLength]
				log.Printf("MESSAGE STRUCTURE:")
				log.Printf("  Length Prefix: %s (indicates %d bytes)", lenPrefix, msgLength)
				log.Printf("  Body (%d bytes): %s", len(msgBody), string(msgBody))

				// Try to pretty-print XML if it looks like XML
				if bytes.HasPrefix(msgBody, []byte("<")) && bytes.HasSuffix(msgBody, []byte(">")) {
					log.Printf("  Message appears to be XML")
				}

				fmt.Println(strings.Repeat("-", 40)) // Separator

				return
			}
		}
	}

	// If we couldn't parse as a structured message, fall back to regular logging

	// Print hex dump of data for detailed inspection
	dump := hex.Dump(data)
	log.Printf("Data Hex Dump:\n%s", dump)

	// Try to print as string if it might be text
	printable := true
	for _, b := range data {
		if b < 32 || b > 126 {
			if b != '\r' && b != '\n' && b != '\t' {
				printable = false
				break
			}
		}
	}

	if printable {
		log.Printf("Data as string: %s", string(data))
	}

	fmt.Println(strings.Repeat("-", 40)) // Separator for readability
}
