package main

import (
	"flag"
	"os"

	"github.com/spider4216/tinyurl/internal/storage"
)

const (
	defSrvAddr     = "127.0.0.1:8080"
	defBaseUrl     = "http://127.0.0.1:8080"
	defLogLvl      = "info"
	defStoreDriver = storage.FileDriver
	defDbCon       = "postgres://user:password@localhost:5432/dbname"
)

type Flags struct {
	ServerAddress string
	BaseUrl       string
	LogLvl        string
	StoreDriver   string
	FileStorePath string
	DbCon         string
}

func NewFlags() Flags {
	return Flags{}
}

func (f *Flags) Init() error {
	defPath, err := os.Getwd()

	if err != nil {
		return err
	}

	defPath += "/store.json"

	host := flag.String("a", defSrvAddr, "Net address host:port")
	url := flag.String("b", defBaseUrl, "Provide base domain with protocol and port")
	logLvl := flag.String("l", defLogLvl, "Log level: debug, info, warning, error, fatal")
	storeDriver := flag.String("s", defStoreDriver, "Store driver: file,map,pgx")
	fileStorePath := flag.String("f", defPath, "File store path")
	dbCon := flag.String("d", defDbCon, "DB conection string")

	flag.Parse()

	f.ServerAddress = *host
	f.BaseUrl = *url
	f.LogLvl = *logLvl
	f.StoreDriver = *storeDriver
	f.FileStorePath = *fileStorePath
	f.DbCon = *dbCon

	return nil
}
