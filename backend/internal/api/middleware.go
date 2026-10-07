package api

import (
	"context"
	"net/http"
	"strings"

	"transcpg/internal/store"
)

type ctxKey int

const kunciUser ctxKey = 0

// auth memverifikasi Bearer token dan menaruh user di context.
func (s *Server) auth(h http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		if token == "" {
			tulisGalat(w, http.StatusUnauthorized, "Token tidak ada.")
			return
		}
		u, ok := s.store.UserDariToken(token)
		if !ok {
			tulisGalat(w, http.StatusUnauthorized, "Sesi berakhir, silakan masuk lagi.")
			return
		}
		ctx := context.WithValue(r.Context(), kunciUser, u)
		h(w, r.WithContext(ctx))
	})
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if after, ok := strings.CutPrefix(h, "Bearer "); ok {
		return strings.TrimSpace(after)
	}
	return ""
}

func userDari(r *http.Request) *store.User {
	u, _ := r.Context().Value(kunciUser).(*store.User)
	return u
}

// cors mengizinkan frontend (dev) mengakses API lintas-origin.
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept")
		w.Header().Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
