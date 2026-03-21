package httpcore

import (
	"fmt"
	"net"
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

// TODO: To be moved to another file later
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

func fileNotFoundResponse() Response {
	return Response{Status: 404, Body: "File not found"}
}

func internalServerErrorResponse() Response {
	return Response{Status: 500, Body: "Internal server error"}
}
