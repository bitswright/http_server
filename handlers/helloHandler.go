package handlers

import "github.com/bitswright/http_server/httpcore"

func Hello(request *httpcore.Request) httpcore.Response {
	name := "world"
	if nameValue, ok := request.QueryParams["name"]; ok {
		name = nameValue
	}
	message := "Hello, " + name + "!"
	return httpcore.Response{Status: 200, Body: message}
}
