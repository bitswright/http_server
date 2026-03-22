package httpcore

import (
	"strings"
)

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
func (r *Router) Handle(method, path string, fn HandlerFunc) {
	key := strings.ToUpper(method) + " " + path
	r.routes[key] = fn
}

// Dispatch() builds the same key as in the incoming request and looks up the handler function in the router
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
			return Response{Status: StatusMethodNotAllowed, Body: "Method not allowed"}
		}
	}
	// For learning purpose, this will be fine -> O(n) lookup time
	// But in production, we should use a better data structure like a trie -> O(logn) or better lookup time
	return Response{Status: StatusNotFound, Body: "Not Found"}
}
