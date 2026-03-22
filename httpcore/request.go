package httpcore

import (
	"bufio"
	"errors"
	"io"
	"net"
	"strconv"
	"strings"
)

// Defining a struct to hold the request information
// this struct is a contract between the parser and the handler
// everything downstream (routing, middleware, business logic) will receive a `*Request`
// designing it well now will save us a lot of time later (refactoring, debugging, maintainability, etc)
type Request struct {
	Method      string            // "GET", "POST", etc.
	Path        string            // eg. "/hello", "/", etc.
	QueryParams map[string]string // eg. {"name": "john", "age": "20"}, etc.
	Version     string            // "HTTP/1.1", "HTTP/2.0", etc.
	Headers     map[string]string // eg. {"Host": "localhost:8081", "User-Agent": "Mozilla/5.0", etc.}
	Body        string            // request body (POST/PUT)
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
	if contentLengthString, ok := headers["content-length"]; ok {
		length, err := strconv.Atoi(contentLengthString)
		if err != nil {
			return nil, err
		}

		if length > 0 {
			bodyBytes := make([]byte, length)
			if _, err := io.ReadFull(reader, bodyBytes); err != nil {
				return nil, err
			}
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
