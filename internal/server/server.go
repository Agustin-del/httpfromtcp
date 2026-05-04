package server

import (
	"fmt"
	"log"
	"net"

	"github.com/Agustin-del/httpfromtcp/internal/request"
	"github.com/Agustin-del/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	handler  Handler
}

type HandlerError struct {
	StatusCode response.StatusCode
	Msg        string
}

type Handler func(w *response.Writer, req *request.Request)

func (hE *HandlerError) Write(w *response.Writer) {
	w.WriteStatusLine(hE.StatusCode)
	hs := response.GetDefaultHeaders(len(hE.Msg), false)
	w.WriteHeaders(hs)
	_, err := w.WriteBody([]byte(hE.Msg))
	if err != nil {
		fmt.Printf("error writing error")
	}
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: listener,
		handler:  handler,
	}

	go server.listen()

	return server, nil
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

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	writer := response.NewWriter(conn)

	req, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.INTERNAL_SERVER_ERROR,
			Msg: "error reading request",
		}
		hErr.Write(writer)
		return
}

	s.handler(writer, req)
}
