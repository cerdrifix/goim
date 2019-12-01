package pages

import (
	"../engine"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
	"net/http"
	"time"
)

const message = "Welcome to GoFlow.\n"

type Handler struct {
	logger *log.Logger
	db     *sqlx.DB
}

func New(logger *log.Logger, db *sqlx.DB) *Handler {
	return &Handler{
		logger: logger,
		db:     db,
	}
}

func (h *Handler) Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		next(w, r)
		defer h.logger.Printf("Request processed in %v", time.Now().Sub(startTime).Milliseconds())
	}
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(message + " " + time.Now().Format("2006-01-02 15:04:05.999999999")))
	if err != nil {
		h.logger.Fatalf("Error writing the response: %v", err)
	}
}

func (h *Handler) CreateProcess(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var payload engine.CreateProcessPayload
	err := decoder.Decode(&payload)
	if err != nil {
		h.logger.Panicf("Error during payload parse: %v", err)
	}

	eng := engine.New(h.logger, h.db, r.Context())

	i, errs := eng.NewInstance(payload)
	if errs != nil {
		h.logger.Printf("%v", err)
		w.WriteHeader(500)
		for _, _err := range errs {
			_, _ = w.Write([]byte(fmt.Sprintf("\n %s", _err.Error())))
		}
		return
	}

	h.logger.Printf("\nPayload: %#v\nInstance#: %d", payload, i)
	w.WriteHeader(200)
	_, _ = w.Write([]byte(fmt.Sprintf("Processo creato correttamente. ID: %s", i)))
}
