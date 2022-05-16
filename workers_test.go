package main

import (
	"context"
	"io"
	"log"
	"net"
	"testing"
)

func setupRemote() context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	l, err := net.Listen("tcp", ":9009")
	if err != nil {
		panic(err)
	}

	go func(ctx context.Context) {
		defer l.Close()
		for {
			// Listen for an incoming connection
			conn, err := l.Accept()
			if err != nil {
				log.Fatal(err)
			}
			defer conn.Close()

			go func(c net.Conn) {
				defer func() {
					c.Close()
				}()

				n, err := io.Copy(c, c)
				if err != nil {
					return
				}
				log.Printf("echoed %d bytes from %s\n", n, c.RemoteAddr().String())
			}(conn)

			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}(ctx)
	return cancel
}

// func TestPassThroughWorker(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		msg  []byte
// 		want []byte
// 	}{
// 		{
// 			"simple message",
// 			[]byte("hello"),
// 			[]byte("00011hello world"),
// 		},
// 		{
// 			"real message",
// 			[]byte(`00264<XML><MessageType>0</MessageType><ProcCode>CSNQ</ProcCode><REFNUM>0220000245250</REFNUM><STAN>0220000245250</STAN><LocalTxnDtTime>2203221157</LocalTxnDtTime><DeliveryChannelCtrlID>ATM</DeliveryChannelCtrlID><PName>ACCOUNTNUMBER</PName><PValue>157336</PValue></XML>`),
// 			[]byte(`01129<XML><MessageType>1</MessageType><ProcCode>CSNQ</ProcCode><STAN>0220000245250</STAN><LocalTxnDtTime>2203221157</LocalTxnDtTime><DeliveryChannelCtrlID>ATM</DeliveryChannelCtrlID><PName>ACCOUNTNUMBER</PName><PValue>157336</PValue><ActCode>0</ActCode><ActDescription>Success</ActDescription><TotalnoofTrans>1</TotalnoofTrans><Customers><Record><Name>GEORGE ONYANGO OYANGE</Name><FirstName>GEORGE ONYANGO</FirstName><MiddleName></MiddleName><LastName>OYANGE</LastName><BaseNumber>157336</BaseNumber><Nationality></Nationality><PoBox></PoBox><Address></Address><City></City><Country></Country><Email>oyange.george73@gmail.com</Email><CardOnlyCustomer></CardOnlyCustomer><SMSMobile></SMSMobile><SMSLang></SMSLang><SMSNationalID></SMSNationalID><SMSPassportNo></SMSPassportNo><SegmentCode></SegmentCode><SegmentDesc></SegmentDesc><QID>27340400282</QID><QIDExpiryDate></QIDExpiryDate><PassportNo>A1810149</PassportNo><PassportExpiryDate></PassportExpiryDate><CompanyRegNo></CompanyRegNo><CompanyRegNoExpiryDate></CompanyRegNoExpiryDate><LOB></LOB><DOB></DOB><CustTypeFlag></CustTypeFlag></Record></Customers><REFNUM>256557</REFNUM></XML>`),
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctx, cancel := context.WithCancel(context.Background())

// 			buffer := bytes.NewBuffer(tt.want)
// 			r := bufio.NewReader(buffer)
// 			w := bufio.NewWriter(buffer)

// 			remote := bufio.NewReadWriter(r, w)

// 			inMsg := make(chan []byte, 1)

// 			outMsg, _ := ProxyWorker(ctx, inMsg, remote)
// 			inMsg <- tt.msg
// 			if got := <-outMsg; !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("PassThroughWorker() = %v, want %v", got, tt.want)
// 			}
// 			cancel()
// 			close(inMsg)
// 		})
// 	}
// }

func BenchmarkPassThroughWorker(b *testing.B) {
	tests := []struct {
		name string
		msg  []byte
		want []byte
	}{
		{
			"real message",
			[]byte(`00264<XML><MessageType>0</MessageType><ProcCode>CSNQ</ProcCode><REFNUM>0220000245250</REFNUM><STAN>0220000245250</STAN><LocalTxnDtTime>2203221157</LocalTxnDtTime><DeliveryChannelCtrlID>ATM</DeliveryChannelCtrlID><PName>ACCOUNTNUMBER</PName><PValue>157336</PValue></XML>`),
			[]byte(`1102<XML><MessageType>1</MessageType><ProcCode>CSNQ</ProcCode><STAN>0220000245250</STAN><LocalTxnDtTime>2203221157</LocalTxnDtTime><DeliveryChannelCtrlID>ATM</DeliveryChannelCtrlID><PName>ACCOUNTNUMBER</PName><PValue>157336</PValue><ActCode>0</ActCode><ActDescription>Success</ActDescription><TotalnoofTrans>1</TotalnoofTrans><Customers><Record><Name>IVAN IVANOV</Name><FirstName>IVAN</FirstName><MiddleName></MiddleName><LastName>IVANOV</LastName><BaseNumber>157336</BaseNumber><Nationality></Nationality><PoBox></PoBox><Address></Address><City></City><Country></Country><Email>example@example.com</Email><CardOnlyCustomer></CardOnlyCustomer><SMSMobile></SMSMobile><SMSLang></SMSLang><SMSNationalID></SMSNationalID><SMSPassportNo></SMSPassportNo><SegmentCode></SegmentCode><SegmentDesc></SegmentDesc><QID>273XXXXXXXX</QID><QIDExpiryDate></QIDExpiryDate><PassportNo>XXXXXXX</PassportNo><PassportExpiryDate></PassportExpiryDate><CompanyRegNo></CompanyRegNo><CompanyRegNoExpiryDate></CompanyRegNoExpiryDate><LOB></LOB><DOB></DOB><CustTypeFlag></CustTypeFlag></Record></Customers><REFNUM>256557</REFNUM></XML>`),
		},
	}
	teardown := setupRemote()

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			ctx, cancel := context.WithCancel(context.Background())

			remote, err := net.Dial("tcp", ":9009")
			if err != nil {
				b.Fatal(err)
			}
			inMsg := make(chan []byte, 1)
			outMsg, _ := ProxyWorker(ctx, inMsg, remote)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				inMsg <- tt.msg
				<-outMsg
			}

			cancel()
			close(inMsg)
		})
	}
	teardown()
}
