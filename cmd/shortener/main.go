package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/spider4216/tinyurl/internal/config"
	"github.com/spider4216/tinyurl/internal/handler"
	"github.com/spider4216/tinyurl/internal/repository"
	"github.com/spider4216/tinyurl/internal/service"
)

func main() {
	flags := InitFlags()

	conf := config.New(flags.domain, flags.srvHost)
	store := map[string]string{}
	repo := repository.New(store)
	service := service.New(repo)
	handler := handler.New(conf, service)

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Post("/", http.HandlerFunc(handler.GenerateId))
		r.Get("/{id}", http.HandlerFunc(handler.GetUrl))
	})

	err := http.ListenAndServe(flags.srvHost, r)

	if err != nil {
		panic(err.Error())
	}
}
