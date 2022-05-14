package main

import (
	"bufio"
	"context"
	"io"
	"log"
	"net"
	"strconv"
	"time"
)

func main() {
	log.Println("starting sender mock")

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	quit := false

	request := []byte(`00264<XML><MessageType>0</MessageType><ProcCode>CRNQ</ProcCode><REFNUM>0220000245250</REFNUM><STAN>0220000245250</STAN><LocalTxnDtTime>2203221157</LocalTxnDtTime><DeliveryChannelCtrlID>ATM</DeliveryChannelCtrlID><PName>ACCOUNTNUMBER</PName><PValue>157336</PValue></XML>`)

	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	tick := time.NewTicker(time.Second)

	for !quit {
		select {
		case <-ctx.Done():
			log.Println(ctx.Err())
			tick.Stop()
			quit = true
		case <-tick.C:
			log.Println("Sending...")
			Sender(ctx, conn, request)
		default:
		}
	}
	cancel()
	log.Println("quiting sender mock")
}

func Sender(ctx context.Context, conn net.Conn, request []byte) {
	var l int
	buf := make([]byte, 4096)

	_, err := conn.Write(request)
	if err != nil {
		log.Println(err)
		return
	}

	b := bufio.NewReader(conn)

	size, err := b.Peek(5)
	if err != nil {
		if err == io.EOF {
			return
		}
		log.Println(err)
	}
	l, err = strconv.Atoi(string(size))
	if err != nil {
		log.Println(err)
	}

	_, err = b.Read(buf[:5+l])
	if err != nil {
		log.Println(err)
	}
}
