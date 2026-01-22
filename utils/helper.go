package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

func ReadIdentifierParam(r *http.Request) (string, error) {
	idParam := chi.URLParam(r, "identifier")
	if idParam == "" {
		return "", errors.New("invalid id parameter")
	}

	return idParam, nil
}

func ReadString(r *http.Request, key string, defaultValue string) string {
	s := r.URL.Query().Get(key)
	if s == "" {
		return defaultValue
	}

	return s
}

func ReadCSV(r *http.Request, key string, defaultValue []string) []string {
	csv := r.URL.Query().Get(key)
	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

func GenerateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func ReadInt(r *http.Request, key string, defaultValue int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}

	return i
}
