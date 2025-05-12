package main

import (
	"bufio"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	fmt.Println("listening :8080")
	shutdown := make(chan struct{})
	var wg sync.WaitGroup
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-shutdown:
					fmt.Println("new connection won't be accepted")
					return
				default:
					fmt.Printf("failed to accept connection: %v", err)
					continue
				}
			}
			fmt.Printf("connection from: %v\n", conn.RemoteAddr())
			wg.Add(1)
			go handle(conn, &wg)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	fmt.Println("shutting down")
	close(shutdown)
	wg.Wait()
	fmt.Println("all handlers finished, shutdown complete")
}

func handle(conn net.Conn, wg *sync.WaitGroup) {
	reader := bufio.NewReader(conn)
	defer func() {
		conn.Close()
		wg.Done()
	}()
	for {
		var lines []string
	read:
		for {
			conn.SetReadDeadline(time.Now().Add(10 * time.Second))
			l, err := reader.ReadString('\n') // blocks, EOF is triggered when client closes connection
			if err != nil {
				if errors.Is(err, io.EOF) {
					fmt.Println("client closed connection")
					return
				}
				fmt.Printf("read error: %v\n", err)
				return
			}
			trimmed := strings.TrimRight(l, "\r\n")
			if trimmed == "" {
				break read
			}
			lines = append(lines, trimmed)
		}
		if len(lines) == 0 {
			fmt.Println("empty message")
			return
		}
		reqLine := strings.Split(lines[0], " ")
		if len(reqLine) != 3 {
			fmt.Println("malformed request line")
			return
		}
		fmt.Printf("request: %s %s %s\n", reqLine[0], reqLine[1], reqLine[2])
		headers := make(map[string]string)
		for _, v := range lines[1:] {
			colonIdx := strings.Index(v, ":")
			if colonIdx == -1 {
				continue
			}
			key := strings.ToLower(v[:colonIdx])
			val := strings.TrimSpace(v[colonIdx+1:])
			headers[key] = val
		}
		var body string
		switch reqLine[1] {
		case "/ping":
			body = "pong\n"
		default:
			body = "Learning HTTP/1.1 🧑‍💻\n"
		}
		md5Sum := md5.Sum([]byte(body))
		var resp strings.Builder
		resp.WriteString("HTTP/1.1 200 OK\r\n")
		resp.WriteString(fmt.Sprintf("Content-Length: %d\r\n", len(body)))
		resp.WriteString("Content-Type: text/plain\r\n")
		resp.WriteString(fmt.Sprintf("Content-MD5: %s\r\n", base64.StdEncoding.EncodeToString(md5Sum[:])))
		shouldClose := false
		if val, ok := headers["connection"]; ok && val == "close" {
			shouldClose = true
			resp.WriteString("Connection: close\r\n")
		}
		resp.WriteString("\r\n")
		resp.WriteString(body)
		conn.Write([]byte(resp.String()))
		if shouldClose {
			fmt.Println("closing connection")
			return
		}
	}
}
