package main

import (
	"UrlShortner/controller"
	"UrlShortner/database"
	"UrlShortner/repository"
	"UrlShortner/service"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	db := database.InitDB()

	repo := &repository.ShortUrlRepository{DB: db}
	urlService := &service.UrlService{Repo: repo}
	urlController := &controller.UrlController{Service: urlService}

	r := mux.NewRouter()
	r.HandleFunc("/shorten", urlController.ShortenUrl).Methods("POST")
	r.HandleFunc("/r/{shortCode}", urlController.Redirect).Methods("GET")

	http.ListenAndServe(":8080", r)
}
