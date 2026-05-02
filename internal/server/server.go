package server

import (
	"fmt"
	"log"
	"net"

	"github.com/Agustin-del/httpfromtcp/internal/response"
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
			s.handle(conn)
		}()
	}
}

func (s *Server) handle(conn net.Conn) {
	err := response.WriteStatusLine(conn, 200)
	if err != nil {
		log.Printf("Write status line error: %v", err)
	}

	hs := response.GetDefaultHeaders(0)
	err = response.WriteHeaders(conn, hs) 
	if err != nil {
		log.Printf("Write headers error: %v", err)
	}
	conn.Close()
}
