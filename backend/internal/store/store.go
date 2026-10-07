// Package store menyimpan data demo MedPath (Clinical Pathway Generator) di memori.
// Produksi: ganti dengan PostgreSQL. Semua data di sini FIKTIF — tanpa identitas pasien.
package store

import (
	"log"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

// User akun aplikasi.
type User struct {
	ID         int    `json:"id"`
	Nama       string `json:"nama"`
	Email      string `json:"email"`
	Peran      string `json:"peran"`
	PeranLabel string `json:"peran_label"`
	Password   string `json:"-"`
	Aktif      bool   `json:"-"`
}

// Store memegang seluruh data demo + sesi token aktif.
type Store struct {
	mu            sync.RWMutex
	users         []*User
	tokens        map[string]int           // token -> user id
	cp            []CP                     // state CP (status bisa berubah lewat transisi)
	detail        map[string]*DetailExtra  // kode -> isi klinis (acuan/rencana/kriteria/sub)
	riwayat       map[string][]RiwayatItem // kode -> riwayat pengesahan
	nextRiwayat   int
	padananKptl   map[string]penetapan // icd9cm -> KPTL
	padananSnomed map[string]penetapan // icd10 -> SNOMED
	pool          *pgxpool.Pool        // nil = murni in-memory
}

// DemoPassword dipakai semua akun demo.
const DemoPassword = "Demo-SPKD-2026"

// New membangun store in-memory (tanpa persistensi). Dipakai bila DB tidak tersedia.
func New() *Store { return newBase() }

// NewDenganDB membangun store lalu menyambungkannya ke PostgreSQL (seed bila kosong, atau load).
func NewDenganDB(pool *pgxpool.Pool) *Store {
	s := newBase()
	if pool != nil {
		if err := s.pasangDB(pool); err != nil {
			log.Printf("Gagal memakai PostgreSQL (%v) — lanjut dengan data in-memory.", err)
			s.pool = nil
		}
	}
	return s
}

func newBase() *Store {
	s := &Store{
		tokens:      map[string]int{},
		detail:      map[string]*DetailExtra{},
		riwayat:     map[string][]RiwayatItem{},
		nextRiwayat: 1,
	}
	seed := []struct{ peran, label string }{
		{"SUPER_ADMIN", "Super Admin"},
		{"KOORDINATOR_CP", "Koordinator CP"},
		{"TIM_CP", "Tim CP"},
		{"DPJP", "DPJP"},
		{"APOTEKER", "Apoteker"},
		{"KOMITE_MEDIK", "KSM/Komite Medik"},
		{"DIREKTUR", "Direktur"},
	}
	for i, p := range seed {
		s.users = append(s.users, &User{
			ID:         i + 1,
			Nama:       p.label + " (demo)",
			Email:      lower(p.peran) + "@demo.local",
			Peran:      p.peran,
			PeranLabel: p.label,
			Password:   DemoPassword,
			Aktif:      true,
		})
	}

	// State CP + isi klinis + riwayat (digenerasi dari ringkasan).
	s.cp = make([]CP, len(daftarCP))
	copy(s.cp, daftarCP)
	for i := range s.cp {
		s.cp[i].StatusLabel = statusLabel[s.cp[i].Status]
		kode := s.cp[i].Kode
		s.detail[kode] = bangunDetail(s.cp[i])
		s.riwayat[kode] = bangunRiwayatAwal(s.cp[i])
	}
	for _, r := range s.riwayat {
		if n := len(r); n > 0 && r[n-1].ID >= s.nextRiwayat {
			s.nextRiwayat = r[n-1].ID + 1
		}
	}
	s.seedPadanan()
	return s
}

// Autentikasi mengembalikan user bila email+password cocok dan akun aktif.
func (s *Store) Autentikasi(email, password string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users {
		if u.Email == email && u.Password == password && u.Aktif {
			return u, true
		}
	}
	return nil, false
}

// SimpanToken mengikat token ke user.
func (s *Store) SimpanToken(token string, userID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[token] = userID
}

// UserDariToken mengambil user dari token sesi.
func (s *Store) UserDariToken(token string) (*User, bool) {
	s.mu.RLock()
	id, ok := s.tokens[token]
	s.mu.RUnlock()
	if !ok {
		return nil, false
	}
	return s.user(id)
}

// HapusToken mengakhiri sesi (logout).
func (s *Store) HapusToken(token string) {
	s.mu.Lock()
	delete(s.tokens, token)
	s.mu.Unlock()
}

func (s *Store) user(id int) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users {
		if u.ID == id {
			return u, true
		}
	}
	return nil, false
}

func lower(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 32
		}
	}
	return string(b)
}
