package main

import (
	"net/http"
	"time"

	"github.com/fronigiri/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	tokenRefresh, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unable to retrieve token", err)
		return
	}
	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), tokenRefresh)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unable to find user", err)
		return
	}
	token, err := auth.MakeJWT(user.ID, cfg.JWTSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable make token", err)
		return
	}
	response := resp{Token: token}

	respondWithJSON(w, http.StatusOK, response)
}

type resp struct {
	Token string `json:"token"`
}
