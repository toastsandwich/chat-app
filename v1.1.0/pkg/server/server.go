package server

import (
	"bufio"
	"fmt"
	"net"
)

type Server struct {
	addr       string
	clients    map[net.Conn]struct{}
	addClients chan net.Conn
	rmClients  chan net.Conn
	messages   chan string
}

func NewServer(addr string) *Server {
	return &Server{
		addr:       addr,
		clients:    make(map[net.Conn]struct{}),
		addClients: make(chan net.Conn),
		rmClients:  make(chan net.Conn),
		messages:   make(chan string),
	}
}

func (s *Server) Start() error {
	fmt.Println("[*] Starting Server....")
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	fmt.Println("[+] Server Up!!!")
	defer ln.Close()

	go s.handleMessages()
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("connection error skipping this client")
			return err
		}
		go s.HandleConn(conn)
		s.addClients <- conn
		s.messages <- fmt.Sprintf("%s has arrived\n", conn.RemoteAddr().String())
	}
}

func (s *Server) HandleConn(conn net.Conn) {
	defer func() {
		s.rmClients <- conn
		conn.Close()
	}()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	for {
		writer.Write([]byte(":::"))
		mssg, _ := reader.ReadString('\n')
		s.messages <- mssg
		fmt.Println(mssg)
		if mssg == "bye\n" {
			s.messages <- fmt.Sprintf("%s has left the chat\n", conn.RemoteAddr().String())
			return
		}
	}
}

func (s *Server) handleMessages() {
	for {
		select {
		case conn := <-s.addClients:
			s.clients[conn] = struct{}{}
			fmt.Println(s.clients)
		case conn := <-s.rmClients:
			delete(s.clients, conn)
		case mssg := <-s.messages:
			for conn := range s.clients {
				writer := bufio.NewWriter(conn)
				msg := fmt.Sprintf("%s :: %s\n", conn.RemoteAddr().String(), mssg)
				_, err := writer.Write([]byte(msg))
				if err != nil {
					conn.Close()
					s.rmClients <- conn
				}
			}
		}
	}
}
