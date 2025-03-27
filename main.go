package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/natretsel/chirpy/internal/database"
)

type apiConfig struct {
	//atomic.Int32 allows for safe increment across multiple go routines
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
	secret         string
	polka_key      string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable is not set")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM environment variable is not set")
	}
	secretToken := os.Getenv("SECRET")
	if secretToken == "" {
		log.Fatal("SECRET environment variable is not set")
	}
	polka_key := os.Getenv("POLKA_KEY")
	if polka_key == "" {
		log.Fatal("POLKA_KEY environment variable is not set")
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("error trying to establish connection to db %v :, %v", dbURL, err)
		os.Exit(1)
	}
	dbQueries := database.New(db)

	const port = "8080"
	const filepathroot = "."

	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
		dbQueries:      dbQueries,
		platform:       platform,
		secret:         secretToken,
		polka_key:      polka_key,
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathroot)))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirpsCreate)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerChirpsGet)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerChirpsGetByID)
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)
	mux.HandleFunc("POST /api/refresh", apiCfg.handlerValidateRefresh)
	mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevoke)
	mux.HandleFunc("PUT /api/users", apiCfg.handlerUpdateInfo)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerDeleteChirp)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerUpgradeChirpyRed)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)

	srv.ListenAndServe()
}
