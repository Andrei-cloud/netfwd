package main

import (
	"bytes"
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"log/slog"
	"net"
	"runtime"
	"sync"
	"time"
)

// Accepter handles incoming TCP connections
func Accepter(ctx context.Context, l net.Listener) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("Accepter shutting down")
			if err := l.Close(); err != nil {
				slog.Error("Error closing listener", "error", err)
			}
			return
		default:
			conn, err := l.Accept()
			if err != nil {
				if isTemporaryError(err) {
					slog.Warn("Temporary error accepting connection", "error", err)
					continue
				}
				slog.Error("Fatal error accepting connection", "error", err)
				return
			}

			slog.Info("Incoming connection established", "remoteAddr", conn.RemoteAddr().String())
			go connectionHandler(ctx, conn)
		}
	}
}

// isTemporaryError checks if the error is temporary and we should continue
func isTemporaryError(err error) bool {
	// As of Go 1.18, netErr.Temporary() is deprecated
	// Instead, we'll check for specific error types that represent temporary conditions
	if e, ok := err.(interface{ Timeout() bool }); ok && e.Timeout() {
		return true
	}

	// Check other types of errors that are typically temporary
	switch err.Error() {
	case "connection reset by peer", "use of closed network connection", "i/o timeout":
		return true
	}

	return false
}

// connectionHandler manages the lifecycle of a client connection
func connectionHandler(ctx context.Context, conn net.Conn) {
	ctx, cancel := context.WithCancel(ctx)

	errCh := make(chan error, 1)
	responseOut := make(chan *[]byte, 1)
	proxyRequest := make(chan *[]byte, 1)
	apiRequest := make(chan *[]byte, 1)

	// Create a map to track message processing times
	processingTimes := make(map[string]time.Time)
	var timesMutex sync.Mutex

	defer func() {
		slog.Info("Closing connection", "remoteAddr", conn.RemoteAddr().String())
		cancel()
		closeChannels(proxyRequest, responseOut, apiRequest)
		close(errCh)
		if err := conn.Close(); err != nil {
			slog.Error("Error closing connection", "error", err)
		}
	}()

	remote, err := net.Dial("tcp", *ForwardAddr)
	if err != nil {
		slog.Error("Unable to establish remote connection", "error", err)
		return
	}
	defer func() {
		if err := remote.Close(); err != nil {
			slog.Error("Error closing remote connection", "error", err)
		}
	}()

	slog.Info("Connected to remote host", "remoteAddr", remote.RemoteAddr().String())

	go SourceSenderWorker(ctx, responseOut, conn, errCh)

	proxyResponse := ProxyWorker(ctx, proxyRequest, remote, errCh)

	// Create API workers based on CPU count for parallel processing
	numWorkers := runtime.NumCPU()
	results := make([]<-chan *[]byte, numWorkers)
	for i := 0; i < numWorkers; i++ {
		results[i] = APIWorker(ctx, apiRequest, errCh)
	}

	apiResponses := FanIn(ctx, results...)
	quit := false

	// Define API-handled process codes - only CSNQ should be processed by API.
	apiProcessCodes := []string{"CSNQ"}

	// Error handling goroutine
	go func() {
		defer func() {
			cancel()
			quit = true
		}()

		for {
			select {
			case err := <-errCh:
				if err == io.EOF {
					slog.Info("Connection closed")
					return
				}
				slog.Error("Worker returned error", "error", err)
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	// Main message processing loop
	for !quit {
		buf, err := Read(conn)
		if err != nil {
			if err == io.EOF {
				slog.Info("Client connection closed")
				cancel()
				return
			}
			slog.Error("Error reading from client", "error", err)
			cancel()
			return
		}

		// Route message based on content - check for any of the API process codes
		routeToAPI := false
		for _, code := range apiProcessCodes {
			if idx := bytes.Index(buf, []byte(code)); idx >= 0 {
				routeToAPI = true
				slog.Info("Message has process code", "code", code, "routing", "API")
				break
			}
		}

		// Track the start time for latency measurement
		msgID := extractMessageID(buf)
		timesMutex.Lock()
		processingTimes[msgID] = time.Now()
		timesMutex.Unlock()

		var response *[]byte
		if routeToAPI {
			apiRequest <- &buf
			response = <-apiResponses
		} else {
			slog.Info("Message has no recognized API process code, passing through to proxy")
			proxyRequest <- &buf
			response = <-proxyResponse
		}

		// Calculate and log latency
		respMsgID := extractMessageID(*response)
		timesMutex.Lock()
		if startTime, ok := processingTimes[msgID]; ok {
			latency := time.Since(startTime)

			// Format latency based on its magnitude for better readability
			var latencyStr string
			switch {
			case latency < time.Microsecond:
				latencyStr = fmt.Sprintf("%.2f ns", float64(latency.Nanoseconds()))
			case latency < time.Millisecond:
				latencyStr = fmt.Sprintf("%.2f µs", float64(latency.Nanoseconds())/1000)
			case latency < time.Second:
				latencyStr = fmt.Sprintf("%.2f ms", float64(latency.Nanoseconds())/1000000)
			default:
				latencyStr = fmt.Sprintf("%.2f s", latency.Seconds())
			}

			slog.Info("Message processed",
				"msgID", msgID,
				"respMsgID", respMsgID,
				"latency", latencyStr,
				"latencyRaw", latency.String())
			delete(processingTimes, msgID)
		}
		timesMutex.Unlock()

		// Send response back to client
		responseOut <- response
	}
}

// closeChannels safely closes multiple channels
func closeChannels(channels ...chan *[]byte) {
	for _, ch := range channels {
		close(ch)
	}
}

// FanIn combines multiple channels into a single channel
func FanIn(ctx context.Context, channels ...<-chan *[]byte) <-chan *[]byte {
	var wg sync.WaitGroup
	multiplexStream := make(chan *[]byte)

	multiplex := func(ctx context.Context, c <-chan *[]byte) {
		defer wg.Done()
		for msg := range c {
			select {
			case <-ctx.Done():
				return
			case multiplexStream <- msg:
			}
		}
	}

	wg.Add(len(channels))
	for _, channel := range channels {
		go multiplex(ctx, channel)
	}

	go func() {
		wg.Wait()
		close(multiplexStream)
	}()

	return multiplexStream
}

// extractMessageID extracts a unique identifier from the message
// It looks for the STAN tag in XML messages which serves as a transaction ID
func extractMessageID(msg []byte) string {
	// Skip the length prefix (first 5 bytes) if present
	if len(msg) > 5 {
		contentStart := 0
		// Check if the first 5 bytes are numeric (length prefix)
		isPrefix := true
		for i := 0; i < 5 && i < len(msg); i++ {
			if msg[i] < '0' || msg[i] > '9' {
				isPrefix = false
				break
			}
		}
		if isPrefix {
			contentStart = 5
		}

		// Look for STAN tag in the message
		const stanTag = "<STAN>"
		const stanEndTag = "</STAN>"

		if idx := bytes.Index(msg[contentStart:], []byte(stanTag)); idx >= 0 {
			start := contentStart + idx + len(stanTag)
			if end := bytes.Index(msg[start:], []byte(stanEndTag)); end > 0 {
				return string(msg[start : start+end])
			}
		}
	}

	// If we can't extract a STAN, use a hash of the message as fallback
	h := fnv.New32a()
	h.Write(msg)
	return fmt.Sprintf("%x", h.Sum32())
}
