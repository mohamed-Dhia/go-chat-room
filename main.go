package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
)

type Msg struct {
	from    net.Addr
	payload []byte
}

type Server struct {
	listenAddr string
	listener   net.Listener
	quitChan   chan struct{}
	msgChan    chan Msg
	peerMap    map[net.Addr]net.Conn
}

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitChan:   make(chan struct{}),
		msgChan:    make(chan Msg, 10),
		peerMap:    make(map[net.Addr]net.Conn),
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}

	defer listener.Close()

	s.listener = listener

	go s.acceptLoop()

	<-s.quitChan
	close(s.msgChan)

	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Println("accept error: ", err)
			continue
		}

		remoteAddr := conn.RemoteAddr()
		s.peerMap[remoteAddr] = conn

		fmt.Println("new connection from: ", conn.RemoteAddr())

		conn.Write([]byte("Welcome to the chat!\n"))

		go s.readLoop(conn)
		go s.broadcastMsg()
	}
}

func (s *Server) readLoop(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 2048)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("read error: ", err)
			continue
		}

		s.msgChan <- Msg{payload: buf[:n], from: conn.RemoteAddr()}

	}
}

func (s *Server) broadcastMsg() {
	for msg := range s.msgChan {
		payload := bytes.TrimSuffix(msg.payload, []byte{13, 10})
		fmt.Printf("received msg from connection(%s): %s\n", msg.from.String(), msg.payload)

		for peer := range s.peerMap {
			if peer == msg.from {
				continue
			}
			msg := fmt.Sprintf("(%s) says: \"%s\"\n", msg.from, payload)
			s.peerMap[peer].Write([]byte(msg))
		}
	}
}

func main() {
	server := NewServer(":3000")

	log.Fatal(server.Start())
}
