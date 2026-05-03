package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"github.com/Agustin-del/httpfromtcp/internal/request"
	"github.com/Agustin-del/httpfromtcp/internal/response"
	"github.com/Agustin-del/httpfromtcp/internal/server"
)

const port = 42069

func handler(w io.Writer, req *request.Request) *server.HandlerError {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		hE := &server.HandlerError{
			StatusCode: response.BAD_REQUEST,
			Msg:"Your problem is not my problem\n",
		}

		return hE

	case "/myproblem":
		hE := &server.HandlerError{
			StatusCode: response.INTERNAL_SERVER_ERROR,
			Msg: "Woopsie, my bad\n",
		}

		return hE
	default:
		w.Write([]byte("All good, frfr\n"))
		return nil
	}
}

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stoped")
}
