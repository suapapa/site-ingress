package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/pkg/errors"
)

func startPortFoward() error {
	err := updateLinks()
	if err != nil {
		return errors.Wrap(err, "fail to start port foward")
	}

	// TODO: currently its not prepared for key update
	cert, err := tls.LoadX509KeyPair(SSL_CERT_FILE, SSL_KEY_FILE)
	if err != nil {
		return errors.Wrap(err, "fail to start port foward")
	}
	tc := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	for _, l := range links {
		if l.PortFoward {
			dest := fmt.Sprintf("%s:%d", l.Link, l.OutPort)
			log.Printf("start portfoward, %s. %d->%s", l.Name, l.InPort, dest)
			go runPortFoward(l.InPort, dest, tc)
		}
	}
	return nil
}

func runPortFoward(inPort int, dest string, tc *tls.Config) {
	log.Printf("listening on port, %d", inPort)
	l, err := tls.Listen("tcp", fmt.Sprintf(":%d", inPort), tc)
	if err != nil {
		log.Printf("ERR: %v", err)
		return
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("ERR: %v", err)
			return
		}

		go handlePortFoward(conn, dest)
	}
}

func handlePortFoward(conn net.Conn, dest string) {
	log.Printf("staring port-foward to %s", dest)
	outConn, err := net.Dial("tcp", dest)
	if err != nil {
		log.Printf("ERR: %v", err)
		return
	}
	go copyIO(conn, outConn)
	go copyIO(outConn, conn)
}

func copyIO(src, dest net.Conn) {
	defer src.Close()
	defer dest.Close()
	io.Copy(src, dest)
}
