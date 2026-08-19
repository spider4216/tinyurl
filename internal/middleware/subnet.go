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
		sub := m.cfg.TrustSubnet

		if sub == "" {
			m.logger.Warn("subnet was not set")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		subnet := m.cfg.ParsedSubnet

		clientIP := r.Header.Get("X-Real-IP")

		if clientIP == "" {
			m.logger.Warn("missed X-Real-IP header")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		parsedIP := net.ParseIP(clientIP)

		if !subnet.Contains(parsedIP) {
			m.logger.Warnf(
				"client IP %s not match subnet %s with mask %s",
				clientIP,
				subnet.IP.String(),
				subnet.Mask.String(),
			)

			w.WriteHeader(http.StatusForbidden)
			return
		}

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
