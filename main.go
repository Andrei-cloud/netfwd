package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Application-wide constants
const (
	lengthSize = 5 // Size of message length prefix
)

// Command line flags
var (
	DestURL     *url.URL
	ListenAddr  = flag.String("l", ":3000", "address to listen to")
	DestPtr     = flag.String("d", "http://localhost:3030/", "HTTP destination endpoint")
	Username    = flag.String("u", "ecms", "user name, mandatory")
	Password    = flag.String("s", "ecms1", "user password, mandatory")
	ForwardAddr = flag.String("f", ":9002", "address to passthrough")
)

func main() {
	// defer profile.Start(profile.MemProfile).Stop()

	// Set up text logger with slog
	textHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(textHandler)
	slog.SetDefault(logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	flag.Parse()

	if err := checkInit(); err != nil {
		slog.Error("Initialization error", "error", err)
		os.Exit(1)
	}

	l, err := net.Listen("tcp", *ListenAddr)
	if err != nil {
		slog.Error("Failed to listen", "address", *ListenAddr, "error", err)
		os.Exit(1)
	}
	slog.Info("Listening", "host", *ListenAddr)

	go Accepter(ctx, l)

	// Handle graceful shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	s := <-interrupt
	slog.Info("Received signal", "signal", s.String())
	cancel()
	time.Sleep(time.Second) // Allow time for cleanup
}

// checkInit validates command line arguments and initializes global variables
func checkInit() error {
	// Check mandatory fields
	if *Username == "" || *Password == "" {
		return errors.New("username and password must be provided")
	}

	if *DestPtr == "" {
		return errors.New("destination HTTP endpoint is required")
	}

	// Parse and validate destination URL
	var err error
	DestURL, err = url.Parse(*DestPtr)
	if err != nil {
		return fmt.Errorf("invalid destination URL: %w", err)
	}

	// Validate listen address
	if _, _, err = net.SplitHostPort(*ListenAddr); err != nil {
		return fmt.Errorf("incoming listen address is invalid: %w", err)
	}

	// Validate forward address
	if _, _, err = net.SplitHostPort(*ForwardAddr); err != nil {
		return fmt.Errorf("forward address is invalid: %w", err)
	}

	return nil
}
