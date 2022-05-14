package main

import (
	"bufio"
	"io"
	"strconv"
)

func Read(c *bufio.Reader) ([]byte, error) {
	var (
		size, buf []byte
		err       error
	)
	size, err = c.Peek(lengthSize)
	if err != nil {
		return nil, err
	}

	l, err := strconv.Atoi(string(size))
	if err != nil {
		return nil, err
	}
	buf = make([]byte, lengthSize+l)
	_, err = io.ReadFull(c, buf[:lengthSize+l])
	if err != nil {
		return nil, err
	}

	return buf, nil
}
