package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/spider4216/tinyurl/internal/handler"
	"github.com/spider4216/tinyurl/internal/repository"
	"github.com/spider4216/tinyurl/internal/service"
)

func main() {
	store := map[string]string{}
	repo := repository.New(store)
	service := service.New(repo)
	handler := handler.New(service)

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Post("/", http.HandlerFunc(handler.GenerateId))
		r.Get("/{id}", http.HandlerFunc(handler.GetUrl))
	})

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	err := srv.ListenAndServe()

	if err != nil {
		fmt.Println("Error", err)
		return
	}
}
