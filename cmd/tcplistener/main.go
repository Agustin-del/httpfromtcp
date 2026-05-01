package main

import (
	"fmt"
	"net"

	"github.com/Agustin-del/httpfromtcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer listener.Close()

	for {

		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("connection accepted\n")

		request, err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Printf("error procesando la request: %v", err)
		}

		fmt.Printf(`Request line:
- Method: %s
- Target: %s
- Version: %s
`+"\n", request.RequestLine.Method,
			request.RequestLine.RequestTarget,
			request.RequestLine.HttpVersion)
		fmt.Printf("Header: \n")

		request.Headers.Iterate(func(k, v string) {
			fmt.Printf("- %s : %s\n", k, v)
		})

		fmt.Printf(`Body:
%s`, string(request.Body))

	}

}
