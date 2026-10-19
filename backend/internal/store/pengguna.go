package store

import (
	"errors"
	"net/mail"
	"strings"
)

func (s *Store) loadUsers() error {
	rows, err := s.db.Query(s.ctx, "SELECT id,nama,email,peran,password_hash,aktif FROM app_users ORDER BY id")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		u := &User{}
		if err := rows.Scan(&u.ID, &u.Nama, &u.Email, &u.Peran, &u.Password, &u.Aktif); err != nil {
			return err
		}
		u.PeranLabel = labelPeran[u.Peran]
		s.users = append(s.users, u)
	}
	return rows.Err()
}
func (s *Store) TambahPengguna(nama, email, peran, password string) error {
	nama = strings.TrimSpace(nama)
	email = strings.ToLower(strings.TrimSpace(email))
	a, err := mail.ParseAddress(email)
	if err != nil || a.Address != email || strings.HasSuffix(email, "@demo.local") || nama == "" || labelPeran[peran] == "" {
		return errors.New("Nama, email non-demo, dan peran yang valid wajib diisi.")
	}
	if len(password) < 12 {
		return errors.New("Kata sandi minimal 12 karakter.")
	}
	encoded := hashPassword(password)
	s.mu.Lock()
	defer s.mu.Unlock()
	id := 1
	for _, u := range s.users {
		if u.Email == email {
			return errors.New("Email sudah digunakan.")
		}
		if u.ID >= id {
			id = u.ID + 1
		}
	}
	u := &User{ID: id, Nama: nama, Email: email, Peran: peran, PeranLabel: labelPeran[peran], Password: encoded, Aktif: true}
	s.users = append(s.users, u)
	s.administrasi.Pengguna = append(s.administrasi.Pengguna, PenggunaAdmin{id, nama, email, peran, true})
	if s.db != nil {
		s.execDB("INSERT INTO app_users(id,nama,email,peran,password_hash,aktif) VALUES($1,$2,$3,$4,$5,true)", id, nama, email, peran, encoded)
	}
	s.simpanAdministrasi()
	return nil
}
func (s *Store) NonaktifkanDemo(actor *User) error {
	if actor.Peran != "SUPER_ADMIN" || strings.HasSuffix(actor.Email, "@demo.local") {
		return errors.New("Masuk dengan Super Admin non-demo terlebih dahulu.")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.administrasi.DemoDinonaktifkan = true
	s.terapkanPengguna()
	s.simpanAdministrasi()
	return nil
}
