package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type requestWebhook struct {
	Event string `json:"event"`
	Data  struct {
		UserID string `json:"user_id"`
	} `json:"data"`
}

func (cfg *apiConfig) handlerWebhooks(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	req := requestWebhook{}
	err := decoder.Decode(&req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't decode json", err)
		return
	}
	if req.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	userID, err := uuid.Parse(req.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "couldn't parse userid", err)
		return
	}
	_, err = cfg.dbQueries.UpgradeUser(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "user not found", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
