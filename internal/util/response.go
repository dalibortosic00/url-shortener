package util

import (
	"encoding/json"
	"log"
	"net/http"
)

func RespondWithJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

func RespondWithError(w http.ResponseWriter, statusCode int, msg string) {
	if statusCode >= 500 {
		log.Printf("Internal Server Error: %s", msg)
	}

	RespondWithJSON(w, statusCode, map[string]string{"error": msg})
}
