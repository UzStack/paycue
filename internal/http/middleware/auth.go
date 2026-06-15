package middleware

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"

	"github.com/UzStack/paycue/internal/domain"
)

// APIKey X-API-Key header orqali kelgan kalitni configdagi kalit bilan
// solishtiradi. Solishtirish timing-attackdan himoya uchun
// constant-time amalga oshiriladi.
func APIKey(key string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provided := r.Header.Get("X-API-Key")
		if subtle.ConstantTimeCompare([]byte(provided), []byte(key)) != 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(domain.Response{
				Status: false,
				Data: domain.Detail{
					Detail: "Invalid or missing X-API-Key",
				},
			})
			return
		}
		next(w, r)
	}
}
