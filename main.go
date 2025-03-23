package main

import (
	"log"
	"net/http"
)

func readinessHandler(w http.ResponseWriter, req *http.Request) {
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
	mux.HandleFunc("/healthz", readinessHandler)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)

	srv.ListenAndServe()
}
