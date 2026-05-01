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
	FileStorePath string
	DbCon         string
}

func NewFlags() Flags {
	return Flags{}
}

func (f *Flags) Init() error {
	host := flag.String("a", defSrvAddr, "Net address host:port")
	url := flag.String("b", defBaseUrl, "Provide base domain with protocol and port")
	logLvl := flag.String("l", defLogLvl, "Log level: debug, info, warning, error, fatal")
	fileStorePath := flag.String("f", "", "File store path")
	dbCon := flag.String("d", "", "DB conection string")

	flag.Parse()

	f.ServerAddress = *host
	f.BaseUrl = *url
	f.LogLvl = *logLvl
	f.FileStorePath = *fileStorePath
	f.DbCon = *dbCon

	return nil
}
