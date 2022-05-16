package main

import (
	"context"
	"io"
	"log"
	"net"
)

func Forward(dest net.Conn, b *[]byte) ([]byte, error) {
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
	return res, nil
}

func ProxyWorker(ctx context.Context, inMsg <-chan []byte, remote net.Conn) (chan []byte, chan error) {
	outMsg := make(chan []byte, 10)
	outErr := make(chan error, 1)

	go func() {
		defer close(outMsg)
		for {
			select {
			case message, ok := <-inMsg:
				if !ok {
					log.Println("inMsg is closed")
					return
				}
				if res, err := Forward(remote, &message); err != nil {
					outErr <- err
				} else {
					log.Printf("received response: %s\n", string(res))
					outMsg <- res
				}
			case <-ctx.Done():
				log.Println("ProxyWorker: CANCEL RECEIVED")
				return
			default:
			}
		}
	}()

	return outMsg, outErr
}

func APIWorker(ctx context.Context, inMsg <-chan []byte) (chan []byte, chan error) {
	outMsg := make(chan []byte, 10)
	outErr := make(chan error, 1)

	go func() {
		defer close(outMsg)
		for {
			select {
			case message, ok := <-inMsg:
				if !ok {
					log.Println("inMsg is closed")
					return
				}
				if res, err := CSNQ(&message); err != nil {
					outErr <- err
				} else {
					log.Printf("received response: %s\n", string(res))
					outMsg <- res
				}
			case <-ctx.Done():
				log.Println("API Worker: CANCEL RECEIVED")
				return
			default:
			}
		}
	}()

	return outMsg, outErr
}

func SourceSenderWorker(ctx context.Context, inMsg <-chan []byte, w io.Writer) {
	for {
		select {
		case message, ok := <-inMsg:
			if !ok {
				log.Println("SourceSenderWorker: inMsg is closed")
				return
			}
			if _, err := w.Write(message); err != nil {
				log.Println("SourceSenderWorker: " + err.Error())
				return
			}
		case <-ctx.Done():
			log.Println("SourceSenderWorker: CANCEL RECEIVED")
			return
		default:
		}
	}
}
