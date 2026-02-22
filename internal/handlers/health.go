package handlers

import (
	"net/http"

	"github.com/dalibortosic00/url-shortener/internal/util"
)

func Health(w http.ResponseWriter, r *http.Request) {
	util.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
