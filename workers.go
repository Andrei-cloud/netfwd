package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
)

func Forward(dest io.ReadWriter, b *[]byte) ([]byte, error) {
	var (
		res []byte
		err error
	)

	if _, err := dest.Write(*b); err != nil {
		return nil, err
	}
	if res, err = Read(bufio.NewReader(dest)); err != nil {
		return nil, err
	}
	return res, nil
}

func PassThroughWorker(ctx context.Context, inMsg <-chan []byte, rw io.ReadWriter) chan []byte {
	outMsg := make(chan []byte, 5)

	go func() {
		defer close(outMsg)
		for {
			select {
			case message, ok := <-inMsg:
				if !ok {
					fmt.Println("inMsg is closed")
					return
				}
				if res, err := Forward(rw, &message); err != nil {
					fmt.Println(err)
					return
				} else {
					log.Printf("received response: %s\n", string(res))
					outMsg <- res
				}

			case <-ctx.Done():
				fmt.Println("PassThroughWorker: CANCEL RECEIVED")
				return
			}
		}
	}()

	return outMsg
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
		}
	}
}
