package main

import (
	"context"
	"flag"
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
	clientPool sync.Pool
	DestURL    *url.URL

	ListenAddr  = flag.String("l", ":3000", "address to listen to")
	DestPtr     = flag.String("d", "http://localhost:3030/", "HTTP destination endpoint")
	Username    = flag.String("u", "ecms", "user name, mandatory")
	Password    = flag.String("s", "ecms1", "user password, mandatory")
	ForwardAddr = flag.String("f", ":9002", "address to passthrough")
)

const lengthSize int = 5

func main() {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	flag.Parse()

	if *Username == "" || *Password == "" {
		log.Fatal("username and password must be provided")
	}

	DestURL, err = url.Parse(*DestPtr)
	if err != nil {
		log.Fatalln(err)
	}

	clientPool = sync.Pool{
		New: func() any {
			return resty.New().EnableTrace().
				SetHeader("User-Agent", "go-frwd/0.0.1").
				SetHeader("Content-Type", "application/json").
				SetBasicAuth(*Username, *Password)
		},
	}

	if *DestPtr == "" {
		log.Fatalln("destination http endpoint must be required")
	}

	host, port, err := net.SplitHostPort(*ListenAddr)
	if err != nil {
		log.Fatalf("incoming listen adress is invalid: %s", err)
	}

	_, _, err = net.SplitHostPort(*ForwardAddr)
	if err != nil {
		log.Fatalf("forward adress is invalid: %s", err)
	}

	l, err := net.Listen("tcp", *ListenAddr)
	if err != nil {
		log.Fatalln(err)
	}
	defer l.Close()
	log.Printf("Listening on host: %s, port: %s\n", host, port)

	go Accepter(ctx, l)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	s := <-interrupt
	log.Printf("received signal: %s", s.String())
	cancel()
}
