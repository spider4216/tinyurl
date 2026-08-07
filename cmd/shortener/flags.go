package main

import (
	"flag"
)

type Flags struct {
	ServerAddress string
	BaseUrl       string
	LogLvl        string
	FileStorePath string
	DbCon         string
	AuditFile     string
	AuditURL      string
	Https         bool
	CfgFile       string
	HttpsSet      bool
}

func NewFlags() Flags {
	return Flags{}
}

func (f *Flags) Init() error {
	host := flag.String("a", "", "Net address host:port")
	url := flag.String("b", "", "Provide base domain with protocol and port")
	logLvl := flag.String("l", "", "Log level: debug, info, warning, error, fatal")
	fileStorePath := flag.String("f", "", "File store path")
	dbCon := flag.String("d", "", "DB conection string")
	auditFile := flag.String("audit-file", "", "Audit to file")
	auditURL := flag.String("audit-url", "", "Audit to HTTP server")
	httpsMode := flag.Bool("s", false, "Audit to HTTP server")
	cfgPath := flag.String("c", "", "Config file path")

	flag.Parse()

	flag.Visit(func(fl *flag.Flag) {
		if fl.Name == "s" {
			f.HttpsSet = true
		}
	})

	f.ServerAddress = *host
	f.BaseUrl = *url
	f.LogLvl = *logLvl
	f.FileStorePath = *fileStorePath
	f.DbCon = *dbCon
	f.AuditFile = *auditFile
	f.AuditURL = *auditURL
	f.Https = *httpsMode
	f.CfgFile = *cfgPath

	return nil
}
