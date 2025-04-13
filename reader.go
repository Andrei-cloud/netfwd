package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
)

// Read reads a message from a network connection with a length prefix
// The format is: [5 bytes length prefix][message body]
func Read(conn net.Conn) ([]byte, error) {
	var buf bytes.Buffer

	// Read length prefix (first 5 bytes)
	if _, err := io.CopyN(&buf, conn, int64(lengthSize)); err != nil {
		return nil, fmt.Errorf("failed to read message length: %w", err)
	}

	// Parse the length prefix to determine message size
	length, err := strconv.Atoi(buf.String())
	if err != nil {
		return nil, fmt.Errorf("invalid message length format: %w", err)
	}

	// Validate message length to prevent potential memory issues
	if length <= 0 || length > 10*1024*1024 { // 10MB max message size
		return nil, fmt.Errorf("invalid message length: %d", length)
	}

	// Read the actual message body
	if _, err := io.CopyN(&buf, conn, int64(length)); err != nil {
		return nil, fmt.Errorf("failed to read message body: %w", err)
	}

	return buf.Bytes(), nil
}
