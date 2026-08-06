package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/AlRowne/chirpy/internal/database"
	"github.com/AlRowne/chirpy/internal/database/auth"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	req := request{}
	err := decoder.Decode(&req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't decode parameters", err)
		return
	}

	user, err := cfg.dbQueries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}
	match, err := auth.CheckPasswordHash(req.Password, user.HashedPassword)
	if err != nil || !match {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}
	token, err := auth.MakeJWT(user.ID, cfg.Secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't make JWT", err)
		return
	}
	refreshParams := database.CreateRefreshTokenParams{
		Token:  auth.MakeRefreshToken(),
		UserID: user.ID,
	}
	if err := cfg.dbQueries.CreateRefreshToken(r.Context(), refreshParams); err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't create refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, loginResponse{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		IsChirpyRed:  user.IsChirpyRed,
		Token:        token,
		RefreshToken: refreshParams.Token,
	})
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "couldn't get token", err)
		return
	}
	dbUser, err := cfg.dbQueries.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized", err)
		return
	}
	token, err := auth.MakeJWT(dbUser.ID, cfg.Secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't make JWT", err)
		return
	}
	respondWithJSON(w, http.StatusOK, refreshResponse{
		Token: token,
	})
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "couldn't get token", err)
		return
	}
	err = cfg.dbQueries.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't revoke token", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
