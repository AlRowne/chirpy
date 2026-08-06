package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
)

func GetBearerToken(headers http.Header) (string, error) {
	hv := headers.Get("Authorization")
	s := strings.Fields(strings.TrimSpace(hv))
	if len(s) != 2 {
		return "", errors.New("no bearer token found")
	}
	if s[0] != "Bearer" {
		return "", errors.New("no bearer token found")
	}
	token := s[1]
	return token, nil
}

func MakeRefreshToken() string {
	key := make([]byte, 32)
	rand.Read(key)
	hexKey := hex.EncodeToString(key)
	return hexKey
}

func GetAPIKey(headers http.Header) (string, error) {
	hv := headers.Get("Authorization")
	s := strings.Fields(strings.TrimSpace(hv))
	if len(s) != 2 {
		return "", errors.New("no API key found")
	}
	if s[0] != "ApiKey" {
		return "", errors.New("no API key found")
	}
	token := s[1]
	return token, nil
}
