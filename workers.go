package main

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"net"

	"github.com/go-resty/resty/v2"
)

func Forward(dest net.Conn, b *[]byte) (*[]byte, error) {
	var (
		res []byte
		err error
	)

	if _, err := dest.Write(*b); err != nil {
		return nil, err
	}
	if res, err = Read(dest); err != nil {
		return nil, err
	}
	return &res, nil
}

func ProxyWorker(ctx context.Context, inMsg <-chan *[]byte, remote net.Conn, errCh chan error) chan *[]byte {
	outMsg := make(chan *[]byte, 1)

	go func() {
		defer close(outMsg)
		for {
			select {
			case message, ok := <-inMsg:
				if !ok {
					log.Println("inMsg is closed")
					return
				}
				if res, err := Forward(remote, message); err != nil {
					errCh <- err
				} else {
					//log.Printf("received response: %s\n", string(res))
					proxyProcessed.Inc()
					outMsg <- res
				}
			case <-ctx.Done():
				log.Println("ProxyWorker: CANCEL RECEIVED")
				return
			}
		}
	}()

	return outMsg
}

func APIWorker(ctx context.Context, inMsg <-chan *[]byte, outErr chan<- error) chan *[]byte {
	outMsg := make(chan *[]byte, 1)

	client := resty.New().
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
		SetHeader("User-Agent", "go-frwd/0.0.1").
		SetHeader("Content-Type", "application/json").
		SetBasicAuth(*Username, *Password)

	go func() {
		defer close(outMsg)
		for {
			select {
			case message, ok := <-inMsg:
				if !ok {
					log.Println("API Worker: in channel is closed")
					return
				}
				if res, err := CSNQ(client, message); err != nil {
					outErr <- err
				} else {
					//log.Printf("received response: %s\n", string(res))
					csnqProcessed.Inc()
					outMsg <- res
				}
			case <-ctx.Done():
				log.Println("API Worker: CANCEL RECEIVED")
				return
			}
		}
	}()

	return outMsg
}

func SourceSenderWorker(ctx context.Context, inMsg <-chan *[]byte, w io.Writer, errCh chan<- error) {
	for {
		select {
		case message, ok := <-inMsg:
			if !ok {
				log.Println("SourceSenderWorker: in channel is closed")
				return
			}
			if _, err := w.Write(*message); err != nil {
				errCh <- err
			}
		case <-ctx.Done():
			log.Println("SourceSenderWorker: CANCEL RECEIVED")
			return
		}
	}
}
