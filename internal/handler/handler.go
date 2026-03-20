package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/spider4216/tinyurl/internal/service"
)

func New(service service.Service) Handler {
	return Handler{
		service: service,
	}
}

type Handler struct {
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

	full := fmt.Sprintf("http://%s/%s", r.Host, id)

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

	w.Write([]byte(url))
}
