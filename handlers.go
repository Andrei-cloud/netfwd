package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"net"
	"runtime"
	"sync"
)

func Accepter(ctx context.Context, l net.Listener) {
	for {
		// Listen for an incoming connection
		select {
		case <-ctx.Done():
			l.Close()
			return
		default:
			conn, err := l.Accept()
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("incomming connection established: %s", conn.RemoteAddr().String())

			go connectionHandler(ctx, conn)
		}
	}
}

func connectionHandler(ctx context.Context, conn net.Conn) {
	ctx, cancel := context.WithCancel(ctx)

	ErrCh := make(chan error)
	responseOut := make(chan *[]byte, 1)
	proxyRequest := make(chan *[]byte, 1)
	ApiRequest := make(chan *[]byte, 1)

	defer func() {
		log.Printf("closing connection with %s\n", conn.RemoteAddr().String())
		close(proxyRequest)
		close(responseOut)
		close(ApiRequest)
		close(ErrCh)
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

	go SourceSenderWorker(ctx, responseOut, conn, ErrCh)

	proxyResponse := ProxyWorker(ctx, proxyRequest, remote, ErrCh)

	numWorkers := runtime.NumCPU()
	results := make([]<-chan *[]byte, numWorkers)
	for i := 0; i < numWorkers; i++ {
		results[i] = APIWorker(ctx, ApiRequest, ErrCh)
	}

	ApiResponses := FanIn(ctx, results...)
	quit := false

	go func() {
		defer func() {
			cancel()
			quit = true
		}()

		for {
			select {
			case err = <-ErrCh:
				if err == io.EOF {
					log.Printf("proxy worker: remote connection closed")
					return
				}
				log.Printf("worker return error: %s", err)
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	for !quit {
		buf, err := Read(conn)
		if err != nil {
			if err == io.EOF {
				log.Println("remote connection closed")
				cancel()
				return
			}
			log.Printf("error reading: %s\n", err)
			cancel()
			return
		}
		//log.Printf("message received: %s\n", string(buf))

		if idx := bytes.Index(buf, []byte("CSNQ")); idx < 0 {
			log.Println("message is not CSNQ, passed through")
			proxyRequest <- &buf
			responseOut <- <-proxyResponse
		} else {
			log.Println("message is  CSNQ, passed to API")
			ApiRequest <- &buf
			responseOut <- <-ApiResponses
		}
	}

}

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
