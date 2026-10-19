// Package store mengelola data demo MedPath dengan snapshot transaksi PostgreSQL.
// Semua data awal FIKTIF — tanpa identitas pasien.
package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
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
	klaim         []Klaim
	administrasi  Administrasi
	mu            sync.RWMutex
	users         []*User
	tokens        map[string]int           // token -> user id
	cp            []CP                     // state CP (status bisa berubah lewat transisi)
	detail        map[string]*DetailExtra  // kode -> isi klinis (acuan/rencana/kriteria/sub)
	riwayat       map[string][]RiwayatItem // kode -> riwayat pengesahan
	nextRiwayat   int
	padananKptl   map[string]penetapan // icd9cm -> KPTL
	padananSnomed map[string]penetapan // icd10 -> SNOMED
	pool          *pgxpool.Pool        // pool bersama; state mutable dibuat ulang per request
	db            pgx.Tx               // transaksi milik satu request
	ctx           context.Context
	dbErr         error
}

// DemoPassword dipakai semua akun demo.
const DemoPassword = "Demo-SPKD-2026"

// New membangun store in-memory untuk pengujian dan demo lokal eksplisit.
func New() *Store { return newBase() }

// NewDenganDB gagal bila database tidak siap; tidak ada fallback yang kehilangan data.
func NewDenganDB(pool *pgxpool.Pool) (*Store, error) {
	if pool == nil {
		return nil, errors.New("PostgreSQL wajib tersedia")
	}
	s := newBase()
	s.pool = pool
	if err := s.pasangDB(pool); err != nil {
		return nil, err
	}
	return s, nil
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
		{"KOMITE_MEDIK", "Komite Medik"},
		{"DIREKTUR", "Direktur"},
	}
	for i, p := range seed {
		s.users = append(s.users, &User{
			ID:         i + 1,
			Nama:       p.label + " (demo)",
			Email:      lower(p.peran) + "@demo.local",
			Peran:      p.peran,
			PeranLabel: p.label,
			Password:   demoPasswordHash(),
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
	s.adminAwal()
	return s
}

// Autentikasi mengembalikan user bila email+password cocok dan akun aktif.
func (s *Store) Autentikasi(email, password string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users {
		if u.Email == email && u.Aktif && cocokPassword(u.Password, password) {
			return u, true
		}
	}
	return nil, false
}

// Token mentah hanya diberikan ke klien; database menyimpan hash dan masa berlaku.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *Store) SimpanToken(token string, userID int) {
	if s.db != nil {
		s.execDB(`INSERT INTO auth_sessions(token_hash,user_id,expires_at) VALUES($1,$2,$3)`, hashToken(token), userID, time.Now().Add(24*time.Hour))
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[token] = userID
}

func (s *Store) UserDariToken(token string) (*User, bool) {
	if s.db != nil {
		var id int
		err := s.db.QueryRow(s.ctx, `SELECT user_id FROM auth_sessions WHERE token_hash=$1 AND expires_at>now()`, hashToken(token)).Scan(&id)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false
		}
		if err != nil {
			s.dbErr = err
			return nil, false
		}
		return s.user(id)
	}
	s.mu.RLock()
	id, ok := s.tokens[token]
	s.mu.RUnlock()
	if !ok {
		return nil, false
	}
	return s.user(id)
}

func (s *Store) HapusToken(token string) {
	if s.db != nil {
		s.execDB(`DELETE FROM auth_sessions WHERE token_hash=$1`, hashToken(token))
		return
	}
	s.mu.Lock()
	delete(s.tokens, token)
	s.mu.Unlock()
}

func (s *Store) user(id int) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users {
		if u.ID == id && u.Aktif {
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
