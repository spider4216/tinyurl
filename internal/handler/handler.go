package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/spider4216/tinyurl/internal/config"
	"github.com/spider4216/tinyurl/internal/service"
)

const maxBodySize = 2 * 1024

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

		if _, err := w.Write([]byte("Method not allowed")); err != nil {
			log.Println("failed to write response:", err)
		}

		return
	}

	lr := io.LimitReader(r.Body, maxBodySize)

	url, err := io.ReadAll(lr)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		if _, err := w.Write([]byte("cannot read body")); err != nil {
			log.Println("failed to write response:", err)
		}

		return
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Println("failed to close request body:", err)
		}
	}()

	id := h.service.GenerateId()
	h.service.StoreData(id, string(url))

	full := fmt.Sprintf("%s/%s", h.conf.Domain, id)

	w.Header().Set("Content-Type", "plain/text")
	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write([]byte(full)); err != nil {
		log.Println("failed to write response:", err)
	}
}

func (h Handler) GetUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		if _, err := w.Write([]byte("Method not allowed")); err != nil {
			log.Println("failed to write response:", err)
		}

		return
	}

	id := r.PathValue("id")

	url := h.service.GetUrl(id)

	if url == "" {
		w.WriteHeader(http.StatusNotFound)

		if _, err := w.Write([]byte("Url not found")); err != nil {
			log.Println("failed to write response:", err)
		}

		return
	}

	w.Header().Set("Content-Type", "plain/text")
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
