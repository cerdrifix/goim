package main

import (
	"./pages"
	"./server"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
)

var (
	ServerAddress      = os.Getenv("SERVER_ADDRESS")
	CertificateFile    = os.Getenv("HTTPS_CERTIFICATE_FILE")
	CertificateKeyFile = os.Getenv("HTTPS_CERTIFICATE_KEY")
	DbHostname         = os.Getenv("DB_SERVER")
	DbPort             = os.Getenv("DB_PORT")
	DbName             = os.Getenv("DB_NAME")
	DbUsername         = os.Getenv("DB_USERNAME")
	DbPassword         = os.Getenv("DB_PASSWORD")
)

func main() {

	logger := log.New(os.Stdout, "\nproc-engine | ", log.Lshortfile|log.LstdFlags)

	db, err := sqlx.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", DbUsername, DbPassword, DbHostname, DbPort, DbName))
	if err != nil {
		logger.Fatalf("Error connecting to database: %v", err)
	}

	// Executing a ping to check if we can connect
	err = db.Ping()
	if err != nil {
		logger.Fatalf("Error. Can't connect to database: %v", err)
	}

	h := pages.New(logger, db)

	mux := http.NewServeMux()
	setupRoutes(mux, h)

	logger.Printf("Creating server on address %v", ServerAddress)
	srv := server.New(mux, ServerAddress)

	err = srv.ListenAndServeTLS(CertificateFile, CertificateKeyFile)
	if err != nil {
		logger.Fatalf("Error during server startup: %v", err)
	}
}

func setupRoutes(mux *http.ServeMux, h *pages.Handler) {
	mux.HandleFunc("/", h.Logger(h.Home))
	mux.HandleFunc("/createProcess", h.Logger(h.CreateProcess))
}
