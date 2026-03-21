package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/bitswright/http_server/handlers"
	"github.com/bitswright/http_server/httpcore"
)

func main() {
	//creating static files
	createStaticPages()

	router := httpcore.NewRouter()
	router.Handle("GET", "/", handlers.Home)
	router.Handle("GET", "/hello", handlers.Hello)
	router.Handle("POST", "/echo", handlers.Echo)

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
		httpcore.HandleConnection(conn, router)
	}
}

const staticDir = "./static"

func createStaticPages() {
	os.MkdirAll(staticDir, 0755)

	indexFilePath := filepath.Join(staticDir, "index.html")
	indexFileContent := []byte("Hello! Served by Go server.")
	os.WriteFile(indexFilePath, indexFileContent, 0644)

	styleFilePath := filepath.Join(staticDir, "style.css")
	styleFileContent := []byte("body { font-family: sans-serif; }")
	os.WriteFile(styleFilePath, styleFileContent, 0644)
}
