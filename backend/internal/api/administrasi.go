package api

import (
	"encoding/json"
	"net/http"
	"transcpg/internal/store"
)

func (s *Server) tambahPengguna(w http.ResponseWriter, r *http.Request) {
	var p struct {
		Nama     string `json:"nama"`
		Email    string `json:"email"`
		Peran    string `json:"peran"`
		Password string `json:"password"`
	}
	if json.NewDecoder(r.Body).Decode(&p) != nil {
		tulisGalat(w, 400, "Permintaan tidak valid.")
		return
	}
	if err := s.store.TambahPengguna(p.Nama, p.Email, p.Peran, p.Password); err != nil {
		tulisGalat(w, 422, err.Error())
		return
	}
	s.administrasi(w, r)
}
func (s *Server) nonaktifkanDemo(w http.ResponseWriter, r *http.Request) {
	if err := s.store.NonaktifkanDemo(userDari(r)); err != nil {
		tulisGalat(w, 422, err.Error())
		return
	}
	s.administrasi(w, r)
}

func (s *Server) hanyaAdmin(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := userDari(r)
		if u == nil || (u.Peran != "SUPER_ADMIN" && u.Peran != "SYSTEM_ADMIN" && u.Peran != "ADMIN_RS") {
			tulisGalat(w, 403, "Akses ini khusus administrator.")
			return
		}
		h(w, r)
	}
}
func (s *Server) administrasi(w http.ResponseWriter, r *http.Request) {
	tulisJSON(w, 200, s.store.Administrasi())
}
func (s *Server) ubahProfil(w http.ResponseWriter, r *http.Request) {
	var p store.ProfilRS
	if json.NewDecoder(r.Body).Decode(&p) != nil {
		tulisGalat(w, 400, "Permintaan tidak valid.")
		return
	}
	if err := s.store.UbahProfil(p); err != nil {
		tulisGalat(w, 422, err.Error())
		return
	}
	s.administrasi(w, r)
}
func (s *Server) ubahPengguna(w http.ResponseWriter, r *http.Request) {
	var p store.PenggunaAdmin
	if json.NewDecoder(r.Body).Decode(&p) != nil {
		tulisGalat(w, 400, "Permintaan tidak valid.")
		return
	}
	if err := s.store.UbahPengguna(p, userDari(r)); err != nil {
		tulisGalat(w, 422, err.Error())
		return
	}
	s.administrasi(w, r)
}
func (s *Server) buatKunci(w http.ResponseWriter, r *http.Request) {
	var p struct {
		Nama string `json:"nama"`
	}
	if json.NewDecoder(r.Body).Decode(&p) != nil {
		tulisGalat(w, 400, "Permintaan tidak valid.")
		return
	}
	key, err := s.store.BuatKunci(p.Nama)
	if err != nil {
		tulisGalat(w, 422, err.Error())
		return
	}
	tulisJSON(w, 201, map[string]string{"kunci": key})
}
func (s *Server) cabutKunci(w http.ResponseWriter, r *http.Request) {
	if err := s.store.CabutKunci(r.PathValue("id")); err != nil {
		tulisGalat(w, 404, err.Error())
		return
	}
	s.administrasi(w, r)
}
func (s *Server) integrasiCP(w http.ResponseWriter, r *http.Request) {
	if !s.store.KunciValid(bearerToken(r)) {
		tulisGalat(w, 401, "Kunci integrasi tidak valid atau sudah dicabut.")
		return
	}
	rows := []store.CP{}
	for _, c := range s.store.ListCP() {
		if c.Status == "AKTIF" {
			rows = append(rows, c)
		}
	}
	tulisJSON(w, 200, map[string]any{"data": rows})
}
func (s *Server) tambahDokumen(w http.ResponseWriter, r *http.Request) {
	var p store.Dokumen
	if json.NewDecoder(r.Body).Decode(&p) != nil {
		tulisGalat(w, 400, "Permintaan tidak valid.")
		return
	}
	d, err := s.store.TambahDokumen(p)
	if err != nil {
		tulisGalat(w, 422, err.Error())
		return
	}
	tulisJSON(w, 201, d)
}
func (s *Server) detailDokumen(w http.ResponseWriter, r *http.Request) {
	d, ok := s.store.DetailDokumen(idParam(r))
	if !ok {
		tulisGalat(w, 404, "Dokumen tidak ditemukan.")
		return
	}
	tulisJSON(w, 200, d)
}
