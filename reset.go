package main

import (
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	godotenv.Load()
	if os.Getenv("PLATFORM") != "dev" {
		respondWithError(w, http.StatusForbidden, "unable to delete all users", nil)
		return
	}
	err := cfg.db.DeleteAllUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to delete all users", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}
