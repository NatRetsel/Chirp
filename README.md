# Twitter clone - Chirp
Guided project in Golang from boot.dev


## Contents
1. [How to use](#1-how-to-use)
2. [Route and handlers](#2-route-and-handlers)
	- [Http request multiplexer](#http-request-multiplexer)
	- [Http server struct](#http-server-struct)
	- [Adding handlers](#adding-a-handler)
	- [Custom handlers](#custom-handler)
	- [Custom handler function wrapper](#custom-handler-function-wrapper)

## 1. How to use
WIP

## 2. Route and Handlers
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
#### Custom handler
The `http.Handler` is an interface:
```go
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}
```
Therefore, any type that implements the `ServeHTTP` method or a function that matches that signature is an `http.Handler`.

Let's implement a handler for the endpoint `/healthz` through a function `handlerReadiness`. We will also use the `/app/` path instead of `/` to serve our `index.html`.

In the `http.ResponseWriter` interface, we will specify the header key-value pair, the http status and the body.

```go
func handlerReadiness(w http.ResponseWriter, req *http.Request) {

	if req.URL.Path != "/healthz" {
		http.NotFound(w, req)
		return
	}
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

  

func main() {

	const port = "8080"
	const filepathroot = "."
	
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir(filepathroot))))
	
	mux.HandleFunc("/healthz", handlerReadiness)
	srv := &http.Server{
		Addr: ":" + port,
		Handler: mux,
	}
	
	log.Printf("Serving on port: %s\n", port)
	
	srv.ListenAndServe()
}
```

#### Custom handler function wrapper
Supposed we want to keep track of of the number of hits for each landing on our `/app` endpoint:
* We need a way to increment such a variable in an asynchronous manner safely.
* We need a way to call the increment method each time the handler for the specific route is called.

##### Safe increment
Besides `Mutex` lock and unlock, there are `atomic` types that support such operations safely.

```go
type apiConfig struct {
	fileserverHits atomic.Int32
}

func main(){
	apiCfg = &apiConfig{
		fileserverHits: atomic.Int32{},
	}
	apiCfg.Add(1)
}
```

##### Wrapping safe increment in HandlerFunc inside middleware
We can embed the increment method inside a wrapper that our `apiConfig` implements - this wrapper wraps the specific handler which we are interested in tracking. Since we know `http.Handler` is an interface, we can define a function of the same signature and return it. Inside this function, we will make use of the wrapped handler's `serveHTTP` method. We then wrap this function with `http.HandlerFunc`.

Note: we could use `http.HandleFunc` but we will need to pass in the route. In the spirit of DRY code, we can simply wrap it with `http.HandlerFunc`.

```go
type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc (next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	cfg.fileserverHits.Add(1)
	next.ServeHTTP(w, r)

	})
}

func main(){
	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
	}
	
	mux := http.NewServeMux()
	
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathroot)))))
}
```

### Routing and restricting methods on endpoints
In the standard library, we can specify method access on the endpoint in the format `[METHOD] [HOST]/[PATH]`. When a request is made to the endpoints with an unspecified method, the server handles it with a `405` response code (method not allowed).

```go

func main(){

	const port = "8080"
	const filepathroot = "." 
	
	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
	}
	
	mux := http.NewServeMux()
	
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathroot)))))
	
	// restrict access only for specific methods
	mux.HandleFunc("GET /healthz", handlerReadiness)
	mux.HandleFunc("GET /metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /reset", apiCfg.handlerReset)
}
```

