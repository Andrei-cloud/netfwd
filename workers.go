package main

import (
	"context"
	"crypto/tls"
	"io"
	"log/slog"
	"net"
	"net/http"
)

// Forward sends a message to a destination connection and reads the response.
func Forward(dest net.Conn, b *[]byte) (*[]byte, error) {
	if _, err := dest.Write(*b); err != nil {
		return nil, err
	}

	res, err := Read(dest)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// ProxyWorker forwards messages to a remote TCP endpoint and returns responses.
func ProxyWorker(
	ctx context.Context,
	inMsg <-chan *[]byte,
	remote net.Conn,
	errCh chan error,
) chan *[]byte {
	outMsg := make(chan *[]byte, 1)

	go func() {
		defer close(outMsg)
		for {
			select {
			case message, ok := <-inMsg:
				if !ok {
					slog.Info("ProxyWorker: input channel closed")
					return
				}

				res, err := Forward(remote, message)
				if err != nil {
					slog.Error("ProxyWorker: forwarding error", "error", err)
					errCh <- err
					continue
				}

				outMsg <- res

			case <-ctx.Done():
				slog.Info("ProxyWorker: context canceled")
				return
			}
		}
	}()

	return outMsg
}

// APIWorker processes messages through the HTTP API.
func APIWorker(ctx context.Context, inMsg <-chan *[]byte, outErr chan<- error) chan *[]byte {
	outMsg := make(chan *[]byte, 1)

	// Create HTTP client with common configuration
	client := createHTTPClient()

	go func() {
		defer close(outMsg)
		for {
			select {
			case message, ok := <-inMsg:
				if !ok {
					slog.Info("APIWorker: input channel closed")
					return
				}

				res, err := CSNQ(client, message)
				if err != nil {
					slog.Error("APIWorker: CSNQ processing error", "error", err)
					outErr <- err
					continue
				}

				outMsg <- res

			case <-ctx.Done():
				slog.Info("APIWorker: context canceled")
				return
			}
		}
	}()

	return outMsg
}

// createHTTPClient creates a preconfigured HTTP client.
func createHTTPClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &http.Client{
		Transport: transport,
	}
}

// SourceSenderWorker sends responses back to the original client.
func SourceSenderWorker(
	ctx context.Context,
	inMsg <-chan *[]byte,
	w io.Writer,
	errCh chan<- error,
) {
	for {
		select {
		case message, ok := <-inMsg:
			if !ok {
				slog.Info("SourceSenderWorker: input channel closed")
				return
			}

			if _, err := w.Write(*message); err != nil {
				slog.Error("SourceSenderWorker: write error", "error", err)
				errCh <- err
			}

		case <-ctx.Done():
			slog.Info("SourceSenderWorker: context canceled")
			return
		}
	}
}
