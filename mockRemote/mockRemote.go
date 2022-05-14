package main

import (
	"io"
	"log"
	"net"
)

func main() {
	l, err := net.Listen("tcp", ":9002")
	if err != nil {
		panic(err)
	}
	defer l.Close()
	log.Println("Listening on " + l.Addr().String())

	for {
		// Listen for an incoming connection
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		log.Println("incoming connection established")

		go func(c net.Conn) {
			defer func() {
				c.Close()
			}()
			buf := make([]byte, 4096)

			for {
				n, err := c.Read(buf)
				if err != nil {
					if err == io.EOF {
						log.Println("remote connection is closed")
						return
					}
					log.Printf("error reading: %s", err)
					return
				}
				if n > 0 {
					l, err := c.Write(buf[:n])
					if err != nil {
						log.Printf("error writing: %s", err)
						return
					}
					log.Printf("echoed %d bytes from %s\n", l, c.RemoteAddr().String())
				}
			}
		}(conn)
	}
}
