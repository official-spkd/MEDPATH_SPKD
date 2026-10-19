package store

import (
	"errors"
	"strings"
)

var ErrItemTidakDitemukan = errors.New("Item tidak ditemukan.")

var errTerkunci = errors.New("Isi CP terkunci pada status ini atau peran Anda tidak berwenang.")

func (s *Store) detailUntuk(kode string) (*CP, *DetailExtra, bool) {
	cp := s.cpByKode(kode)
	if cp == nil {
		return nil, nil, false
	}
	d := s.detail[kode]
	if d == nil {
		d = bangunDetail(*cp)
		s.detail[kode] = d
	}
	return cp, d, true
}

func maxID[T any](items []T, id func(T) int) int {
	m := 0
	for _, it := range items {
		if v := id(it); v > m {
			m = v
		}
	}
	return m + 1
}

// TambahAcuan menautkan dokumen panduan sebagai acuan CP.
func (s *Store) TambahAcuan(kode string, u *User, dokumenID int, catatan string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp, d, ok := s.detailUntuk(kode)
	if !ok {
		return errors.New("CP tidak ditemukan.")
	}
	if !bolehUbah(u.Peran, cp.Status, ISIAcuan) {
		return errTerkunci
	}
	var dk *Dokumen
	for i := range s.administrasi.Dokumen {
		if s.administrasi.Dokumen[i].ID == dokumenID {
			dk = &s.administrasi.Dokumen[i]
			break
		}
	}
	if dk == nil {
		return errors.New("Dokumen panduan tidak ditemukan.")
	}
	for _, a := range d.Acuan {
		if a.Dokumen == dk.Nama && a.Tahun == dk.Tahun {
			return errors.New("Dokumen ini sudah menjadi acuan CP.")
		}
	}
	a := Acuan{
		ID:      maxID(d.Acuan, func(a Acuan) int { return a.ID }),
		Dokumen: dk.Nama, Sumber: dk.Sumber, Tahun: dk.Tahun, Catatan: catatan,
	}
	d.Acuan = append(d.Acuan, a)
	cp.JumlahAcuan = len(d.Acuan)
	s.simpanAcuanDB(kode, a)
	s.simpanCPState(cp)
	return nil
}

// HapusAcuan melepas acuan.
func (s *Store) HapusAcuan(kode string, u *User, id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp, d, ok := s.detailUntuk(kode)
	if !ok {
		return errors.New("CP tidak ditemukan.")
	}
	if !bolehUbah(u.Peran, cp.Status, ISIAcuan) {
		return errTerkunci
	}
	if !adaID(d.Acuan, id, func(a Acuan) int { return a.ID }) {
		return ErrItemTidakDitemukan
	}
	d.Acuan = hapusByID(d.Acuan, id, func(a Acuan) int { return a.ID })
	cp.JumlahAcuan = len(d.Acuan)
	s.hapusAcuanDB(kode, id)
	s.simpanCPState(cp)
	return nil
}

// TambahKriteria menambah kriteria inklusi/eksklusi.
func (s *Store) TambahKriteria(kode string, u *User, jenis, uraian string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp, d, ok := s.detailUntuk(kode)
	if !ok {
		return errors.New("CP tidak ditemukan.")
	}
	if !bolehUbah(u.Peran, cp.Status, ISIKriteria) {
		return errTerkunci
	}
	if jenis != "INKLUSI" && jenis != "EKSKLUSI" {
		return errors.New("Jenis kriteria harus INKLUSI atau EKSKLUSI.")
	}
	uraian = strings.TrimSpace(uraian)
	if uraian == "" {
		return errors.New("Uraian kriteria wajib diisi.")
	}
	k := Kriteria{ID: maxID(d.Kriteria, func(k Kriteria) int { return k.ID }), Jenis: jenis, Uraian: uraian}
	d.Kriteria = append(d.Kriteria, k)
	s.simpanKriteriaDB(kode, k)
	return nil
}

// HapusKriteria menghapus kriteria.
func (s *Store) HapusKriteria(kode string, u *User, id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp, d, ok := s.detailUntuk(kode)
	if !ok {
		return errors.New("CP tidak ditemukan.")
	}
	if !bolehUbah(u.Peran, cp.Status, ISIKriteria) {
		return errTerkunci
	}
	if !adaID(d.Kriteria, id, func(k Kriteria) int { return k.ID }) {
		return ErrItemTidakDitemukan
	}
	d.Kriteria = hapusByID(d.Kriteria, id, func(k Kriteria) int { return k.ID })
	s.hapusKriteriaDB(kode, id)
	return nil
}

// TambahRencana menambah satu baris rencana isi klinis.
func (s *Store) TambahRencana(kode string, u *User, r Rencana) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp, d, ok := s.detailUntuk(kode)
	if !ok {
		return errors.New("CP tidak ditemukan.")
	}
	if !bolehUbah(u.Peran, cp.Status, ISIRencana) {
		return errTerkunci
	}
	r.Nama = strings.TrimSpace(r.Nama)
	if r.Nama == "" {
		return errors.New("Nama rencana wajib diisi.")
	}
	if r.Hari < 1 {
		return errors.New("Hari minimal 1.")
	}
	if r.Severity != "I" && r.Severity != "II" && r.Severity != "III" {
		return errors.New("Severity harus I, II, atau III.")
	}
	if r.Jenis == "" {
		r.Jenis = "OBAT"
	}
	if r.Sifat != "WAJIB" && r.Sifat != "KONDISIONAL" {
		r.Sifat = "WAJIB"
	}
	if r.Dosis == "" {
		r.Dosis = "—"
	}
	if r.Rute == "" {
		r.Rute = "—"
	}
	r.ID = maxID(d.Rencana, func(x Rencana) int { return x.ID })
	d.Rencana = append(d.Rencana, r)
	cp.JumlahRencana = len(d.Rencana)
	s.simpanRencanaDB(kode, r)
	s.simpanCPState(cp)
	return nil
}

// HapusRencana menghapus satu baris rencana.
func (s *Store) HapusRencana(kode string, u *User, id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp, d, ok := s.detailUntuk(kode)
	if !ok {
		return errors.New("CP tidak ditemukan.")
	}
	if !bolehUbah(u.Peran, cp.Status, ISIRencana) {
		return errTerkunci
	}
	if !adaID(d.Rencana, id, func(x Rencana) int { return x.ID }) {
		return ErrItemTidakDitemukan
	}
	d.Rencana = hapusByID(d.Rencana, id, func(x Rencana) int { return x.ID })
	cp.JumlahRencana = len(d.Rencana)
	s.hapusRencanaDB(kode, id)
	s.simpanCPState(cp)
	return nil
}

func adaID[T any](items []T, id int, getID func(T) int) bool {
	for _, it := range items {
		if getID(it) == id {
			return true
		}
	}
	return false
}

func hapusByID[T any](items []T, id int, getID func(T) int) []T {
	out := make([]T, 0, len(items))
	for _, it := range items {
		if getID(it) != id {
			out = append(out, it)
		}
	}
	return out
}
