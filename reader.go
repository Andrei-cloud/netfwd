package main

import (
	"bytes"
	"io"
	"log"
	"net"
	"strconv"
)

// func Read(conn net.Conn) ([]byte, error) {
// 	var (
// 		err error
// 		l   int
// 	)

// 	netLen := make([]byte, 5)
// 	//conn.SetReadDeadline(time.Now().Add(time.Second))
// 	for {
// 		n, err := conn.Read(netLen)
// 		if err != nil {
// 			return nil, err
// 		}
// 		if n == 5 {
// 			l, err = strconv.Atoi(string(netLen))
// 			if err != nil {
// 				return nil, fmt.Errorf("invalid msg length received: %w", err)
// 			}
// 			break
// 		}
// 	}

// 	buf := make([]byte, l)
// 	_, err = conn.Read(buf)
// 	if err != nil {
// 		return nil, err
// 	}
// 	data := append(netLen, buf...)
// 	log.Println(string(data))
// 	return data, nil
// }

func Read(conn net.Conn) ([]byte, error) {
	var (
		buf bytes.Buffer
		err error
		l   int
	)

	_, err = io.CopyN(&buf, conn, int64(lengthSize))
	if err != nil {
		return nil, err
	}

	l, err = strconv.Atoi(buf.String())
	if err != nil {
		return nil, err
	}

	_, err = io.CopyN(&buf, conn, int64(l))
	if err != nil {
		return nil, err
	}

	log.Println(buf.String())
	return buf.Bytes(), nil
}
