package main

// Application is the struct that handles the requests
type Application struct{}

// handleRequest handles the requests and returns the status code and message
func (a *Application) handleRequest(url, method string) (int, string) {
	if url == "/app/status" && method == "GET" {
		return 200, "Ok"
	}

	if url == "/create/user" && method == "POST" {
		return 201, "User Created"
	}
	return 404, "Not Ok"
}
