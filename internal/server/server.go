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
		msg := "error reading request"
		response.WriteStatusLine(conn, response.BAD_REQUEST)
		hs := response.GetDefaultHeaders(len(msg))
		response.WriteHeaders(conn, hs)
		if _, err := conn.Write([]byte(msg)); err != nil {
			log.Printf("error writing error body\n")
		}
		log.Printf("error reading request: %v", err)
		return
	}

	writer := bytes.NewBuffer([]byte{})
	hE := s.handler(writer, req)
	if hE != nil {
		hE.WriteError(conn)
		return
	}

	if err := response.WriteStatusLine(conn, response.OK); err != nil {
		msg := "error writing status line"
		hs := response.GetDefaultHeaders(len(msg))
		response.WriteStatusLine(conn, response.INTERNAL_SERVER_ERROR)
		response.WriteHeaders(conn, hs)
		if _, err := conn.Write([]byte(msg)); err != nil {
			log.Printf("error writing error body\n")
		}
		log.Printf("Write status line error: %v", err)
		return
	}

	body := writer.Bytes()
	hs := response.GetDefaultHeaders(len(body))
	if err := response.WriteHeaders(conn, hs); err != nil {
		msg := "error writing headers"
		hs := response.GetDefaultHeaders(len(msg))
		response.WriteStatusLine(conn, response.INTERNAL_SERVER_ERROR)
		response.WriteHeaders(conn, hs)
		if _, err := conn.Write([]byte(msg)); err != nil {
			log.Printf("error writing error body\n")
		}
		log.Printf("Write headers error: %v", err)
		return
	}

	if _, err := conn.Write(body); err != nil {
		msg := "error writing body"
		hs := response.GetDefaultHeaders(len(msg))
		response.WriteStatusLine(conn, response.INTERNAL_SERVER_ERROR)
		response.WriteHeaders(conn, hs)
		log.Printf("error writing body: %v", err)
		return
	}
}
