package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"
)

// Defining a struct to hold the request information
type Request struct {
	Method      string            // "GET", "POST", etc.
	Path        string            // eg. "/hello", "/", etc.
	QueryParams map[string]string // eg. {"name": "john", "age": "20"}, etc.
	Version     string            // "HTTP/1.1", "HTTP/2.0", etc.
	Headers     map[string]string // eg. {"Host": "localhost:8081", "User-Agent": "Mozilla/5.0", etc.}
	Body        string            // request body (POST/PUT)
}

// this struct is a contract between the parser and the handler
// everything downstream (routing, middleware, business logic) will receive a `*Request`
// designing it well now will save us a lot of time later (refactoring, debugging, maintainability, etc)

// Defining a struct to hold the response information
type Response struct {
	Version string            // "HTTP/1.1", "HTTP/2.0", etc.
	Status  int               // eg. 200, 404, 500, etc.
	Headers map[string]string // eg. {"Content-Type": "text/html", "Content-Length": "13", etc.}
	Body    string            // response body (HTML, JSON, etc.)
}

type HandlerFunc func(request *Request) Response

// HandlerFunc is a Go function type with fixed signature (takes a *Request, returns a Response)
// this contract is to be satisfied by all handlers in the server
// it is a contract between the router and the handler
// the router will call the handler with the request and the handler will return the response
// the router will then write the response to the client
// the handler will be responsible for handling the request and returning the response

type Router struct {
	routes map[string]HandlerFunc
	// in routes key will be "METHOD /path"
	// O(1) lookup time for routing irrespective of the number of routes
}

func NewRouter() *Router {
	return &Router{routes: make(map[string]HandlerFunc)}
}

// NewRouter() is a constructor function that returns a new Router instance
// it is important because if routes are not initialized, then a nil map will panic on write
// hence, NewRouter() gurantees routes map is always ready to use

func (r *Router) Handle(method, path string, handlerFunction HandlerFunc) {
	key := strings.ToUpper(method) + " " + path
	r.routes[key] = handlerFunction
}

func (r *Router) Dispatch(request *Request) Response {
	key := strings.ToUpper(request.Method) + " " + request.Path
	handler, ok := r.routes[key]

	if ok {
		return handler(request)
	}

	// Path exists but different method -> 405 Method Not Allowed
	for route := range r.routes {
		parts := strings.SplitN(route, " ", 2)
		if parts[1] == request.Path {
			return Response{Status: 405, Body: "Method not allowed"}
		}
	}
	// For learning purpose, this will be fine -> O(n) lookup time
	// But in production, we should use a better data structure like a trie -> O(logn) or better lookup time
	return Response{Status: 404, Body: "Not Found"}
}

// Dispatch() builds the same key as in the incoming request and looks up the handler function in the router

func main() {
	router := NewRouter()
	router.Handle("GET", "/", homeHandler)
	router.Handle("GET", "/hello", helloHandler)
	router.Handle("POST", "/echo", echoHandler)

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
		handleConnection(conn, router)
	}
}

func handleConnection(conn net.Conn, router *Router) {
	// 6. Close the connection after the request is processed
	// This is important to avoid resource leaks and allow the server to handle multiple connections concurrently
	defer conn.Close()

	request, err := parseRequest(conn)
	if err != nil {
		writeResponse(conn, &Response{Status: 400, Body: "Bad Request: " + err.Error()})
		return
	}

	fmt.Printf("--> %s %s\n", request.Method, request.Path)
	response := router.Dispatch(request)
	fmt.Printf("<-- %d %s\n", response.Status, response.Body)
	writeResponse(conn, &response)
}

func parseRequest(conn net.Conn) (*Request, error) {
	// 4. Wrap connection in a bufio.Reader and read the request line by line until empty line is encountered
	reader := bufio.NewReader(conn)
	// using bufio.Reader because TCP is a stream (any number of bytes can be received in a stream)
	// we need to accumulate these bytes internally and give clean line-by-line interface to the user
	// It is one of the problems every HTTP server has to solve

	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	requestLineParts := strings.Fields(requestLine)
	if len(requestLineParts) != 3 {
		return nil, errors.New("invalid request line")
	}
	rawPath := requestLineParts[1]
	path, rawQuery, _ := strings.Cut(rawPath, "?")
	queryParams := make(map[string]string)
	if rawQuery != "" {
		for _, param := range strings.Split(rawQuery, "&") {
			key, value, ok := strings.Cut(param, "=")
			if ok {
				queryParams[key] = value
			}
		}
	}
	// strings.Fields splits on any whitespace (spaces, tabs) and ignores leading/trailing whitespace.
	// It's more robust than strings.Split(line, " ") which breaks on double spaces or tabs.

	headers := make(map[string]string)
	for {
		headerline, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		headerline = strings.TrimSpace(headerline)
		if headerline == "" {
			// empty line indicates that all headers have been read
			break
		}

		key, value, ok := strings.Cut(headerline, ": ")
		// we're using strings.Cut because value may contain ": " as well
		// so we need to cut the string at the first occurrence of ": "
		// and the first part will be the key and the second part will be the value
		// if the string does not contain ": ", then ok will be false
		// and we will return an error
		if !ok {
			continue
		}
		// headers are case-insensitive,
		// hence we convert the key to lowercase to make it consistent
		headers[strings.ToLower(key)] = value
		// headers carry metadata the server needs:
		// what content type the client accepts, how long the body is, whether to keep the connection alive.
	}

	// this is important because:
	//     POST /login with a JSON body, form submissions, file uploads —
	//     all of these rely on Content-Length to tell the server where the body ends.
	body := ""
	if contentLength, ok := headers["content-length"]; ok {
		length := 0
		_, err := fmt.Sscanf(contentLength, "%d", &length)
		// fmt.Sscanf is used to parse the content length string into an integer
		// it returns the number of items successfully parsed and assigned
		// and an error if the string is not a valid integer
		// we don't need the number of items, so we ignore it
		// and we check if the error is nil
		if err != nil {
			return nil, err
		}

		if length > 0 {
			bodyBytes := make([]byte, length)
			reader.Read(bodyBytes)
			// reader.Read() shouldn't be used without knowing how many bytes to expect.
			// On a GET request with no body, Read() would hang waiting for data that isn't coming.
			body = string(bodyBytes)
		}
	}
	// concept of content length is important because it allows the server to know how much data to expect in the body
	// as tcp is a stream, the server does not know when the body ends
	// and it allows the server to read the body in chunks, which is more efficient than reading the entire body at once
	// and it allows the server to handle requests with large bodies, which is important for streaming media, file uploads, etc.

	return &Request{
		Method:      requestLineParts[0],
		Path:        path,
		QueryParams: queryParams,
		Version:     requestLineParts[2],
		Headers:     headers,
		Body:        body,
	}, nil
}

func writeResponse(conn net.Conn, response *Response) {
	statusText := getStatusText(response.Status)
	contentLength := fmt.Sprintf("%d", len(response.Body))
	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}
	response.Headers["content-length"] = contentLength
	if response.Version == "" {
		response.Version = "HTTP/1.1"
	}
	rawResponse := fmt.Sprintf("%s %d %s\r\n", response.Version, response.Status, statusText)
	for key, value := range response.Headers {
		rawResponse += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	rawResponse += "\r\n" + response.Body

	conn.Write([]byte(rawResponse))
}

func getStatusText(status int) string {
	statusTextMap := map[int]string{
		200: "OK",
		400: "Bad Request",
		404: "Not Found",
		405: "Method Not Allowed",
		500: "Internal Server Error",
	}
	if text, ok := statusTextMap[status]; ok {
		return text
	}
	return "undefined"
}

// each handler is a function that takes a *Request, do it's logic and returns a Response
// no frameworkmagic just a function

func homeHandler(request *Request) Response {
	return Response{Status: 200, Body: "Welcome!"}
}

func helloHandler(request *Request) Response {
	name := "world"
	if nameValue, ok := request.QueryParams["name"]; ok {
		name = nameValue
	}
	message := "Hello, " + name + "!"
	return Response{Status: 200, Body: message}
}

func echoHandler(request *Request) Response {
	if request.Body == "" {
		return Response{Status: 400, Body: "Send a body to echo"}
	}
	return Response{Status: 200, Body: request.Body}
}
