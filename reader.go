package main

import (
	"bytes"
	"io"
	"net"
	"strconv"
)

func Read(conn net.Conn) (b []byte, err error) {
	var (
		buf bytes.Buffer
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

	//log.Println(buf.String())
	return buf.Bytes(), nil
}
