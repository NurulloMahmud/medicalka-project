package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Envelope map[string]interface{}

func WriteJSON(w http.ResponseWriter, status int, data Envelope) error {
	js, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}

	js = append(js, '\n')
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func InternalServerError(w http.ResponseWriter, r *http.Request, err error, log *log.Logger) {
	requestURL := r.URL.String()
	requestMethod := r.Method
	log.Printf("[ERROR] SERVER ERROR!\nURL: %s\nMethod: %s\nError: %v\n", requestURL, requestMethod, err.Error())
	WriteJSON(w, http.StatusInternalServerError, Envelope{"error": "internal server error"})
}

func BadRequest(w http.ResponseWriter, r *http.Request, err error, log *log.Logger) {
	requestURL := r.URL.String()
	requestMethod := r.Method
	log.Printf("[ERROR] BAD REQUEST!\nURL: %s\nMethod: %s\nError: %v\n", requestURL, requestMethod, err.Error())
	WriteJSON(w, http.StatusBadRequest, Envelope{"error": err.Error()})
}

func Unauthorized(w http.ResponseWriter, r *http.Request, msg string) {
	requestURL := r.URL.String()
	requestMethod := r.Method
	log.Printf("[ERROR] UNAUTHORIZED REQUEST!\nURL: %s\nMethod: %s\nError: %v\n", requestURL, requestMethod, msg)
	WriteJSON(w, http.StatusUnauthorized, Envelope{"error": msg})
}

func Forbidden(w http.ResponseWriter, r *http.Request, msg string) {
	requestURL := r.URL.String()
	requestMethod := r.Method
	log.Printf("[ERROR] FORBIDDEN REQUEST!\nURL: %s\nMethod: %s\nError: %v\n", requestURL, requestMethod, msg)
	WriteJSON(w, http.StatusForbidden, Envelope{"error": msg})
}

func RateLimitExceeded(w http.ResponseWriter, r *http.Request) {
	requestURL := r.URL.String()
	requestMethod := r.Method
	msg := fmt.Sprintf("Too many request by IP: %s", r.RemoteAddr)
	log.Printf("[ERROR] RATE LIMIT EXCEEDED!\nURL: %s\nMethod: %s\nError: %v\n", requestURL, requestMethod, msg)
	WriteJSON(w, http.StatusTooManyRequests, Envelope{"error": "Too many requests, please try again later"})
}

func NotFound(w http.ResponseWriter, r *http.Request, log *log.Logger) {
	requestURL := r.URL.String()
	requestMethod := r.Method
	log.Printf("[ERROR] NOT FOUND!\nURL: %s\nMethod: %s\n", requestURL, requestMethod)
	WriteJSON(w, http.StatusNotFound, Envelope{"error": "resource not found"})
}
