package store

import "sort"

// Dashboard payload ringkasan untuk halaman utama (gaya eksekutif).
type Dashboard struct {
	Kartu         KartuDashboard `json:"kartu"`
	Readiness     Readiness      `json:"readiness"`
	KemajuanAcuan KemajuanAcuan  `json:"kemajuan_acuan"`
	PerStatus     map[string]int `json:"per_status"`
	PerKelompok   []KelompokBar  `json:"per_kelompok"`
	Prioritas     []Prioritas    `json:"prioritas"`
	ButirKategori map[string]int `json:"butir_per_kategori"`
	Tren          []TitikTren    `json:"tren"`
	Tindakan      []Tindakan     `json:"tindakan"`
}

type KartuDashboard struct {
	JumlahCP        int `json:"jumlah_cp"`
	CpTanpaAcuan    int `json:"cp_tanpa_acuan"`
	DokumenPanduan  int `json:"dokumen_panduan"`
	ButirAcuan      int `json:"butir_acuan"`
	CpAktif         int `json:"cp_aktif"`
	DalamPengesahan int `json:"dalam_pengesahan"`
	TotalEpisode    int `json:"total_episode"`
}

type Readiness struct {
	Skor     int                 `json:"skor"`
	Label    string              `json:"label"`
	Komponen []KomponenReadiness `json:"komponen"`
}

type KomponenReadiness struct {
	Nama  string `json:"nama"`
	Nilai int    `json:"nilai"`
	Maks  int    `json:"maks"`
}

type KemajuanAcuan struct {
	Sudah    int `json:"sudah"`
	Menunggu int `json:"menunggu"`
	Dilewati int `json:"dilewati"`
}

type KelompokBar struct {
	Kode     string `json:"kode"`
	Nama     string `json:"nama"`
	Episode  int    `json:"episode"`
	JumlahCP int    `json:"jumlah_cp"`
}

type Prioritas struct {
	Kode          string `json:"kode"`
	Nama          string `json:"nama"`
	Kelompok      string `json:"kelompok_diagnosis"`
	JumlahEpisode int    `json:"jumlah_episode"`
	Status        string `json:"status"`
	StatusLabel   string `json:"status_label"`
	JumlahAcuan   int    `json:"jumlah_acuan"`
}

type TitikTren struct {
	Periode string `json:"periode"`
	Episode int    `json:"episode"`
	CpAktif int    `json:"cp_aktif"`
}

type Tindakan struct {
	Ikon      string `json:"ikon"`
	Judul     string `json:"judul"`
	Detail    string `json:"detail"`
	Prioritas string `json:"prioritas"` // Tinggi | Sedang | Rendah
}

// Dashboard menyusun ringkasan dari data CP demo.
func (s *Store) Dashboard() Dashboard {
	cps := s.ListCP()

	perStatus := map[string]int{
		"DRAFT": 0, "REVISI": 0, "REVIEW_TIM_CP": 0,
		"REVIEW_KOMITE": 0, "MENUNGGU_DIREKTUR": 0, "AKTIF": 0,
	}
	perKelompokEpisode := map[string]int{}
	perKelompokCP := map[string]int{}
	var tanpaAcuan, aktif, diproses, totalEpisode, sudahAcuan int

	for _, c := range cps {
		perStatus[c.Status]++
		perKelompokEpisode[c.CMG] += c.JumlahEpisode
		perKelompokCP[c.CMG]++
		totalEpisode += c.JumlahEpisode
		if c.JumlahAcuan == 0 {
			tanpaAcuan++
		} else {
			sudahAcuan++
		}
		if c.Status == "AKTIF" {
			aktif++
		}
		if sedangDiproses(c.Status) {
			diproses++
		}
	}

	// Kelompok diagnosis diurutkan dari episode terbanyak.
	var kelompok []KelompokBar
	for kode, ep := range perKelompokEpisode {
		nama := kelompokCMG[kode]
		if nama == "" {
			nama = kode
		}
		kelompok = append(kelompok, KelompokBar{kode, nama, ep, perKelompokCP[kode]})
	}
	sort.Slice(kelompok, func(i, j int) bool { return kelompok[i].Episode > kelompok[j].Episode })

	// Prioritas penyusunan: volume terbanyak.
	sorted := append([]CP(nil), cps...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].JumlahEpisode > sorted[j].JumlahEpisode })
	var prioritas []Prioritas
	for i, c := range sorted {
		if i >= 8 {
			break
		}
		prioritas = append(prioritas, Prioritas{
			Kode: c.Kode, Nama: c.Nama, Kelompok: c.Kelompok,
			JumlahEpisode: c.JumlahEpisode, Status: c.Status,
			StatusLabel: c.StatusLabel, JumlahAcuan: c.JumlahAcuan,
		})
	}

	butirTotal := 0
	for _, n := range butirKategori {
		butirTotal += n
	}

	// Readiness index = kesiapan library CP (acuan, rencana, pengesahan).
	komponen := []KomponenReadiness{
		{"Acuan klinis ditetapkan", sudahAcuan, len(cps)},
		{"Rencana klinis terisi", countRencana(cps), len(cps)},
		{"CP tersahkan aktif", aktif, len(cps)},
		{"Dokumen panduan terhubung", len(daftarDokumen), 14},
		{"Butir acuan terekstrak", butirTotal, 220},
	}
	skor := readinessSkor(komponen)

	return Dashboard{
		Kartu: KartuDashboard{
			JumlahCP:        len(cps),
			CpTanpaAcuan:    tanpaAcuan,
			DokumenPanduan:  len(daftarDokumen),
			ButirAcuan:      butirTotal,
			CpAktif:         aktif,
			DalamPengesahan: diproses,
			TotalEpisode:    totalEpisode,
		},
		Readiness: Readiness{
			Skor:     skor,
			Label:    readinessLabel(skor),
			Komponen: komponen,
		},
		KemajuanAcuan: KemajuanAcuan{Sudah: sudahAcuan, Menunggu: tanpaAcuan, Dilewati: 6},
		PerStatus:     perStatus,
		PerKelompok:   kelompok,
		Prioritas:     prioritas,
		ButirKategori: butirKategori,
		Tren: []TitikTren{
			{"2026-05", 820, 9},
			{"2026-06", 910, 11},
			{"2026-07", 1180, 13},
			{"2026-08", 1340, 14},
			{"2026-09", 1510, 16},
			{"2026-10", totalEpisode, aktif},
		},
		Tindakan: []Tindakan{
			{"stamp", "5 CP menunggu tindakan pengesahan", "Tim CP · Komite Medik · Direktur", "Tinggi"},
			{"book-dashed", "4 CP belum punya acuan klinis", "Perlu ditetapkan Tim CP sebelum diajukan", "Tinggi"},
			{"file-pen", "1 CP dikembalikan untuk revisi", "G-4-11 Epilepsi & Status Epileptikus", "Sedang"},
		},
	}
}

func countRencana(cps []CP) int {
	n := 0
	for _, c := range cps {
		if c.JumlahRencana > 0 {
			n++
		}
	}
	return n
}

func readinessSkor(k []KomponenReadiness) int {
	var total float64
	for _, c := range k {
		if c.Maks > 0 {
			r := float64(c.Nilai) / float64(c.Maks)
			if r > 1 {
				r = 1
			}
			total += r
		}
	}
	return int(total / float64(len(k)) * 100)
}

func readinessLabel(skor int) string {
	switch {
	case skor >= 85:
		return "Library CP sangat siap"
	case skor >= 70:
		return "Library CP cukup siap"
	case skor >= 50:
		return "Library CP perlu dilengkapi"
	default:
		return "Library CP masih awal"
	}
}

// Approval baris antrean pengesahan (CP yang sedang diproses / perlu revisi).
func (s *Store) Approval() []CP {
	var out []CP
	for _, c := range s.ListCP() {
		if sedangDiproses(c.Status) || c.Status == "REVISI" {
			out = append(out, c)
		}
	}
	return out
}

// JumlahAntrean dipakai untuk badge menu Approval.
func (s *Store) JumlahAntrean() int { return len(s.Approval()) }
