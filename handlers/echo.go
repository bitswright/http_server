package handlers

import "github.com/bitswright/http_server/httpcore"

func Echo(request *httpcore.Request) httpcore.Response {
	if request.Body == "" {
		return httpcore.Response{Status: 400, Body: "Send a body to echo"}
	}
	return httpcore.Response{Status: 200, Body: request.Body}
}
