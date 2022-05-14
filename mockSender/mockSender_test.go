package main

import (
	"context"
	"net"
	"testing"
)

func BenchmarkSender(b *testing.B) {
	tests := []struct {
		name string
		r    []byte
	}{
		{
			"CSNQ request",
			[]byte(`00264<XML><MessageType>0</MessageType><ProcCode>CSNQ</ProcCode><REFNUM>0220000245250</REFNUM><STAN>0220000245250</STAN><LocalTxnDtTime>2203221157</LocalTxnDtTime><DeliveryChannelCtrlID>ATM</DeliveryChannelCtrlID><PName>ACCOUNTNUMBER</PName><PValue>157336</PValue></XML>`),
		},
		{
			"CRNQ request",
			[]byte(`00264<XML><MessageType>0</MessageType><ProcCode>CRNQ</ProcCode><REFNUM>0220000245250</REFNUM><STAN>0220000245250</STAN><LocalTxnDtTime>2203221157</LocalTxnDtTime><DeliveryChannelCtrlID>ATM</DeliveryChannelCtrlID><PName>ACCOUNTNUMBER</PName><PValue>157336</PValue></XML>`),
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			ctx, cancel := context.WithCancel(context.Background())

			conn, err := net.Dial("tcp", ":3000")
			if err != nil {
				b.Fatal(err)
			}
			defer conn.Close()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Sender(ctx, conn, tt.r)
			}
			cancel()
		})
	}
}
