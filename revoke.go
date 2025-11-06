package main

import (
	"net/http"

	"github.com/fronigiri/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	tokenRefresh, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unable to retrieve token", err)
		return
	}
	_, err2 := cfg.db.RevokeToken(r.Context(), tokenRefresh)
	if err2 != nil {
		respondWithError(w, http.StatusUnauthorized, "unable to revoke token", err2)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
