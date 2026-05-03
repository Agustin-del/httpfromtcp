package server

import (
	"bytes"
	"fmt"
	"io"
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

type Handler func(w io.Writer, req *request.Request) *HandlerError

func (hE *HandlerError) WriteError(w io.Writer) {
	response.WriteStatusLine(w, hE.StatusCode)
	hs := response.GetDefaultHeaders(len(hE.Msg))
	response.WriteHeaders(w, hs)
	_, err := w.Write([]byte(hE.Msg))
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

	req, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.INTERNAL_SERVER_ERROR,
			Msg: "error reading request",
		}
		hErr.WriteError(conn)
		return
	}

	writer := bytes.NewBuffer([]byte{})
	hE := s.handler(writer, req)
	if hE != nil {
		hE.WriteError(conn)
		return
	}

	if err := response.WriteStatusLine(conn, response.OK); err != nil {
		hErr := &HandlerError{
			StatusCode: response.INTERNAL_SERVER_ERROR,
			Msg: "error writing status line",
		}
		hErr.WriteError(conn)
		return
	}

	body := writer.Bytes()
	hs := response.GetDefaultHeaders(len(body))
	if err := response.WriteHeaders(conn, hs); err != nil {
		hErr := &HandlerError{
			StatusCode: response.INTERNAL_SERVER_ERROR,
			Msg: "error writing headers",
		}
		hErr.WriteError(conn)
		return
	}

	if _, err := conn.Write(body); err != nil {
		hErr := &HandlerError{
			StatusCode: response.INTERNAL_SERVER_ERROR,
			Msg: "error writing body",
		}
		hErr.WriteError(conn)
		return
	}
}
