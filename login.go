package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/fronigiri/chirpy/internal/auth"
)

const (
	defaultExpSec = 3600
	maxExpSec     = 3600
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password         string `json:"password"`
		Email            string `json:"email"`
		ExpiresInSeconds *int   `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	exp := defaultExpSec
	if params.ExpiresInSeconds != nil {
		v := *params.ExpiresInSeconds
		if v <= 0 {
			exp = defaultExpSec
		} else if v > maxExpSec {
			exp = maxExpSec
		} else {
			exp = v
		}
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	user, err := cfg.db.GetUserEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "incorrect email or password", err)
	}

	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil || !match {
		respondWithError(w, http.StatusUnauthorized, "incorrect email or password", err)
	}
	token, err := auth.MakeJWT(user.ID, cfg.JWTSecret, time.Duration(exp)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unable to create token", err)
		return
	}

	u := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     token,
	}
	respondWithJSON(w, http.StatusOK, u)

}
