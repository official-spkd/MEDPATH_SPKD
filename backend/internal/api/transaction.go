package api

import (
	"bytes"
	"errors"
	"log"
	"net/http"

	"transcpg/internal/store"
)

var errRejected = errors.New("request rejected")

type bufferedResponse struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (b *bufferedResponse) Header() http.Header { return b.header }
func (b *bufferedResponse) WriteHeader(status int) {
	if b.status == 0 {
		b.status = status
	}
}
func (b *bufferedResponse) Write(p []byte) (int, error) {
	if b.status == 0 {
		b.status = http.StatusOK
	}
	return b.body.Write(p)
}

func (s *Server) transactional(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	response := &bufferedResponse{header: make(http.Header)}
	write := r.Method != http.MethodGet && r.Method != http.MethodHead
	err := s.store.Request(r.Context(), write, func(current *store.Store) error {
		New(current).mux.ServeHTTP(response, r)
		if response.status >= 400 {
			return errRejected
		}
		return nil
	})
	if err != nil && !errors.Is(err, errRejected) {
		// Jangan log DSN, parameter SQL, atau isi klinis.
		log.Printf("MedPath: transaksi database gagal (%T)", err)
		tulisGalat(w, http.StatusServiceUnavailable, "Database belum tersedia. Silakan coba lagi.")
		return
	}
	for key, values := range response.header {
		w.Header()[key] = values
	}
	if response.status != 0 {
		w.WriteHeader(response.status)
	}
	_, _ = w.Write(response.body.Bytes())
}
