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

	request := []byte(`00264<XML><MessageType>0</MessageType><ProcCode>CSNQ</ProcCode><REFNUM>0220000245250</REFNUM><STAN>0220000245250</STAN><LocalTxnDtTime>2203221157</LocalTxnDtTime><DeliveryChannelCtrlID>ATM</DeliveryChannelCtrlID><PName>ACCOUNTNUMBER</PName><PValue>157336</PValue></XML>`)

	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	tick := time.NewTicker(1 * time.Millisecond)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(15*time.Second))
	quit := false

	start := time.Now()
	counter := 0
	for !quit {
		select {
		case <-ctx.Done():
			log.Println(ctx.Err())
			tick.Stop()
			quit = true
		case <-tick.C:
			log.Print("Sending...")
			start := time.Now()
			err := Sender(ctx, conn, request)
			fmt.Println("received in: ", time.Since(start))
			counter++
			if err != nil {
				log.Println(err)
				tick.Stop()
				quit = true
			}
		}
	}
	cancel()
	log.Println("quiting sender mock")
	log.Printf("Run time: %s processed: %d\n", time.Since(start), counter)
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
