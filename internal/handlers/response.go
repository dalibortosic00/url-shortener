package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondWithJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

func respondWithError(w http.ResponseWriter, statusCode int, msg string) {
	if statusCode >= 500 {
		log.Printf("Internal Server Error: %s", msg)
	}

	respondWithJSON(w, statusCode, map[string]string{"error": msg})
}
