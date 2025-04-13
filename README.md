# Netfwd - Network Forwarding Service

Netfwd is a high-performance TCP-to-HTTP bridge application that accepts TCP connections and routes messages either through HTTP APIs or by forwarding them to another TCP endpoint based on message content.

## Overview

Netfwd acts as a middleware that:

1. Listens for incoming TCP connections
2. Analyzes messages to determine their type
3. Routes messages based on their type:
   - CSNQ messages are transformed from XML to JSON and sent to an HTTP API endpoint
   - Other messages are forwarded directly to a remote TCP endpoint
4. Returns responses back to the original client

The service is designed for high-performance operation with support for concurrent connections and parallel processing of API requests.

## Features

- TCP message proxying
- Protocol transformation (XML to JSON and back)
- Content-based routing
- Concurrent connection handling
- Parallel API request processing
- Graceful shutdown on interruption

## Architecture

The application follows a worker-based architecture with the following components:

- **Accepter**: Accepts incoming TCP connections and creates handlers for each
- **Connection Handler**: Routes messages based on content analysis
- **Proxy Worker**: Forwards messages to remote TCP endpoints
- **API Worker**: Transforms and forwards messages to HTTP endpoints
- **Source Sender Worker**: Sends responses back to the original clients

## Installation

### Prerequisites

- Go 1.18 or later

### Building from Source

```bash
git clone https://github.com/andrei-cloud/netfwd.git
cd netfwd
go build
```

## Usage

```bash
./netfwd [options]
```

### Options

```
-l string   Address to listen on (default ":3000")
-d string   HTTP destination endpoint (default "http://localhost:3030/")
-u string   Username for HTTP authentication (default "ecms")
-s string   Password for HTTP authentication (default "ecms1")
-f string   Address to pass through non-CSNQ messages (default ":9002")
```

### Example

Start the service with custom settings:

```bash
./netfwd -l :8080 -d https://api.example.com/endpoint -u user -s pass -f :9000
```

## Test Utilities

The project includes several mock applications for testing:

- **mockRemote**: Simulates a remote TCP endpoint that echoes messages
- **mockSender**: Simulates a client sending messages to the service
- **mockWeb**: Simulates an HTTP API endpoint

Run these utilities in separate terminal sessions:

```bash
# Start the mock remote server
go run mockRemote/mockRemote.go

# Start the mock web server
go run mockWeb/mockWeb.go

# Start netfwd
go run .

# Run the mock sender to test
go run mockSender/mockSender.go
```

## Message Flow

1. TCP client connects to netfwd
2. Client sends a message
3. Netfwd analyzes the message:
   - If the message contains "CSNQ", it's processed through the API path
   - Otherwise, it's forwarded to the remote TCP endpoint
4. Processing path:
   - API path: XML → JSON → HTTP request → JSON response → XML
   - TCP path: Direct forwarding
5. Response is sent back to the client

## Performance Benchmarks

The codebase includes benchmarks for:
- Message transformation (XML ↔ JSON)
- Proxy performance
- End-to-end performance

Run benchmarks with:

```bash
go test -bench=.
```

## Development

### Project Structure

- **main.go**: Entry point and configuration
- **handlers.go**: Connection handling and message routing
- **workers.go**: Worker implementations
- **request.go/response.go**: Message transformation
- **reader.go**: Low-level socket reading
- **csnq.go**: API client implementation
- **mock* directories**: Test utilities
