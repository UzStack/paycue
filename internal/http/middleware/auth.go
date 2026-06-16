package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/UzStack/paycue/internal/domain"
	"github.com/UzStack/paycue/internal/repository"
)

type ctxKey string

const userKey ctxKey = "user"

// Auth Authorization: Bearer <token> (yoki X-API-Key) headerini tekshiradi va
// userni request contextiga joylaydi. Token noto'g'ri bo'lsa 401 qaytadi.
func Auth(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			unauthorized(w, "Authorization tokeni yo'q")
			return
		}
		user, err := repository.GetUserByToken(db, token)
		if err != nil {
			unauthorized(w, "Token noto'g'ri")
			return
		}
		ctx := context.WithValue(r.Context(), userKey, user)
		next(w, r.WithContext(ctx))
	}
}

// UserFrom request contextidan autentifikatsiya qilingan userni oladi.
func UserFrom(r *http.Request) *domain.User {
	u, _ := r.Context().Value(userKey).(*domain.User)
	return u
}

func extractToken(r *http.Request) string {
	if h := r.Header.Get("Authorization"); h != "" {
		if strings.HasPrefix(strings.ToLower(h), "bearer ") {
			return strings.TrimSpace(h[7:])
		}
		return strings.TrimSpace(h)
	}
	return strings.TrimSpace(r.Header.Get("X-API-Key"))
}

func unauthorized(w http.ResponseWriter, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(domain.Response{
		Status: false,
		Data:   domain.Detail{Detail: detail},
	})
}
