package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/spider4216/tinyurl/internal/config"
	"github.com/spider4216/tinyurl/internal/handler"
	"github.com/spider4216/tinyurl/internal/repository"
	"github.com/spider4216/tinyurl/internal/service"
)

const (
	defaultAddress = "http://127.0.0.1"
	defaultPort    = 8080
)

func main() {
	flags := InitFlags()

	conf := config.New(flags.Domain, flags.SrvHost, flags.SrvPort)
	store := map[string]string{}
	repo := repository.New(store)
	service := service.New(repo)
	handler := handler.New(conf, service)

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Post("/", http.HandlerFunc(handler.GenerateId))
		r.Get("/{id}", http.HandlerFunc(handler.GetUrl))
	})

	err := http.ListenAndServe(conf.HostAsString(), r)

	if err != nil {
		panic(err.Error())
	}
}

type host struct {
	host string
	port int
}

func (h *host) String() string {
	return fmt.Sprintf("%s:%d", h.host, h.port)
}

func (h *host) Set(flagValue string) error {
	parts := strings.Split(flagValue, ":")
	h.host = parts[0]
	v, err := strconv.Atoi(parts[1])

	if err != nil {
		return err
	}

	h.port = v

	return nil
}

type Flags struct {
	SrvHost string
	SrvPort int
	Domain  string
}

func InitFlags() Flags {
	flags := Flags{
		SrvPort: defaultPort,
		Domain:  defaultAddress,
	}
	host := &host{}

	_ = flag.Value(host)

	flag.Var(host, "a", "Net address host:port")
	baseAddress := flag.String("b", defaultAddress, "Provide base domain with protocol")
	flag.Parse()

	if host.host != "" {
		flags.SrvHost = host.host
	}

	if host.port != 0 {
		flags.SrvPort = host.port
	}

	if *baseAddress != "" {
		flags.Domain = *baseAddress
	}

	return flags
}
