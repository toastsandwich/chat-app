package main

import (
	"bufio"
	"log"
	"net"
)

const (
	addr = "localhost:7568"
	tcp  = "tcp"
)

type server struct {
	addr      string
	clients   map[net.Conn]struct{}
	addClient chan net.Conn
	rmClient  chan net.Conn
	message   chan string
}

func newServer(addr string) *server {
	return &server{
		addr:      addr,
		clients:   make(map[net.Conn]struct{}),
		addClient: make(chan net.Conn),
		rmClient:  make(chan net.Conn),
		message:   make(chan string),
	}
}

func (s *server) start() {
	ln, err := net.Listen(tcp, s.addr)
	if err != nil {
		log.Fatal(err)
	}
	go s.handleMessages()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go s.handleConn(conn)
	}
}

func (s *server) handleConn(conn net.Conn) {
	defer func() {
		s.rmClient <- conn
		conn.Close()
	}()
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		s.message <- message
	}
}

func (s *server) handleMessages() {
	for {
		select {
		case conn := <-s.addClient:
			s.clients[conn] = struct{}{}
		case conn := <-s.rmClient:
			delete(s.clients, conn)
		case mssg := <-s.message:
			for conn, _ := range s.clients {
				writer := bufio.NewWriter(conn)
				writer.WriteString(mssg)
			}
		}
	}
}

func main() {
	app := newServer(addr)
	app.start()
}
