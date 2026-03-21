package httpcore

import (
	"fmt"
	"mime"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func HandleConnection(conn net.Conn, router *Router) {
	// 6. Close the connection after the request is processed
	// This is important to avoid resource leaks and allow the server to handle multiple connections concurrently
	defer conn.Close()

	request, err := parseRequest(conn)
	if err != nil {
		writeResponse(conn, &Response{Status: 400, Body: "Bad Request: " + err.Error()})
		return
	}

	fmt.Printf("--> %s %s\n", request.Method, request.Path)
	var response Response
	if strings.HasPrefix(request.Path, "/static/") {
		// pragmatic shortcut, production routers handles prefix matching inside router itself using trie
		response = staticHandler(request)
	} else {
		response = router.Dispatch(request)
	}
	fmt.Printf("<-- %d %s\n", response.Status, response.Body)

	writeResponse(conn, &response)
}

const staticDir = "./static"

func staticHandler(request *Request) Response {
	// 1. Stripping /static/ (if exists) from the path to get the relative file path
	relativePath := strings.TrimPrefix(request.Path, "/static/")
	if relativePath == "" {
		relativePath = "index.html"
	}

	// 2. Join it with static root directory
	fullPath := filepath.Join(staticDir, relativePath)
	// filepath.Join() is aware of the OS and will handle the path separator correctly

	// 3. Cleaning the path
	// to prevent directory traversal attacks, eg. /static/../../etc/password
	fullPath = filepath.Clean(fullPath)
	// checking if the given path is belonging to the static folder or not
	// if not we return 404 (Not Found) error not 403 (Access is denied) to not let attacker know file exist or not
	absStaticDirPath, _ := filepath.Abs(staticDir)
	absRequestedFilePath, _ := filepath.Abs(fullPath)
	if !strings.HasPrefix(absRequestedFilePath, absStaticDirPath) {
		return fileNotFoundResponse()
	}

	// 4. Reading the file
	// os.ReadFile() reads the complete file into byte slice in one call
	// 		Fine for now, but for large file, need to use os.Open to stream the response
	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fileNotFoundResponse()
		}
		// I/O error, Permission denied, etc
		// Never return real error message to the user (could expose file system, system info, strack trace, etc)
		// Could log the error internal but return a standard response
		return internalServerErrorResponse()
	}

	// 5. Detecting Content-Type from extension
	// Browser uses Content-Type header not file extension - to decide way to render a response
	// we can use mime.TypeByExtension as it maps extension to MIME using OS's MIME registry
	fileExtension := filepath.Ext(fullPath)
	contentType := mime.TypeByExtension(fileExtension)
	if contentType == "" {
		// unknown type -> download
		contentType = "application/octet-stream"
	}

	return Response{
		Status:  200,
		Headers: map[string]string{"Content-Type": contentType},
		Body:    string(data),
	}
}
