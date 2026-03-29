package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/spider4216/tinyurl/internal/service"
)

const maxBodySize = 2 * 1024

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

	lr := io.LimitReader(r.Body, maxBodySize)

	url, err := io.ReadAll(lr)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("cannot read body"))

		return
	}

	defer r.Body.Close()

	id := h.service.GenerateId()
	h.service.StoreData(id, string(url))

	full := fmt.Sprintf("http://%s/%s", r.Host, id)

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
