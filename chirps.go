package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/fronigiri/chirpy/internal/auth"
	"github.com/fronigiri/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChirps(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid or missing bearer token", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid token", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := getCleanedBody(params.Body, badWords)
	chirpParams := database.AddChirpParams{
		Body:   cleaned,
		UserID: userID,
	}

	c, err := cfg.db.AddChirp(r.Context(), chirpParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't add chirp to database", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, toChirpResponse(c))
}

func (cfg *apiConfig) handlerAllChirps(w http.ResponseWriter, r *http.Request) {

	s := r.URL.Query().Get("author_id")
	sorting := r.URL.Query().Get("sort")
	if s != "" {
		uuid, err := uuid.Parse(s)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "chirp ID not valid UUID", err)
			return
		}
		chirps, err2 := cfg.db.GetAuthorChirps(r.Context(), uuid)
		if err2 != nil {
			respondWithError(w, http.StatusInternalServerError, "couldn't fetch all chirps for user given", err2)
			return
		}
		if chirps == nil {
			chirps = []database.Chirp{}
		}
		if sorting == "desc" {
			sort.Slice(chirps, func(i, j int) bool { return chirps[i].CreatedAt.After(chirps[j].CreatedAt) })
		}
		res := make([]chirpResponse, len(chirps))
		for i, c := range chirps {
			res[i] = toChirpResponse(c)
		}

		respondWithJSON(w, http.StatusOK, res)
	} else {
		chirps, err := cfg.db.AllChirps(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "couldn't fetch all chirps from database", err)
			return
		}
		if chirps == nil {
			chirps = []database.Chirp{}
		}
		if sorting == "desc" {
			sort.Slice(chirps, func(i, j int) bool { return chirps[i].CreatedAt.After(chirps[j].CreatedAt) })
		}
		res := make([]chirpResponse, len(chirps))
		for i, c := range chirps {
			res[i] = toChirpResponse(c)
		}
		respondWithJSON(w, http.StatusOK, res)
	}
}

func (cfg *apiConfig) handlerChirpID(w http.ResponseWriter, r *http.Request) {
	s := r.PathValue("chirpID")
	uuid, err := uuid.Parse(s)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "chirp ID not valid UUID", err)
		return
	}

	chirp, err := cfg.db.GetChirpID(r.Context(), uuid)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "unable to find chirp with provided ID", err)
		return
	}
	respondWithJSON(w, http.StatusOK, toChirpResponse(chirp))
}

func (cfg *apiConfig) handlerChirpsDelete(w http.ResponseWriter, r *http.Request) {
	s := r.PathValue("chirpID")
	uuid, err := uuid.Parse(s)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "chirp ID not valid UUID", err)
		return
	}

	chirp, err := cfg.db.GetChirpID(r.Context(), uuid)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "unable to find chirp with provided ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unable to retrieve token", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	if chirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "provided user is not the author of this chirp", nil)
		return
	}
	err2 := cfg.db.DeleteChirp(r.Context(), chirp.ID)
	if err2 != nil {
		respondWithError(w, http.StatusUnauthorized, "unable to delete chirp", err2)
	}

	w.WriteHeader(http.StatusNoContent)
}

type chirpResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    string    `json:"user_id"`
}

func toChirpResponse(c database.Chirp) chirpResponse {
	return chirpResponse{
		ID:        c.ID.String(),
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		Body:      c.Body,
		UserID:    c.UserID.String(),
	}
}

func getCleanedBody(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}
	cleaned := strings.Join(words, " ")
	return cleaned
}
