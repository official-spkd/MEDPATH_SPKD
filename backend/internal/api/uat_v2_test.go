package api

import (
	"net/http/httptest"
	"strings"
	"testing"
	"transcpg/internal/store"
)

func TestUATV2KontrakHTTP(t *testing.T) {
	s := store.New()
	s.SimpanToken("qa-tim", 3)
	s.SimpanToken("qa-komite", 6)
	s.SimpanToken("qa-super", 1)
	h := New(s).Handler()
	tests := []struct {
		name, method, path, token, body string
		status                          int
		contains                        string
	}{
		{"TC-API-001", "GET", "kesehatan", "", "", 200, "medpath-api"},
		{"TC-API-015", "DELETE", "cp/C-4-12/acuan/999", "qa-tim", "", 404, "Item tidak ditemukan"},
		{"TC-PGS-009", "POST", "cp/I-4-14/transisi", "qa-komite", `{"aksi":"KEMBALIKAN","alasan":"  "}`, 422, "Alasan pengembalian wajib"},
		{"TC-KPTL-013", "POST", "padanan/kptl", "qa-tim", `{"icd9cm":"00.00","kptl":"KPTL-01"}`, 422, "tidak ada di daftar"},
		{"AdminProtected", "GET", "admin", "qa-tim", "", 403, "administrator"},
		{"AdminAllowed", "GET", "admin", "qa-super", "", 200, "profil"},
		{"DokumenProtected", "POST", "dokumen-panduan", "qa-tim", `{}`, 403, "administrator"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/api/v1/"+tc.path, strings.NewReader(tc.body))
			if tc.token != "" {
				r.Header.Set("Authorization", "Bearer "+tc.token)
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != tc.status || !strings.Contains(w.Body.String(), tc.contains) {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
	t.Setenv("ALLOWED_ORIGINS", "https://frontend.example")
	for _, origin := range []string{"https://frontend.example", "https://contoh-asing.test"} {
		r := httptest.NewRequest("OPTIONS", "/api/v1/kesehatan", nil)
		r.Header.Set("Origin", origin)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if origin == "https://frontend.example" {
			if w.Header().Get("Access-Control-Allow-Origin") != origin || w.Code != 204 {
				t.Fatal("origin resmi ditolak")
			}
		} else if w.Header().Get("Access-Control-Allow-Origin") != "" || w.Code != 403 {
			t.Fatal("origin asing diizinkan")
		}
	}
}
