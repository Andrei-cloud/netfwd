package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//to minimize cleint coldstart
var (
	DestURL *url.URL

	ListenAddr  = flag.String("l", ":3000", "address to listen to")
	DestPtr     = flag.String("d", "http://localhost:3030/", "HTTP destination endpoint")
	Username    = flag.String("u", "ecms", "user name, mandatory")
	Password    = flag.String("s", "ecms1", "user password, mandatory")
	ForwardAddr = flag.String("f", ":9002", "address to passthrough")
)

const lengthSize int = 5

func main() {
	//defer profile.Start(profile.MemProfile).Stop()

	var err error
	ctx, cancel := context.WithCancel(context.Background())
	flag.Parse()

	err = checkInit()
	if err != nil {
		log.Fatal(err)
	}

	l, err := net.Listen("tcp", *ListenAddr)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Listening on host: %s\n", *ListenAddr)

	go Accepter(ctx, l)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	s := <-interrupt
	log.Printf("received signal: %s", s.String())
	cancel()
	time.Sleep(time.Second)
}

func checkInit() error {
	var err error

	if *Username == "" || *Password == "" {
		return errors.New("username and password must be provided")
	}

	if *DestPtr == "" {
		return fmt.Errorf("destination http endpoint must be required")
	}

	DestURL, err = url.Parse(*DestPtr)
	if err != nil {
		return fmt.Errorf("invalid destination handler: %w", err)
	}
	_, _, err = net.SplitHostPort(*ListenAddr)
	if err != nil {
		return fmt.Errorf("incoming listen adress is invalid: %w", err)
	}

	_, _, err = net.SplitHostPort(*ForwardAddr)
	if err != nil {
		return fmt.Errorf("forward adress is invalid: %w", err)
	}

	return nil
}
