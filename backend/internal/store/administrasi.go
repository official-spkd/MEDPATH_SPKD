package store

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"strings"
	"time"
)

type ProfilRS struct {
	Nama        string `json:"nama"`
	Tipe        string `json:"tipe"`
	Kepemilikan string `json:"kepemilikan"`
	Regional    string `json:"regional"`
	Kelas       string `json:"kelas"`
}
type PenggunaAdmin struct {
	ID    int    `json:"id"`
	Nama  string `json:"nama"`
	Email string `json:"email"`
	Peran string `json:"peran"`
	Aktif bool   `json:"aktif"`
}
type KunciIntegrasi struct {
	ID     string `json:"id"`
	Nama   string `json:"nama"`
	Dibuat string `json:"dibuat"`
	Aktif  bool   `json:"aktif"`
	Hash   string `json:"hash,omitempty"`
}
type Administrasi struct {
	DemoDinonaktifkan bool             `json:"demo_dinonaktifkan"`
	Profil            ProfilRS         `json:"profil"`
	Pengguna          []PenggunaAdmin  `json:"pengguna"`
	Kunci             []KunciIntegrasi `json:"kunci"`
	Dokumen           []Dokumen        `json:"dokumen"`
}

func (s *Store) adminAwal() {
	s.administrasi = Administrasi{Profil: ProfilRS{"RSUD Demo SPKD", "B", "Pemerintah", "1", "1"}, Kunci: []KunciIntegrasi{}, Dokumen: append([]Dokumen{}, daftarDokumen...)}
	for _, u := range s.users {
		s.administrasi.Pengguna = append(s.administrasi.Pengguna, PenggunaAdmin{u.ID, u.Nama, u.Email, u.Peran, u.Aktif})
	}
}
func (s *Store) loadAdministrasi() error {
	var raw []byte
	err := s.db.QueryRow(s.ctx, "SELECT isi FROM app_state WHERE kunci='administrasi'").Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, &s.administrasi); err != nil {
		return err
	}
	s.terapkanPengguna()
	return nil
}
func (s *Store) terapkanPengguna() {
	for _, p := range s.administrasi.Pengguna {
		for _, u := range s.users {
			if u.ID == p.ID {
				u.Nama = p.Nama
				u.Peran = p.Peran
				u.PeranLabel = labelPeran[p.Peran]
				u.Aktif = p.Aktif
			}
		}
	}
	if s.administrasi.DemoDinonaktifkan {
		for _, u := range s.users {
			if strings.HasSuffix(u.Email, "@demo.local") {
				u.Aktif = false
			}
		}
		for i := range s.administrasi.Pengguna {
			if strings.HasSuffix(s.administrasi.Pengguna[i].Email, "@demo.local") {
				s.administrasi.Pengguna[i].Aktif = false
			}
		}
	}
}
func (s *Store) simpanAdministrasi() {
	if s.db == nil {
		return
	}
	raw, err := json.Marshal(s.administrasi)
	if err != nil {
		s.dbErr = err
		return
	}
	s.execDB("INSERT INTO app_state(kunci,isi) VALUES('administrasi',$1::jsonb) ON CONFLICT(kunci) DO UPDATE SET isi=excluded.isi", string(raw))
}
func (s *Store) Administrasi() Administrasi {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Salinan dalam agar pemanggil tidak dapat mengubah state atau memperoleh hash.
	raw, _ := json.Marshal(s.administrasi)
	var out Administrasi
	_ = json.Unmarshal(raw, &out)
	for i := range out.Kunci {
		out.Kunci[i].Hash = ""
	}
	return out
}
func (s *Store) UbahProfil(p ProfilRS) error {
	p.Nama = strings.TrimSpace(p.Nama)
	if p.Nama == "" || p.Tipe == "" || p.Regional == "" || p.Kelas == "" {
		return errors.New("Nama, tipe, regional, dan kelas rumah sakit wajib diisi.")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.administrasi.Profil = p
	s.simpanAdministrasi()
	return nil
}
func (s *Store) UbahPengguna(p PenggunaAdmin, actor *User) error {
	if actor.Peran != "SUPER_ADMIN" {
		return errors.New("Perubahan pengguna khusus Super Admin.")
	}
	if p.ID == actor.ID {
		return errors.New("Akun sendiri tidak dapat diubah melalui pengelolaan ini.")
	}
	if labelPeran[p.Peran] == "" || strings.TrimSpace(p.Nama) == "" {
		return errors.New("Nama dan peran yang valid wajib diisi.")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, u := range s.administrasi.Pengguna {
		if u.ID == p.ID {
			p.Email = u.Email
			p.Nama = strings.TrimSpace(p.Nama)
			s.administrasi.Pengguna[i] = p
			s.terapkanPengguna()
			s.simpanAdministrasi()
			return nil
		}
	}
	return ErrItemTidakDitemukan
}
func (s *Store) BuatKunci(nama string) (string, error) {
	nama = strings.TrimSpace(nama)
	if nama == "" {
		return "", errors.New("Nama integrasi wajib diisi.")
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	raw := "mp_" + hex.EncodeToString(b)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.administrasi.Kunci = append(s.administrasi.Kunci, KunciIntegrasi{hex.EncodeToString(b[:8]), nama, time.Now().UTC().Format(time.RFC3339), true, hashToken(raw)})
	s.simpanAdministrasi()
	return raw, nil
}
func (s *Store) CabutKunci(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, k := range s.administrasi.Kunci {
		if k.ID == id {
			s.administrasi.Kunci[i].Aktif = false
			s.simpanAdministrasi()
			return nil
		}
	}
	return ErrItemTidakDitemukan
}
func (s *Store) KunciValid(raw string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	h := hashToken(raw)
	for _, k := range s.administrasi.Kunci {
		if k.Aktif && k.Hash == h {
			return true
		}
	}
	return false
}
func (s *Store) TambahDokumen(d Dokumen) (*Dokumen, error) {
	d.Nama = strings.TrimSpace(d.Nama)
	d.Penerbit = strings.TrimSpace(d.Penerbit)
	if d.Nama == "" || d.Penerbit == "" || d.Tahun < 1900 || d.Tahun > time.Now().Year()+1 {
		return nil, errors.New("Nama, penerbit, dan tahun dokumen yang valid wajib diisi.")
	}
	if d.Sumber != "PNPK" && d.Sumber != "PPK_RS" && d.Sumber != "PPK_ASOSIASI" {
		return nil, errors.New("Sumber dokumen tidak valid.")
	}
	isi := []string{}
	for _, v := range d.IsiButir {
		if t := strings.TrimSpace(v); t != "" {
			isi = append(isi, t)
		}
	}
	d.IsiButir = isi
	d.Butir = len(isi)
	s.mu.Lock()
	defer s.mu.Unlock()
	d.ID = maxID(s.administrasi.Dokumen, func(d Dokumen) int { return d.ID })
	s.administrasi.Dokumen = append(s.administrasi.Dokumen, d)
	s.simpanAdministrasi()
	return &d, nil
}
func (s *Store) DetailDokumen(id int) (*Dokumen, bool) {
	for _, d := range s.ListDokumen() {
		if d.ID == id {
			return &d, true
		}
	}
	return nil, false
}
