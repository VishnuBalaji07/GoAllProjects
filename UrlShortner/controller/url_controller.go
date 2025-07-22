package controller

import (
	"UrlShortner/service"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type UrlRequest struct {
	URL string `json:"url"`
}

type UrlResponse struct {
	ShortCode string `json:"shortCode"`
}

type UrlController struct {
	Service *service.UrlService
}

func (c *UrlController) ShortenUrl(w http.ResponseWriter, r *http.Request) {
	var req UrlRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.URL == "" {
		http.Error(w, "Invalid request. Expected JSON: {\"url\": \"https://example.com\"}", http.StatusBadRequest)
		return
	}

	shortCode := c.Service.CreateShortUrl(req.URL)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UrlResponse{ShortCode: shortCode})
}

func (c *UrlController) Redirect(w http.ResponseWriter, r *http.Request) {
	shortCode := mux.Vars(r)["shortCode"]

	originalUrl := c.Service.GetOriginalUrl(shortCode)
	if originalUrl != nil {
		http.Redirect(w, r, *originalUrl, http.StatusFound)
	} else {
		http.NotFound(w, r)
	}
}
