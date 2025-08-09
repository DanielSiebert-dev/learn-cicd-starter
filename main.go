// This file is the entry point for the Notely web application, written in Go. The app serves a simple web page and API for notes, optionally connecting to a database for data storage. It handles environment settings, web requests, and security basics. The overall flow is:
 // 1. Load settings from .env.
 // 2. Set up the web server and routes.
 // 3. Connect to a database if configured.
 // 4. Start listening for incoming requests on a port.
 // If no database is set, it runs in a limited mode without data operations.

package main

import (
	"database/sql"
	"embed"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/DanielSiebert-dev/learn-cicd-starter/internal/database"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// Configuration structure to hold app-wide settings, like the database connection.
type apiConfig struct {
	DB *database.Queries
}

// Embed static files (e.g., HTML) into the binary so the app can serve them without external files.
 //go:embed static/*
var staticFiles embed.FS

func main() {
	// Load environment variables from .env file for configuration (port, DB URL, etc.). If missing, use defaults and log a warning.
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("warning: assuming default configuration. .env unreadable: %v", err)
	}

	// Get the port from environment; fatal if not set.
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}

	apiCfg := apiConfig{}

	// Attempt to connect to the database using the URL from environment. If missing, run without DB features and log.
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("DATABASE_URL environment variable is not set")
		log.Println("Running without CRUD endpoints")
	} else {
		db, err := sql.Open("libsql", dbURL)
		if err != nil {
			log.Fatal(err)
		}
		dbQueries := database.New(db)
		apiCfg.DB = dbQueries
		log.Println("Connected to database!")
	}

	// Set up the main router for handling web requests, with CORS for cross-origin security.
	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Route for the root path: Serve the embedded index.html as the main page.
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		f, err := staticFiles.Open("static/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()
		if _, err := io.Copy(w, f); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Set up API routes under /v1, only if DB is connected (for data operations).
	v1Router := chi.NewRouter()
	if apiCfg.DB != nil {
		v1Router.Post("/users", apiCfg.handlerUsersCreate)
		v1Router.Get("/users", apiCfg.middlewareAuth(apiCfg.handlerUsersGet))
		v1Router.Get("/notes", apiCfg.middlewareAuth(apiCfg.handlerNotesGet))
		v1Router.Post("/notes", apiCfg.middlewareAuth(apiCfg.handlerNotesCreate))
	}
	v1Router.Get("/healthz", handlerReadiness)

	router.Mount("/v1", v1Router)

	// Configure and start the HTTP server with timeout for security against attacks.
	srv := &http.Server{
		Addr:                 ":" + port,
		Handler:              router,
		ReadHeaderTimeout:    10 * time.Second,  // Timeout to prevent slow attacks on the server.
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}