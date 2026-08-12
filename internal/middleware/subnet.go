package middleware

import (
	"net"
	"net/http"
)

// WithSubnet проверяет, что переданный в заголовке запроса X-Real-IP IP-адрес
// клиента входит в доверенную подсеть, в противном случае возвращает
// статус ответа 403 Forbidden.
func (m Middleware) WithSubnet(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		subnet := m.cfg.TrustSubnet

		if subnet == "" {
			m.logger.Warn("subnet was not set")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		_, network, err := net.ParseCIDR(subnet)
		if err != nil {
			m.logger.Warn("cannot parse CIDR")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		clientIP := r.Header.Get("X-Real-IP")

		if clientIP == "" {
			m.logger.Warn("missed X-Real-IP header")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		parsedIP := net.ParseIP(clientIP)

		if !network.Contains(parsedIP) {
			m.logger.Warnf(
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
