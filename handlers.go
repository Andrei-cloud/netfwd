package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"net"
)

func Accepter(ctx context.Context, l net.Listener) {
	for {
		// Listen for an incoming connection
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		log.Println("incomming connection established")

		go connectionHandler(ctx, conn)
	}
}

func connectionHandler(ctx context.Context, conn net.Conn) {
	ctx, cancel := context.WithCancel(ctx)

	defer func() {
		log.Printf("closing connection with %s\n", conn.RemoteAddr().String())
		cancel()
		conn.Close()
	}()

	remote, err := net.Dial("tcp", *ForwardAddr)
	if err != nil {
		log.Printf("unable to establish remote connection: %v", err)
		return
	}
	defer remote.Close()
	log.Printf("Connected to remote host: %s\n", remote.RemoteAddr().String())

	proxyRequest := make(chan []byte, 10)
	proxyResponse, proxyErr := ProxyWorker(ctx, proxyRequest, remote)

	ApiRequest := make(chan []byte, 10)
	ApiResponse, ApiErr := APIWorker(ctx, ApiRequest)

	responseOut := make(chan []byte, 10)
	go SourceSenderWorker(ctx, responseOut, conn)

	quit := false

	go func() {
		defer func() {
			cancel()
			quit = true
		}()

		for {
			select {
			case err = <-proxyErr:
				if err == io.EOF {
					log.Printf("proxy worker: remote connection closed")
					return
				}
				log.Printf("proxy worker return error: %s", err)
				return
			case err = <-ApiErr:
				log.Printf("API worker return error: %s", err)
				return
			case <-ctx.Done():
				return
			default:
			}
		}
	}()

	for !quit {
		buf, err := Read(conn)
		if err != nil {
			if err == io.EOF {
				log.Println("remote connection closed")
				return
			}
			log.Printf("error reading: %s\n", err)
			return
		}
		//log.Printf("message received: %s\n", string(buf))

		if idx := bytes.Index(buf, []byte("CSNQ")); idx < 0 {
			log.Println("message is not CSNQ, passed through")
			proxyRequest <- buf
			responseOut <- <-proxyResponse
		} else {
			log.Println("message is  CSNQ, passed to API")
			ApiRequest <- buf
			responseOut <- <-ApiResponse
		}
	}

}
