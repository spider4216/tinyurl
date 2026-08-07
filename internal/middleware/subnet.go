package middleware

import (
	"net"
	"net/http"
)

func (m Middleware) WithSubnet(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		subnet := m.cfg.TrustSubnet

		if subnet == "" {
			m.logger.Error("subnet was not set")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		_, network, err := net.ParseCIDR(subnet)

		if err != nil {
			m.logger.Error("cannot parse CIDR")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		clientIP := r.Header.Get("X-Real-IP")

		if clientIP == "" {
			m.logger.Error("missed X-Real-IP header")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		parsedIP := net.ParseIP(clientIP)

		if !network.Contains(parsedIP) {
			m.logger.Errorf(
				"client IP %s not match subnet %s with mask %s",
				clientIP,
				network.IP.String(),
				network.Mask.String(),
			)

			w.WriteHeader(http.StatusForbidden)
			return
		}

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
