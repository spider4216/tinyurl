package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
)

const (
	defaultAddress = "http://127.0.0.1"
	defaultPort    = 8080
)

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

type flags struct {
	srvHost string
	srvPort int
	domain  string
}

func InitFlags() flags {
	flags := flags{
		srvPort: defaultPort,
		domain:  defaultAddress,
	}
	host := &host{}

	_ = flag.Value(host)

	flag.Var(host, "a", "Net address host:port")
	baseAddress := flag.String("b", defaultAddress, "Provide base domain with protocol")
	flag.Parse()

	if host.host != "" {
		flags.srvHost = host.host
	}

	if host.port != 0 {
		flags.srvPort = host.port
	}

	if *baseAddress != "" {
		flags.domain = *baseAddress
	}

	return flags
}
