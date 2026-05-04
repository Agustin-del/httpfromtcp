package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/Agustin-del/httpfromtcp/internal/headers"
	"github.com/Agustin-del/httpfromtcp/internal/request"
	"github.com/Agustin-del/httpfromtcp/internal/response"
	"github.com/Agustin-del/httpfromtcp/internal/server"
)

const port = 42069
const baseHTTPBin = "https://httpbin.org"

func respond500(w *response.Writer) {

	msg := `<html>
	<head>
		<title>500 Internal Server Error</title>
	</head>
	<body>
		<h1>Internal Server Error</h1>
		<p>Okay, you know what? This one is on me.</p>
	</body>
</html>`
	if err := w.WriteStatusLine(response.INTERNAL_SERVER_ERROR); err != nil {
		log.Printf("error escribiendo el status line: %v", err)
	}
	hs := response.GetDefaultHeaders(len(msg), false)
	hs.Replace("content-type", "text/html")
	w.WriteHeaders(hs)
	w.WriteBody([]byte(msg))
}

func handler(w *response.Writer, req *request.Request) {
	path, found := strings.CutPrefix(req.RequestLine.RequestTarget, "/httpbin")
	if found {

		switch path {
		case "/html":
			res, err := http.Get(baseHTTPBin + path)
			if err != nil {
				respond500(w)
				log.Printf("error en la request /html: %v", err)
				return
			}

			if err := w.WriteStatusLine(200); err != nil {
				respond500(w)
				log.Printf("error escribiendo status line de la request de /html: %v", err)
				return
			}

			hs := response.GetDefaultHeaders(0, true)
			hs.Set("trailer", "x-content-sha256, x-content-length")
			if err := w.WriteHeaders(hs); err != nil {
				respond500(w)
				log.Printf("error escribiendo headers de la request de /html: %v", err)
				return
			}

			totalSize := 0
			hash := sha256.New()
			buf := make([]byte, 1024)
			for {
				bytesReceived, err := res.Body.Read(buf)
				if err != nil {
					if err == io.EOF {
						_, err := w.WriteChunkedBodyDone()
						if err != nil {
							log.Printf("error escribiendo el final del chunked body: %v", err)
							respond500(w)
							return
						}
						hs := headers.NewHeaders()
						err = w.WriteTrailers(*hs)
						if err != nil {
							respond500(w)
							log.Printf("error escribiendo trailers: %v", err)
							return
						}
						break
					}
					respond500(w)
					log.Printf("error receiving response de tercero: %v", err)
					break
				}

				chunk := buf[:bytesReceived]
				if _, err := hash.Write(chunk); err != nil {
					respond500(w)
					log.Printf("error calculating hash response de tercero: %v", err)
					return
				}
				totalSize += len(chunk)

				_, err = w.WriteChunkedBody(chunk)
				if err != nil {
					log.Printf("error writing chunked body: %v", err)
					return
				}
			}

			hashedSum := hash.Sum(nil)
			trailers := headers.NewHeaders()
			trailers.Set("x-content-sha256", fmt.Sprintf("%x", hashedSum))
			trailers.Set("x-content-length", fmt.Sprintf("%d", totalSize))

			w.WriteTrailers(*trailers)

		default:
			res, err := http.Get(baseHTTPBin + path)
			if err != nil {
				respond500(w)
				log.Printf("error llamando a httpbin: %v", err)
				return
			}

			cl := res.Header.Get("content-length")
			if cl != "" {
				cLength, err := strconv.Atoi(cl)
				if err != nil {
					respond500(w)
					log.Printf("error en el content-length de la request al servicio de tercero")
					return
				}

				body := make([]byte, cLength)
				if _, err := res.Body.Read(body); err != nil {
					respond500(w)
					log.Printf("error leyendo el body de la request de tercero: %v", err)
					return
				}

				if err := w.WriteStatusLine(200); err != nil {
					respond500(w)
					log.Printf("error escribiendo status line de la request de tercero: %v", err)
					return
				}

				hs := response.GetDefaultHeaders(cLength, false)
				hs.Replace("content-type", res.Header.Get("content-type"))
				if err := w.WriteHeaders(hs); err != nil {
					respond500(w)
					log.Printf("error escribiendo headers de la request de tercero: %v", err)
					return
				}

				if _, err := w.WriteBody(body); err != nil {
					respond500(w)
					log.Printf("error escribiendo el body de la request de tercero: %v", err)
					return
				}
			} else {

				if err := w.WriteStatusLine(200); err != nil {
					respond500(w)
					log.Printf("error escribiendo status line de la request de tercero: %v", err)
					return
				}

				hs := response.GetDefaultHeaders(0, true)
				if err := w.WriteHeaders(hs); err != nil {
					respond500(w)
					log.Printf("error escribiendo headers de la request de tercero: %v", err)
					return
				}
				buf := make([]byte, 1024)
				totalBytes := 0
				for {
					bytesReceived, err := res.Body.Read(buf)
					if err != nil {
						if err == io.EOF {
							_, err := w.WriteChunkedBodyDone()
							if err != nil {

								respond500(w)
								log.Printf("error writing done chunked: %v", err)
							}
							break
						}
						respond500(w)
						log.Printf("error receiving response de tercero: %v", err)
						break
					}

					totalBytes += bytesReceived

					_, err = w.WriteChunkedBody(buf[:bytesReceived])
					if err != nil {
						log.Printf("error writing chunked body: %v", err)
						return
					}
				}

				trailers := headers.NewHeaders()
				w.WriteTrailers(*trailers)
			}
		}
	} else {

		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			msg := `<html>
	<head>
		<title>400 Bad Request</title>
	</head>
	<body>
		<h1>Bad Request</h1>
		<p>Your request honestly kinda sucked.</p>
	</body>
</html>`
			if err := w.WriteStatusLine(response.BAD_REQUEST); err != nil {
				log.Printf("error escribiendo el status line: %v", err)
			}
			hs := response.GetDefaultHeaders(len(msg), false)
			hs.Replace("content-type", "text/html")
			w.WriteHeaders(hs)
			w.WriteBody([]byte(msg))

		case "/myproblem":
			respond500(w)
			return
		case "/video":
			file, err := os.Open("assets/vim.mp4")
			if err != nil {
				respond500(w)
				return
			}

			if err := w.WriteStatusLine(response.OK); err != nil {
				hE := &server.HandlerError{
					StatusCode: response.INTERNAL_SERVER_ERROR,
					Msg:        "error respondiendo",
				}
				log.Printf("error escribiendo status line")
				hE.Write(w)
				return
			}

			hs := response.GetDefaultHeaders(0, true)
			hs.Replace("content-type", "video/mp4")
			hs.Set("trailer", "x-content-sha256, x-content-length")
				if err := w.WriteHeaders(hs); err != nil {
				respond500(w)
				log.Printf("error escribiendo headers")
				return
			}

			hash := sha256.New()
			totalLength := 0
			buffer := make([]byte, 1024)
			for {
				n, err := file.Read(buffer)
				if err != nil {
					if err == io.EOF {
						break
					}
					respond500(w)
					log.Printf("Problema leyendo el video: %v", err)
					return
				}

				chunk := buffer[:n]
				hash.Write(chunk)
				totalLength += n
				w.WriteChunkedBody(chunk)
			}

			hashSum := hash.Sum(nil)
			trailers := headers.NewHeaders()
			trailers.Set("x-content-sha256", fmt.Sprintf("%x", hashSum))
			trailers.Set("x-content-length", fmt.Sprintf("%d", totalLength))
			w.WriteChunkedBodyDone()

		default:
			msg := `<html>
	<head>
		<title>200 OK</title>
	</head>
	<body>
		<h1>Success!</h1>
		<p>Your request was an absolute banger.</p>
	</body>
</html>`
			if err := w.WriteStatusLine(response.OK); err != nil {
				hE := &server.HandlerError{
					StatusCode: response.INTERNAL_SERVER_ERROR,
					Msg:        "error respondiendo",
				}
				log.Printf("error escribiendo status line")
				hE.Write(w)
				return
			}

			hs := response.GetDefaultHeaders(len(msg), false)
			hs.Replace("content-type", "text/html")
			if err := w.WriteHeaders(hs); err != nil {
				hE := &server.HandlerError{
					StatusCode: response.INTERNAL_SERVER_ERROR,
					Msg:        "error respondiendo",
				}
				log.Printf("error escribiendo headers")
				hE.Write(w)
				return
			}

			w.WriteBody([]byte(msg))
		}
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
