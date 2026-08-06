package main

import (
	"sync/atomic"

	"github.com/AlRowne/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	Platform       string
	Secret         string
	PolkaKey       string
}
