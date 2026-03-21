package handlers

import "github.com/bitswright/http_server/httpcore"

// each handler is a function that takes a *Request, do it's logic and returns a Response
// no frameworkmagic just a function

func Home(request *httpcore.Request) httpcore.Response {
	return httpcore.Response{Status: 200, Body: "Welcome!"}
}
