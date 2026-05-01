package server

import (
	"fmt"
	"log"
	"net"
)

type Server struct {
	listener net.Listener
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d",port))
	if err != nil {
		return nil, err
	}

	server := Server{
		listener: listener,
	}	
	
	go func () {
		server.listen()
	}()

	return &server, nil
}

func (s *Server) Close() error {
	if err := s.listener.Close(); err != nil {
		return err
	}
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("error: aceptando request: %v", err)
			return
		}

		go func() {
			s.handler(conn)
		}()
	}
}

func (s *Server) handler(conn net.Conn) {
	_, err := conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length:13\r\n\r\nHello world!\n"))
	if err != nil {
		log.Printf("Write error: %v", err)
	}
	conn.Close()
}
