package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-resty/resty/v2"
)

//to minimize cleint coldstart
var (
	clientPool   sync.Pool
	destPtr      *string
	traceInfoPtr *bool
	destURL      *url.URL

	fwdIn chan []byte
)

const lengthSize int = 5

func main() {
	var err error
	portPtr := flag.String("p", "3000", "port to listen to")
	traceInfoPtr = flag.Bool("t", false, "print trace info")
	destPtr = flag.String("d", "http://localhost:3030/", "HTTP destination endpoint")
	username := flag.String("u", "ecms", "user name, mandatory")
	password := flag.String("s", "ecms1", "user password, mandatory")

	forwardAddr := flag.String("f", ":9002", "address to passthrough")

	flag.Parse()

	if *username == "" || *password == "" {
		log.Fatal("username and password must be provided")
	}

	destURL, err = url.Parse(*destPtr)
	if err != nil {
		log.Fatalln(err)
	}

	if *destPtr == "" {
		log.Fatalln("destination http endpoint must be required")
	}

	// Listen for incoming connections.
	addr := fmt.Sprintf(":%s", *portPtr)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer l.Close()
	host, port, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		panic(err)
	}
	log.Printf("Listening on host: %s, port: %s\n", host, port)

	clientPool = sync.Pool{
		New: func() any {
			return resty.New().EnableTrace().
				SetHeader("User-Agent", "go-frwd/0.0.1").
				SetHeader("Content-Type", "application/json").
				SetBasicAuth(*username, *password)
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	host, port, err = net.SplitHostPort(*forwardAddr)
	if err != nil {
		panic(err)
	}

	remote, err := net.Dial("tcp", *forwardAddr)
	if err != nil {
		log.Fatalf("unable to establish remode connection: %v", err)
	}
	log.Printf("Connected to remote host: %s, port: %s\n", host, port)

	fwdIn = make(chan []byte, 5)
	pt := PassThroughWorker(ctx, fwdIn, remote)

	go func(ctx context.Context) {
		for {
			// Listen for an incoming connection
			conn, err := l.Accept()
			if err != nil {
				log.Fatal(err)
			}

			log.Println("incomming connection established")

			go func(ctx context.Context, c net.Conn) {
				defer func() {
					log.Printf("closing connection with %s\n", c.RemoteAddr().String())
					c.Close()
				}()

				out := make(chan []byte, 5)
				defer close(out)
				go SourceSenderWorker(ctx, out, c)

				cbuf := bufio.NewReader(c)
				quit := false

				for !quit {
					buf, err := Read(cbuf)
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
						fwdIn <- buf
						m := <-pt
						out <- m
					} else {
						buf = buf[lengthSize:]
						msg := CSNQ(&buf)
						if msg != nil {
							out <- *msg
						}
					}

					select {
					case <-ctx.Done():
						log.Println("cancellation received")
						quit = true
					default:
					}
				}
			}(ctx, conn)
		}
	}(ctx)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	s := <-interrupt
	log.Printf("received signal: %s", s.String())
	cancel()
}
