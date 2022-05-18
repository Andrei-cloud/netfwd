package main

import (
	"context"
	"net"
	"testing"
)

func Test_connectionHandler(t *testing.T) {
	type args struct {
		ctx  context.Context
		conn net.Conn
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connectionHandler(tt.args.ctx, tt.args.conn)
		})
	}
}
