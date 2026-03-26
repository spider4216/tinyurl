package main

import (
	"flag"
)

const defaultAddress = "http://127.0.0.1:8080"

type flags struct {
	srvHost string
	domain  string
}

func InitFlags() flags {
	flags := flags{
		domain: defaultAddress,
	}

	host := flag.String("a", defaultAddress, "Net address host:port")
	baseAddress := flag.String("b", defaultAddress, "Provide base domain with protocol and port")
	flag.Parse()

	if host != nil {
		flags.srvHost = *host
	}

	if baseAddress != nil {
		flags.domain = *baseAddress
	}

	return flags
}
