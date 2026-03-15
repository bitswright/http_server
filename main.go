package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	// 1. Open a TCP socket

	// 1.1. Asking OS to create a TCP socket and bind it to port 8081
	listener, err := net.Listen("tcp", ":8081")
	// net.Listen() does 3 things:
	//     1. socket() - creates a socket file descriptor
	//     2. bind() - binds the socket to the port
	//     3. listen() - tells kernel to start queuing up incoming connections
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	fmt.Println("Server is running on port 8081")
	fmt.Println("Waiting for incoming connections...")

	// 2. Block and wait for incoming connections
	for {
		conn, err := listener.Accept()
		// Accept() is a blocking call that waits for an incoming connection
		// TCP 3-way handshake is handled in Accept (SYN -> SYN-ACK -> ACK) before returning
		// hence, conn is a live TCP connection
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// conn is now a live TCP connection
		// 3. Handling the request
		// giving away request processing to a separate function
		// if this function is not made to be a goroutine, then the server will be blocked and unable to accept new connections
		// until current request/connection is processed
		fmt.Println("\n\nNew connection accepted")
		fmt.Println("Handling request...")
		handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	// 6. Close the connection after the request is processed
	// This is important to avoid resource leaks and allow the server to handle multiple connections concurrently
	defer conn.Close()

	// 4. Wrap connection in a bufio.Reader and read the request line by line until empty line is encountered
	reader := bufio.NewReader(conn)
	// using bufio.Reader because TCP is a stream (any number of bytes can be received in a stream)
	// we need to accumulate these bytes internally and give clean line-by-line interface to the user
	// It is one of the problems every HTTP server has to solve
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		if line == "\r\n" || line == "\n" {
			// "\r\n" is the official HTTP request header terminator, everything after this is the HTTP request body
			break
		}
		fmt.Print(line)
	}

	// 5. Write the response back to the client
	response := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/html\r\n" +
		"Content-Length: 13\r\n" +
		"\r\n" +
		"Hello, World!"

	conn.Write([]byte(response))
}
