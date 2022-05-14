package main

import (
	"reflect"
	"testing"
)

func TestRequestX2J(t *testing.T) {
	tests := []struct {
		name    string
		req     []byte
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{"request",
			[]byte("<XML><MessageType>0</MessageType><ProcCode>CSNQ</ProcCode><REFNUM>0220000245250</REFNUM><STAN>0220000245250</STAN><LocalTxnDtTime>2203221157</LocalTxnDtTime><DeliveryChannelCtrlID>ATM</DeliveryChannelCtrlID><PName>ACCOUNTNUMBER</PName><PValue>157336</PValue></XML>"),
			[]byte(`{"RequestInfo":{"requestId":"0220000245250","userId":"0220000245250","basenumber":"157336","chanelId":"ATM","requestTime":"2203221157"},"searchparametername":"Baseno","searchparametervalue":"157336"}`),
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RequestX2J(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequestX2J() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RequestX2J() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func BenchmarkRequestX2J(b *testing.B) {
	tests := []struct {
		name    string
		req     []byte
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{"request",
			[]byte("<XML><MessageType>0</MessageType><ProcCode>CSNQ</ProcCode><REFNUM>0220000245250</REFNUM><STAN>0220000245250</STAN><LocalTxnDtTime>2203221157</LocalTxnDtTime><DeliveryChannelCtrlID>ATM</DeliveryChannelCtrlID><PName>ACCOUNTNUMBER</PName><PValue>157336</PValue></XML>"),
			[]byte(`{"RequestInfo":{"requestId":"0220000245250","userId":"0220000245250","basenumber":"157336","chanelId":"ATM","requestTime":"2203221157"},"searchparametername":"Baseno","searchparametervalue":"157336"}`),
			false},
	}
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				RequestX2J(tt.req)
			}
		})
	}
}
