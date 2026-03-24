package config

import (
	"fmt"
)

type Config struct {
	ServerHost string
	ServerPort int
	Domain     string
}

func New(domain string, srvbHost string, srvPort int) Config {
	return Config{
		Domain:     domain,
		ServerHost: srvbHost,
		ServerPort: srvPort,
	}
}

func (c *Config) HostAsString() string {
	return fmt.Sprintf("%s:%d", c.ServerHost, c.ServerPort)
}
