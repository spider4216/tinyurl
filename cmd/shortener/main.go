package main

import (
	"net/http"

	"github.com/spider4216/tinyurl/internal/handler"
	"github.com/spider4216/tinyurl/internal/repository"
	"github.com/spider4216/tinyurl/internal/service"
)

func main() {
	store := map[string]string{}
	repo := repository.New(store)
	service := service.New(repo)
	handler := handler.New(service)

	mux := http.NewServeMux()

	mux.Handle("/", http.HandlerFunc(handler.GenerateId))
	mux.Handle("/{id}", http.HandlerFunc(handler.GetUrl))

	err := http.ListenAndServe(":8080", mux)

	if err != nil {
		panic(err.Error())
	}
}
