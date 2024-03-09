package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type Server struct {
	addr      string
	clients   map[net.Conn]bool
	addClient chan net.Conn
	rmClient  chan net.Conn
	messages  chan string
}

func NewServer(addr string) *Server {
	return &Server{
		addr:      addr,
		clients:   make(map[net.Conn]bool),
		addClient: make(chan net.Conn),
		rmClient:  make(chan net.Conn),
		messages:  make(chan string),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	go s.handleMessages()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accepting error")
			continue
		}
		s.messages <- fmt.Sprintf("%s has joined the room\n", conn.RemoteAddr().String())
		s.addClient <- conn
		go s.handleConnections(conn)
	}
}

func (s *Server) handleConnections(conn net.Conn) {
	defer func() {
		s.rmClient <- conn
		conn.Close()
	}()
	reader := bufio.NewReader(conn)
	for {
		mssg, _ := reader.ReadString('\n')
		s.messages <- mssg

		if mssg == "bye\n" {
			return
		}
	}

}

func (s *Server) handleMessages() {
	for {
		select {
		case conn := <-s.addClient:
			s.clients[conn] = true
		case conn := <-s.rmClient:
			delete(s.clients, conn)
		case mssg := <-s.messages:
			for conn := range s.clients {
				_, err := conn.Write([]byte(mssg))
				if err != nil {
					s.rmClient <- conn
					conn.Close()
				}
			}
		}
	}
}
