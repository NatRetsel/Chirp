# http_server_golang
Http server in go

Creating a http server from scratch - will be using Golang in this exercise. 


### Http request Multiplexer
We need a way to differentiate and tell which handler gets assigned to specific http requests. A http request multiplexer (mux) is used to route incoming http requests to specific handlers based on the URL path. 

We will then create our handlers, routes and register them to our mux.

`mux := http.NewServerMux()`

### Http server struct
In Golang, we provide the http server struct an address port as well as a handler - our mux. Here's an example of an empty server, with no routes or handlers registered with our mux.

```go
func main(){
	const port = "8080"
	mux := http.NewServerMux()
	srv := &http.Server{
		Addr: ":" + port,
		Handler: mux,
	}
	srv.ListenAndServe()
}
```

#### Adding a handler
Assuming we are building a fileserver and have a homepage titled `index.html` in the root directory, we use the `.Handle()` method of mux to add a handler to the root path (`/`). We will use a standard `http.Fileserver` as the handler.

```go
func main(){
	const port = "8080"
	const filepathroot = "."
	mux := http.NewServerMux()

	// Specify upon landing at root path, http.FileServer will handle
	// serving contents of file system rooted at root.
	// As a special case, returned file server redirects any request
	// ending in "/index.html" to the same path without specifying
	// "index.html"
	mux.Handle("/", http.FileServer(http.Dir(filepathroot)))
	srv := &http.Server{
		Addr: ":" + port,
		Handler: mux,
	}
	srv.ListenAndServe()
}
```
