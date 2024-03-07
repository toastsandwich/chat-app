package main

import (
	"fmt"
	"net"
)

const (
	addr = "192.168.1.106:8647"
	tcp  = "tcp"
)

type server struct {
	clients    map[net.Conn]bool
	addClients chan net.Conn
	rmClients  chan net.Conn
	messages   chan string
}

func newServer() *server {
	return &server{
		clients:    make(map[net.Conn]bool),
		addClients: make(chan net.Conn),
		rmClients:  make(chan net.Conn),
		messages:   make(chan string),
	}
}

func (s *server) handleConnection(conn net.Conn) {
	defer func() {
		s.rmClients <- conn
		conn.Close()
	}()
	buf := make([]byte, 2048)
	for {
		n, _ := conn.Read(buf)
		mssg := string(buf[:n])
		s.messages <- mssg
	}
}

func (s *server) handleMessages() {
	for {
		select {
		case conn := <-s.rmClients:
			delete(s.clients, conn)
		case conn := <-s.addClients:
			s.clients[conn] = true
		case mssg := <-s.messages:
			for conn := range s.clients {
				_, err := conn.Write([]byte(fmt.Sprintf("%s :: %s\n", conn.RemoteAddr().String(), mssg)))
				if err != nil {
					s.rmClients <- conn
					conn.Close()
				}
			}
		}
	}
}

func (s *server) start() error {
	ln, err := net.Listen(tcp, addr)
	defer func() {
		err := ln.Close()
		if err != nil {
			return
		}
	}()
	if err != nil {
		return err
	}
	go s.handleMessages()
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		s.addClients <- conn
		go s.handleConnection(conn)
	}
}

func main() {
	app := newServer()
	fmt.Println("server hosted on ", addr)
	// log.Fatal(app.start())
	app.start()
}
