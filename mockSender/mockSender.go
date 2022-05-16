package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

const lengthSize = 5

func main() {
	log.Println("starting sender mock")

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
	quit := false

	request := []byte(`00264<XML><MessageType>0</MessageType><ProcCode>CRNQ</ProcCode><REFNUM>0220000245250</REFNUM><STAN>0220000245250</STAN><LocalTxnDtTime>2203221157</LocalTxnDtTime><DeliveryChannelCtrlID>ATM</DeliveryChannelCtrlID><PName>ACCOUNTNUMBER</PName><PValue>157336</PValue></XML>`)

	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	tick := time.NewTicker(10 * time.Millisecond)

	for !quit {
		select {
		case <-ctx.Done():
			log.Println(ctx.Err())
			tick.Stop()
			quit = true
		case <-tick.C:
			log.Println("Sending...")
			err := Sender(ctx, conn, request)
			if err != nil {
				log.Println(err)
				tick.Stop()
				quit = true
			}
		default:
		}
	}
	cancel()
	log.Println("quiting sender mock")
}

func Sender(ctx context.Context, conn net.Conn, request []byte) error {
	var l int

	_, err := conn.Write(request)
	if err != nil {
		return err

	}

	netLen := make([]byte, 5)
	//conn.SetReadDeadline(time.Now().Add(time.Second))
	for {
		n, err := conn.Read(netLen)
		if err != nil {
			return err
		}
		if n == 5 {
			l, err = strconv.Atoi(string(netLen))
			if err != nil {
				return fmt.Errorf("invalid msg length received: %w", err)
			}
			break
		}
	}

	buf := make([]byte, l)
	_, err = conn.Read(buf)
	if err != nil {
		return err
	}
	//data := append(netLen, buf...)
	//log.Println(string(data))

	return nil
}
