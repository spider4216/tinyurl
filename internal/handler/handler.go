package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spider4216/tinyurl/internal/config"
	"github.com/spider4216/tinyurl/internal/service"
)

func New(conf config.Config, service service.Service) Handler {
	return Handler{
		conf:    conf,
		service: service,
	}
}

type Handler struct {
	conf    config.Config
	service service.Service
}

func (h Handler) GenerateId(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))

		return
	}

	url, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("cannot read body"))

		return
	}

	defer r.Body.Close()

	id := h.service.GenerateId()
	h.service.StoreData(id, string(url))

	full := fmt.Sprintf("%s:%d/%s", h.conf.Domain, h.conf.ServerPort, id)

	// Игнорировать порт если передан SERVER_PORT
	if sp := os.Getenv("SERVER_PORT"); sp != "" {
		full = fmt.Sprintf("%s/%s", h.conf.Domain, id)
	}

	w.Header().Set("Content-Type", "plain/text")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(full))
}

func (h Handler) GetUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))

		return
	}

	id := r.PathValue("id")

	url := h.service.GetUrl(id)

	if url == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Url not found"))

		return
	}

	w.Header().Set("Content-Type", "plain/text")
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
