package store

import (
	"errors"
	"fmt"
	"time"
)

// --- Aturan "siapa boleh apa, di status apa" (port dari AturanCp.php) ---

var peranAdmin = []string{"ADMIN_RS", "SYSTEM_ADMIN", "SUPER_ADMIN"}

func penyusun() []string {
	return append([]string{"TIM_CP", "KOORDINATOR_CP", "DPJP", "DOKTER"}, peranAdmin...)
}
func peranTimCp() []string {
	return append([]string{"TIM_CP", "KOORDINATOR_CP"}, peranAdmin...)
}

type aturanTransisi struct {
	Aksi  string `json:"aksi"`
	Ke    string `json:"ke"`
	Label string `json:"label"`
	Peran []string
}

var labelAksi = map[string]string{
	"AJUKAN":     "Ajukan ke Tim CP",
	"SETUJUI":    "Setujui",
	"SAHKAN":     "Sahkan jadi Aktif",
	"KEMBALIKAN": "Kembalikan untuk revisi",
}

var labelPeran = map[string]string{
	"SUPER_ADMIN": "Super Admin", "SYSTEM_ADMIN": "System Admin", "ADMIN_RS": "Admin RS",
	"KOORDINATOR_CP": "Koordinator CP", "TIM_CP": "Tim CP", "DPJP": "DPJP", "DOKTER": "Dokter",
	"APOTEKER": "Apoteker", "KOMITE_MEDIK": "Komite Medik", "DIREKTUR": "Direktur",
}

func transisiDari(status string) []aturanTransisi {
	komite := append([]string{"KOMITE_MEDIK"}, peranAdmin...)
	direktur := []string{"DIREKTUR", "SYSTEM_ADMIN", "SUPER_ADMIN"}
	switch status {
	case "DRAFT", "REVISI":
		return []aturanTransisi{{Aksi: "AJUKAN", Ke: "REVIEW_TIM_CP", Peran: penyusun()}}
	case "REVIEW_TIM_CP":
		return []aturanTransisi{
			{Aksi: "SETUJUI", Ke: "REVIEW_KOMITE", Peran: peranTimCp()},
			{Aksi: "KEMBALIKAN", Ke: "REVISI", Peran: peranTimCp()},
		}
	case "REVIEW_KOMITE":
		return []aturanTransisi{
			{Aksi: "SETUJUI", Ke: "MENUNGGU_DIREKTUR", Peran: komite},
			{Aksi: "KEMBALIKAN", Ke: "REVISI", Peran: komite},
		}
	case "MENUNGGU_DIREKTUR":
		return []aturanTransisi{
			{Aksi: "SAHKAN", Ke: "AKTIF", Peran: direktur},
			{Aksi: "KEMBALIKAN", Ke: "REVISI", Peran: direktur},
		}
	case "AKTIF":
		return []aturanTransisi{
			{Aksi: "KEMBALIKAN", Ke: "REVISI", Peran: append([]string{"KOMITE_MEDIK", "DIREKTUR"}, peranAdmin...)},
		}
	}
	return nil
}

func berperan(peran string, daftar []string) bool {
	for _, p := range daftar {
		if p == peran {
			return true
		}
	}
	return false
}

// aksiUntuk: aksi yang boleh dilakukan peran ini pada status ini (dengan label).
func aksiUntuk(peran, status string) []aturanTransisi {
	out := []aturanTransisi{}
	for _, a := range transisiDari(status) {
		if berperan(peran, a.Peran) {
			a.Label = labelAksi[a.Aksi]
			a.Peran = nil
			out = append(out, a)
		}
	}
	return out
}

// peranDitunggu: label peran yang ditunggu tindakannya pada status ini.
func peranDitunggu(status string) []string {
	out := []string{}
	if !sedangDiproses(status) {
		return out
	}
	seen := map[string]bool{}
	for _, a := range transisiDari(status) {
		if a.Aksi == "KEMBALIKAN" {
			continue
		}
		for _, p := range a.Peran {
			if berperan(p, peranAdmin) || seen[p] {
				continue
			}
			seen[p] = true
			out = append(out, labelPeran[p])
		}
	}
	return out
}

// --- Siapa boleh mengubah isi CP pada status apa (port dari AturanCp::pengubahIsi) ---

const (
	ISIAcuan    = "ACUAN"
	ISIRencana  = "RENCANA"
	ISIKriteria = "KRITERIA"
)

func pengubahIsi(status, objek string) []string {
	switch status {
	case "DRAFT", "REVISI":
		if objek == ISIRencana {
			return append(penyusun(), "APOTEKER")
		}
		return penyusun()
	case "REVIEW_TIM_CP":
		return peranTimCp()
	// Terkunci: REVIEW_KOMITE, MENUNGGU_DIREKTUR, AKTIF.
	default:
		return []string{}
	}
}

func bolehUbah(peran, status, objek string) bool {
	return berperan(peran, pengubahIsi(status, objek))
}

// BolehUbah menandai objek apa saja yang boleh diubah pengguna ini.
type BolehUbah struct {
	Acuan    bool `json:"acuan"`
	Rencana  bool `json:"rencana"`
	Kriteria bool `json:"kriteria"`
}

// --- 7 Syarat kelengkapan sebelum AKTIF (port dari PemeriksaSyarat.php) ---

type Syarat struct {
	Kode            string   `json:"kode"`
	Judul           string   `json:"judul"`
	Menghambat      bool     `json:"menghambat"`
	Terpenuhi       bool     `json:"terpenuhi"`
	PenanggungJawab string   `json:"penanggung_jawab"`
	Kekurangan      []string `json:"kekurangan"`
}

func syarat(kode, judul string, menghambat bool, pj string, kurang []string) Syarat {
	if kurang == nil {
		kurang = []string{}
	}
	return Syarat{kode, judul, menghambat, len(kurang) == 0, pj, kurang}
}

func periksaSyarat(c CP, d *DetailExtra) []Syarat {
	// ACUAN_CP
	var kurangAcuan []string
	if len(d.Acuan) == 0 {
		kurangAcuan = []string{"Belum ada dokumen panduan yang ditetapkan sebagai acuan CP."}
	}
	// RENCANA_SEVERITY: tiap severity dengan episode cukup (≥10) harus punya rencana
	punyaRencana := map[string]bool{}
	for _, r := range d.Rencana {
		punyaRencana[r.Severity] = true
	}
	var kurangRencana []string
	for _, sv := range d.Severity {
		if sv.Episode >= 10 && !punyaRencana[sv.Severity] {
			kurangRencana = append(kurangRencana, fmt.Sprintf("Severity %s (%d episode) belum punya rencana isi klinis.", sv.Severity, sv.Episode))
		}
	}
	// ACUAN_SUB_CP
	var kurangSub []string
	for _, s := range d.SubCP {
		if !s.AdaAcuan {
			kurangSub = append(kurangSub, fmt.Sprintf("Sub-CP %s belum punya acuan.", s.ICD10))
		}
	}
	// PADANAN_KPTL / TARIF / SNOMED / ICD10 — demo: terpenuhi bila isi klinis sudah ada
	var kurangKptl, kurangTarif, kurangSnomed []string
	if c.JumlahRencana == 0 {
		kurangKptl = []string{"Prosedur belum punya padanan KPTL."}
		kurangTarif = []string{"Tarif INA-CBG tiap severity belum lengkap."}
	}
	if c.JumlahAcuan == 0 {
		kurangSnomed = []string{"Diagnosis belum punya padanan SNOMED-CT yang sah."}
	}

	return []Syarat{
		syarat("ACUAN_CP", "Acuan standar CP — minimal satu dokumen", true, "Tim CP · KSM", kurangAcuan),
		syarat("RENCANA_SEVERITY", "Rencana isi klinis tiap severity dengan data cukup", true, "Tim CP · DPJP · Farmasi", kurangRencana),
		syarat("ACUAN_SUB_CP", "Acuan tiap sub-CP (bila CP perlu dipecah)", false, "Tim CP · KSM terkait", kurangSub),
		syarat("PADANAN_KPTL", "Padanan KPTL tiap prosedur", false, "Tim Koding", kurangKptl),
		syarat("TARIF_SEVERITY", "Tarif INA-CBG tiap severity", false, "Admin · data tarif", kurangTarif),
		syarat("PADANAN_SNOMED", "Padanan SNOMED-CT tiap diagnosis", false, "Tim Koding", kurangSnomed),
		syarat("ICD10_MASTER", "Kode ICD-10 sesuai master", false, "Tim Koding", nil),
	}
}

// --- Detail lengkap untuk halaman CP ---

type CPDetail struct {
	CP
	Severity          []SeverityStat   `json:"severity"`
	SubCP             []SubCP          `json:"sub_cp"`
	Acuan             []Acuan          `json:"acuan"`
	Kriteria          []Kriteria       `json:"kriteria"`
	Rencana           []Rencana        `json:"rencana"`
	Syarat            []Syarat         `json:"syarat"`
	AksiTersedia      []aturanTransisi `json:"aksi_tersedia"`
	MenungguPeranList []string         `json:"menunggu_peran_list"`
	BolehUbah         BolehUbah        `json:"boleh_ubah"`
	JumlahRiwayat     int              `json:"jumlah_riwayat"`
}

// Detail merakit seluruh isi CP untuk pengguna dengan peran tertentu.
func (s *Store) Detail(kode, peran string) (*CPDetail, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := s.cpByKode(kode)
	if cp == nil {
		return nil, false
	}
	c := *cp
	c.StatusLabel = statusLabel[c.Status]
	d := s.detail[kode]
	if d == nil {
		d = &DetailExtra{}
	}
	return &CPDetail{
		CP:                c,
		Severity:          d.Severity,
		SubCP:             d.SubCP,
		Acuan:             d.Acuan,
		Kriteria:          d.Kriteria,
		Rencana:           d.Rencana,
		Syarat:            periksaSyarat(c, d),
		AksiTersedia:      aksiUntuk(peran, c.Status),
		MenungguPeranList: peranDitunggu(c.Status),
		BolehUbah: BolehUbah{
			Acuan:    bolehUbah(peran, c.Status, ISIAcuan),
			Rencana:  bolehUbah(peran, c.Status, ISIRencana),
			Kriteria: bolehUbah(peran, c.Status, ISIKriteria),
		},
		JumlahRiwayat: len(s.riwayat[kode]),
	}, true
}

// Riwayat mengembalikan riwayat pengesahan (terbaru dulu).
func (s *Store) Riwayat(kode string) []RiwayatItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.riwayat[kode]
	out := make([]RiwayatItem, len(src))
	for i, r := range src {
		out[len(src)-1-i] = r
	}
	return out
}

// Transisi menjalankan aksi pengesahan. Mengembalikan error yang ramah bila ditolak.
func (s *Store) Transisi(kode, aksi string, u *User, alasan string) (*CP, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := s.cpByKode(kode)
	if cp == nil {
		return nil, errors.New("CP tidak ditemukan.")
	}
	var cocok *aturanTransisi
	for _, a := range transisiDari(cp.Status) {
		if a.Aksi == aksi {
			ac := a
			cocok = &ac
			break
		}
	}
	if cocok == nil {
		return nil, fmt.Errorf("Aksi %q tidak tersedia pada status %s.", aksi, statusLabel[cp.Status])
	}
	if !berperan(u.Peran, cocok.Peran) {
		return nil, errors.New("Peran Anda tidak berwenang melakukan aksi ini.")
	}
	// Syarat menghambat harus terpenuhi sebelum AKTIF.
	if cocok.Ke == "AKTIF" {
		for _, sy := range periksaSyarat(*cp, s.detail[kode]) {
			if sy.Menghambat && !sy.Terpenuhi {
				return nil, fmt.Errorf("Belum bisa disahkan: %s belum terpenuhi.", sy.Judul)
			}
		}
	}

	dari := cp.Status
	cp.Status = cocok.Ke
	cp.StatusLabel = statusLabel[cp.Status]
	cp.StatusSejak = time.Now().Format("2006-01-02")
	daftar := peranDitunggu(cp.Status)
	if len(daftar) > 0 {
		cp.MenungguPeran = daftar[0]
	} else {
		cp.MenungguPeran = "—"
	}

	rw := RiwayatItem{
		ID: s.nextRiwayat, DariStatus: dari, KeStatus: cp.Status, Aksi: aksi,
		Peran: labelPeran[u.Peran], User: u.Nama, Alasan: alasan,
		Waktu: time.Now().Format("2006-01-02 15:04"),
	}
	s.riwayat[kode] = append(s.riwayat[kode], rw)
	s.nextRiwayat++

	s.simpanCPState(cp)
	s.simpanRiwayat(kode, rw)

	hasil := *cp
	return &hasil, nil
}

// bangunRiwayatAwal menghasilkan jejak pengesahan yang mengarah ke status sekarang.
func bangunRiwayatAwal(c CP) []RiwayatItem {
	// Jalur penuh menuju AKTIF.
	type langkah struct{ dari, ke, aksi, peran string }
	penuh := []langkah{
		{"DRAFT", "REVIEW_TIM_CP", "AJUKAN", "KOORDINATOR_CP"},
		{"REVIEW_TIM_CP", "REVIEW_KOMITE", "SETUJUI", "TIM_CP"},
		{"REVIEW_KOMITE", "MENUNGGU_DIREKTUR", "SETUJUI", "KOMITE_MEDIK"},
		{"MENUNGGU_DIREKTUR", "AKTIF", "SAHKAN", "DIREKTUR"},
	}
	// Berapa langkah yang sudah terjadi sampai status sekarang.
	sampai := map[string]int{
		"DRAFT": 0, "REVISI": 0, "REVIEW_TIM_CP": 1,
		"REVIEW_KOMITE": 2, "MENUNGGU_DIREKTUR": 3, "AKTIF": 4,
	}[c.Status]

	if sampai == 0 {
		return []RiwayatItem{}
	}
	base, _ := time.Parse("2006-01-02", c.StatusSejak)
	if base.IsZero() {
		base = time.Now()
	}
	var out []RiwayatItem
	id := 1
	for i := 0; i < sampai; i++ {
		l := penuh[i]
		t := base.AddDate(0, 0, -(sampai-i)*2)
		out = append(out, RiwayatItem{
			ID: id, DariStatus: l.dari, KeStatus: l.ke, Aksi: l.aksi,
			Peran: labelPeran[l.peran], User: labelPeran[l.peran] + " (demo)",
			Alasan: "", Waktu: t.Format("2006-01-02 15:04"),
		})
		id++
	}
	// CP REVISI: tambahkan langkah pengembalian di akhir.
	if c.Status == "REVISI" {
		out = append(out, RiwayatItem{
			ID: id, DariStatus: "REVIEW_TIM_CP", KeStatus: "REVISI", Aksi: "KEMBALIKAN",
			Peran: labelPeran["TIM_CP"], User: "Tim CP (demo)",
			Alasan: "Rencana klinis severity III belum lengkap.",
			Waktu:  base.Format("2006-01-02 15:04"),
		})
	}
	return out
}
