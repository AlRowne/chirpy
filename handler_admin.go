package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	if cfg.Platform != "dev" {
		respondWithError(w, http.StatusForbidden, "forbidden", nil)
		return
	}
	err := cfg.dbQueries.DeleteUsers(r.Context())
	fmt.Println("All users have been deleted")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't delete users", err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
