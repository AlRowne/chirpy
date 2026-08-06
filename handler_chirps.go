package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/AlRowne/chirpy/internal/database"
	"github.com/AlRowne/chirpy/internal/database/auth"
	"github.com/google/uuid"
)

type chirpResponse struct {
	Id         uuid.UUID `json:"id"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
	Body       string    `json:"body"`
	User_id    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerCreateChirps(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "no bearer token found", err)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.Secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "couldn't validate token", err)
		return
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't decode parameters", err)
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}
	chirpParams := database.CreateChirpParams{
		Body:   params.Body,
		UserID: userID,
	}
	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), chirpParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't created chirp", err)
		return
	}
	respondWithJSON(w, http.StatusCreated, chirpResponse{
		Id:         chirp.ID,
		Created_at: chirp.CreatedAt,
		Updated_at: chirp.UpdatedAt,
		Body:       chirp.Body,
		User_id:    chirp.UserID,
	})
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.dbQueries.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't fetch chirps", err)
		return
	}
	jsonChirps := []chirpResponse{}
	for _, chirp := range chirps {
		jsonChirp := chirpResponse{
			Id:         chirp.ID,
			Created_at: chirp.CreatedAt,
			Updated_at: chirp.UpdatedAt,
			Body:       chirp.Body,
			User_id:    chirp.UserID,
		}
		jsonChirps = append(jsonChirps, jsonChirp)
	}
	respondWithJSON(w, http.StatusOK, jsonChirps)
}

func (cfg *apiConfig) handlerGetChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpIDString := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't parse uuid", err)
		return
	}
	chirp, err := cfg.dbQueries.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "couldn't fetch chirp", err)
		return
	}
	respondWithJSON(w, http.StatusOK, chirpResponse{
		Id:         chirp.ID,
		Created_at: chirp.CreatedAt,
		Updated_at: chirp.UpdatedAt,
		Body:       chirp.Body,
		User_id:    chirp.UserID,
	})

}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	chirpIDString := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't parse uuid", err)
		return
	}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized", err)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.Secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized", err)
		return
	}
	chirp, err := cfg.dbQueries.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "couldn't fetch chirp", err)
		return
	}
	if chirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "forbidden", nil)
		return
	}
	err = cfg.dbQueries.DeleteChirp(r.Context(), chirp.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't delete chirp", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func badWordReplacement(s string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Fields(s)
	cleanedWords := []string{}

	for _, w := range words {
		if slices.Contains(badWords, strings.ToLower(w)) {
			cleanedWords = append(cleanedWords, "****")
		} else {
			cleanedWords = append(cleanedWords, w)
		}
	}

	return strings.Join(cleanedWords, " ")

}
