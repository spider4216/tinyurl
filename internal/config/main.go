package config

type Config struct {
	ServerHost string
	Domain     string
}

func New(domain string, srvbHost string) Config {
	return Config{
		Domain:     domain,
		ServerHost: srvbHost,
	}
}
