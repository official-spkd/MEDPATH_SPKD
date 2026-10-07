// Package api menyediakan HTTP handler MedPath di atas net/http (Go 1.22+ ServeMux).
package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"transcpg/internal/panduan"
	"transcpg/internal/store"
)

// idParam membaca segmen {id} sebagai integer (0 bila bukan angka).
func idParam(r *http.Request) int {
	id, _ := strconv.Atoi(r.PathValue("id"))
	return id
}

// Server merangkai router dengan store data.
type Server struct {
	store *store.Store
	mux   *http.ServeMux
}

// New membangun server + mendaftarkan rute /api/v1.
func New(s *store.Store) *Server {
	srv := &Server{store: s, mux: http.NewServeMux()}
	srv.rute()
	return srv
}

// Handler mengembalikan http.Handler dengan CORS terpasang.
func (s *Server) Handler() http.Handler { return cors(s.mux) }

func (s *Server) rute() {
	// Publik
	s.mux.HandleFunc("POST /api/v1/auth/login", s.login)
	s.mux.HandleFunc("GET /api/v1/kesehatan", s.kesehatan)

	// Terproteksi (Bearer token)
	s.mux.Handle("POST /api/v1/auth/logout", s.auth(s.logout))
	s.mux.Handle("GET /api/v1/auth/me", s.auth(s.me))
	s.mux.Handle("GET /api/v1/dashboard", s.auth(s.dashboard))
	s.mux.Handle("GET /api/v1/cp", s.auth(s.daftarCP))
	s.mux.Handle("GET /api/v1/cp/{kode}", s.auth(s.detailCP))
	s.mux.Handle("GET /api/v1/cp/{kode}/riwayat", s.auth(s.riwayatCP))
	s.mux.Handle("POST /api/v1/cp/{kode}/transisi", s.auth(s.transisiCP))
	s.mux.Handle("POST /api/v1/cp/{kode}/acuan", s.auth(s.tambahAcuan))
	s.mux.Handle("DELETE /api/v1/cp/{kode}/acuan/{id}", s.auth(s.hapusAcuan))
	s.mux.Handle("POST /api/v1/cp/{kode}/kriteria", s.auth(s.tambahKriteria))
	s.mux.Handle("DELETE /api/v1/cp/{kode}/kriteria/{id}", s.auth(s.hapusKriteria))
	s.mux.Handle("POST /api/v1/cp/{kode}/rencana", s.auth(s.tambahRencana))
	s.mux.Handle("DELETE /api/v1/cp/{kode}/rencana/{id}", s.auth(s.hapusRencana))
	s.mux.Handle("GET /api/v1/approval", s.auth(s.approval))
	s.mux.Handle("GET /api/v1/dokumen-panduan", s.auth(s.dokumen))

	// Padanan kode
	s.mux.Handle("GET /api/v1/padanan/kptl", s.auth(s.daftarKptl))
	s.mux.Handle("POST /api/v1/padanan/kptl", s.auth(s.simpanKptl))
	s.mux.Handle("DELETE /api/v1/padanan/kptl/{icd9cm}", s.auth(s.hapusKptl))
	s.mux.Handle("GET /api/v1/padanan/kptl/{icd9cm}/usulan", s.auth(s.usulanKptl))
	s.mux.Handle("GET /api/v1/padanan/snomed", s.auth(s.daftarSnomed))
	s.mux.Handle("POST /api/v1/padanan/snomed", s.auth(s.simpanSnomed))
	s.mux.Handle("DELETE /api/v1/padanan/snomed/{icd10}", s.auth(s.hapusSnomed))
	s.mux.Handle("GET /api/v1/padanan/snomed/{icd10}/usulan", s.auth(s.usulanSnomed))

	// Evaluasi
	s.mux.Handle("GET /api/v1/evaluasi", s.auth(s.evaluasi))
	s.mux.Handle("GET /api/v1/evaluasi/{kode}", s.auth(s.evaluasiDetail))

	// Panduan Lengkap Super Admin
	s.mux.Handle("GET /api/v1/panduan", s.auth(s.hanyaSuperAdmin(s.panduan)))
}

// --- Handlers ---

func (s *Server) kesehatan(w http.ResponseWriter, _ *http.Request) {
	tulisJSON(w, http.StatusOK, map[string]string{"status": "ok", "layanan": "transcpg-api"})
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		tulisGalat(w, http.StatusBadRequest, "Permintaan tidak valid.")
		return
	}
	u, ok := s.store.Autentikasi(strings.TrimSpace(strings.ToLower(in.Email)), in.Password)
	if !ok {
		// Pesan sama untuk email salah & password salah.
		tulisGalat(w, http.StatusUnauthorized, "Email atau password salah.")
		return
	}
	token := tokenBaru()
	s.store.SimpanToken(token, u.ID)
	tulisJSON(w, http.StatusOK, map[string]any{"token": token, "user": u})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	s.store.HapusToken(bearerToken(r))
	tulisJSON(w, http.StatusOK, map[string]string{"message": "Berhasil keluar."})
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	tulisJSON(w, http.StatusOK, map[string]any{"user": userDari(r)})
}

func (s *Server) dashboard(w http.ResponseWriter, _ *http.Request) {
	tulisJSON(w, http.StatusOK, s.store.Dashboard())
}

func (s *Server) daftarCP(w http.ResponseWriter, _ *http.Request) {
	tulisJSON(w, http.StatusOK, map[string]any{"data": s.store.ListCP()})
}

func (s *Server) detailCP(w http.ResponseWriter, r *http.Request) {
	u := userDari(r)
	d, ok := s.store.Detail(r.PathValue("kode"), u.Peran)
	if !ok {
		tulisGalat(w, http.StatusNotFound, "Clinical Pathway tidak ditemukan.")
		return
	}
	tulisJSON(w, http.StatusOK, d)
}

func (s *Server) riwayatCP(w http.ResponseWriter, r *http.Request) {
	tulisJSON(w, http.StatusOK, map[string]any{"data": s.store.Riwayat(r.PathValue("kode"))})
}

func (s *Server) transisiCP(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Aksi   string `json:"aksi"`
		Alasan string `json:"alasan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		tulisGalat(w, http.StatusBadRequest, "Permintaan tidak valid.")
		return
	}
	cp, err := s.store.Transisi(r.PathValue("kode"), in.Aksi, userDari(r), in.Alasan)
	if err != nil {
		tulisGalat(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	tulisJSON(w, http.StatusOK, map[string]any{"cp": cp, "message": "Status CP diperbarui."})
}

func (s *Server) tambahAcuan(w http.ResponseWriter, r *http.Request) {
	var in struct {
		DokumenID int    `json:"dokumen_id"`
		Catatan   string `json:"catatan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		tulisGalat(w, http.StatusBadRequest, "Permintaan tidak valid.")
		return
	}
	s.hasilUbah(w, r, s.store.TambahAcuan(r.PathValue("kode"), userDari(r), in.DokumenID, in.Catatan))
}

func (s *Server) hapusAcuan(w http.ResponseWriter, r *http.Request) {
	s.hasilUbah(w, r, s.store.HapusAcuan(r.PathValue("kode"), userDari(r), idParam(r)))
}

func (s *Server) tambahKriteria(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Jenis  string `json:"jenis"`
		Uraian string `json:"uraian"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		tulisGalat(w, http.StatusBadRequest, "Permintaan tidak valid.")
		return
	}
	s.hasilUbah(w, r, s.store.TambahKriteria(r.PathValue("kode"), userDari(r), in.Jenis, in.Uraian))
}

func (s *Server) hapusKriteria(w http.ResponseWriter, r *http.Request) {
	s.hasilUbah(w, r, s.store.HapusKriteria(r.PathValue("kode"), userDari(r), idParam(r)))
}

func (s *Server) tambahRencana(w http.ResponseWriter, r *http.Request) {
	var in store.Rencana
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		tulisGalat(w, http.StatusBadRequest, "Permintaan tidak valid.")
		return
	}
	s.hasilUbah(w, r, s.store.TambahRencana(r.PathValue("kode"), userDari(r), in))
}

func (s *Server) hapusRencana(w http.ResponseWriter, r *http.Request) {
	s.hasilUbah(w, r, s.store.HapusRencana(r.PathValue("kode"), userDari(r), idParam(r)))
}

// hasilUbah membalas dengan detail CP terbaru atau galat.
func (s *Server) hasilUbah(w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		tulisGalat(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	d, _ := s.store.Detail(r.PathValue("kode"), userDari(r).Peran)
	tulisJSON(w, http.StatusOK, d)
}

func (s *Server) approval(w http.ResponseWriter, _ *http.Request) {
	tulisJSON(w, http.StatusOK, map[string]any{"data": s.store.Approval()})
}

func (s *Server) dokumen(w http.ResponseWriter, _ *http.Request) {
	tulisJSON(w, http.StatusOK, map[string]any{"data": s.store.ListDokumen()})
}

// --- Padanan ---

func (s *Server) daftarKptl(w http.ResponseWriter, r *http.Request) {
	tulisJSON(w, http.StatusOK, s.store.DaftarKptl(userDari(r).Peran))
}

func (s *Server) simpanKptl(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ICD9CM string `json:"icd9cm"`
		Kptl   string `json:"kptl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		tulisGalat(w, http.StatusBadRequest, "Permintaan tidak valid.")
		return
	}
	if err := s.store.SimpanKptl(userDari(r), in.ICD9CM, in.Kptl); err != nil {
		tulisGalat(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	tulisJSON(w, http.StatusOK, s.store.DaftarKptl(userDari(r).Peran))
}

func (s *Server) hapusKptl(w http.ResponseWriter, r *http.Request) {
	if err := s.store.HapusKptl(userDari(r), r.PathValue("icd9cm")); err != nil {
		tulisGalat(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	tulisJSON(w, http.StatusOK, s.store.DaftarKptl(userDari(r).Peran))
}

func (s *Server) usulanKptl(w http.ResponseWriter, r *http.Request) {
	tulisJSON(w, http.StatusOK, map[string]any{"data": s.store.UsulanKptl(r.PathValue("icd9cm"))})
}

func (s *Server) daftarSnomed(w http.ResponseWriter, r *http.Request) {
	tulisJSON(w, http.StatusOK, s.store.DaftarSnomed(userDari(r).Peran))
}

func (s *Server) simpanSnomed(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ICD10  string `json:"icd10"`
		Snomed string `json:"snomed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		tulisGalat(w, http.StatusBadRequest, "Permintaan tidak valid.")
		return
	}
	if err := s.store.SimpanSnomed(userDari(r), in.ICD10, in.Snomed); err != nil {
		tulisGalat(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	tulisJSON(w, http.StatusOK, s.store.DaftarSnomed(userDari(r).Peran))
}

func (s *Server) hapusSnomed(w http.ResponseWriter, r *http.Request) {
	if err := s.store.HapusSnomed(userDari(r), r.PathValue("icd10")); err != nil {
		tulisGalat(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	tulisJSON(w, http.StatusOK, s.store.DaftarSnomed(userDari(r).Peran))
}

func (s *Server) usulanSnomed(w http.ResponseWriter, r *http.Request) {
	tulisJSON(w, http.StatusOK, map[string]any{"data": s.store.UsulanSnomed(r.PathValue("icd10"))})
}

// --- Panduan ---

func (s *Server) panduan(w http.ResponseWriter, _ *http.Request) {
	daftar, err := panduan.Daftar()
	if err != nil {
		tulisGalat(w, http.StatusInternalServerError, "Panduan gagal dimuat.")
		return
	}
	tulisJSON(w, http.StatusOK, map[string]any{"data": daftar})
}

// --- Evaluasi ---

func (s *Server) evaluasi(w http.ResponseWriter, _ *http.Request) {
	tulisJSON(w, http.StatusOK, s.store.Evaluasi())
}

func (s *Server) evaluasiDetail(w http.ResponseWriter, r *http.Request) {
	det, err := s.store.EvaluasiDetailCP(r.PathValue("kode"))
	if err != nil {
		tulisGalat(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	tulisJSON(w, http.StatusOK, det)
}

// --- Util ---

func tokenBaru() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func tulisJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func tulisGalat(w http.ResponseWriter, status int, pesan string) {
	tulisJSON(w, status, map[string]string{"message": pesan})
}
