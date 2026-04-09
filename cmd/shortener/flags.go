package main

import (
	"flag"
)

const (
	defSrvAddr = "127.0.0.1:8080"
	defBaseUrl = "http://127.0.0.1:8080"
	defLogLvl  = "info"
)

type Flags struct {
	ServerAddress string
	BaseUrl       string
	LogLvl        string
}

func NewFlags() Flags {
	return Flags{}
}

func (f *Flags) Init() {
	host := flag.String("a", defSrvAddr, "Net address host:port")
	url := flag.String("b", defBaseUrl, "Provide base domain with protocol and port")
	logLvl := flag.String("l", defLogLvl, "Log level: debug, info, warning, error, fatal")

	flag.Parse()

	f.ServerAddress = *host
	f.BaseUrl = *url
	f.LogLvl = *logLvl
}
