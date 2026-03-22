package httpcore

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

// Defining a struct to hold the response information
type Response struct {
	Version string            // "HTTP/1.1", "HTTP/2.0", etc.
	Status  int               // eg. 200, 404, 500, etc.
	Headers map[string]string // eg. {"Content-Type": "text/html", "Content-Length": "13", etc.}
	Body    string            // response body (HTML, JSON, etc.)
}

func writeResponse(conn net.Conn, response *Response) {
	statusText := getStatusText(response.Status)
	contentLength := strconv.Itoa(len(response.Body))
	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}
	response.Headers["content-length"] = contentLength
	version := response.Version
	if version == "" {
		version = "HTTP/1.1"
	}
	var rawResponseStringBuilder strings.Builder
	rawResponseStringBuilder.WriteString(fmt.Sprintf("%s %d %s\r\n", version, response.Status, statusText))
	for key, value := range response.Headers {
		rawResponseStringBuilder.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	rawResponseStringBuilder.WriteString("\r\n")
	rawResponseStringBuilder.WriteString(response.Body)

	if _, err := conn.Write([]byte(rawResponseStringBuilder.String())); err != nil {
		log.Printf("failed to write response to %s: %v", conn.RemoteAddr(), err)
	}
}

func fileNotFoundResponse() Response {
	return Response{Status: StatusNotFound, Body: "File not found"}
}

func internalServerErrorResponse() Response {
	return Response{Status: StatusInternalServerError, Body: "Internal server error"}
}
