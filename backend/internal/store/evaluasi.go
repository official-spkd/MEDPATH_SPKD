package store

import "fmt"

// Evaluasi CP aktif: membandingkan klaim SESUDAH CP aktif dengan target (statistik saat disahkan).
// Biaya riil RS tidak dimuat — sisi biaya hanya tarif INA-CBG. Data demo FIKTIF.

type EvaluasiRow struct {
	Kode           string   `json:"kode"`
	Nama           string   `json:"nama"`
	Kelompok       string   `json:"kelompok_diagnosis"`
	AktifSejak     string   `json:"aktif_sejak"`
	EpisodeSesudah int      `json:"episode_sesudah"`
	PerluDitinjau  bool     `json:"perlu_ditinjau"`
	Temuan         []string `json:"temuan"`
}

type MutuSeverity struct {
	Severity       string  `json:"severity"`
	Episode        int     `json:"episode"`
	LosMedian      int     `json:"los_median"`
	LosP75         int     `json:"los_p75"`
	TargetLos      int     `json:"target_los"`
	TargetP75      int     `json:"target_p75"`
	PersenLewatP75 float64 `json:"persen_lewat_p75"`
	EpisodeIcu     int     `json:"episode_icu"`
}

type EvaluasiDetail struct {
	EvaluasiRow
	KendaliMutu  []MutuSeverity `json:"kendali_mutu"`
	KendaliBiaya KendaliBiaya   `json:"kendali_biaya"`
}

type KendaliBiaya struct {
	BaurSeveritySebelum map[string]float64 `json:"baur_severity_sebelum"`
	BaurSeveritySesudah map[string]float64 `json:"baur_severity_sesudah"`
	TarifRataRata       int                `json:"tarif_rata_rata"`
	TotalKlaim          int64              `json:"total_klaim"`
	PersenDiLuarPola    float64            `json:"persen_diagnosis_di_luar_pola"`
}

type EvaluasiResp struct {
	Ringkasan map[string]any `json:"ringkasan"`
	Data      []EvaluasiRow  `json:"data"`
}

// hash sederhana & deterministik dari kode.
func hashKode(kode string) int {
	h := 0
	for _, c := range kode {
		h = h*31 + int(c)
	}
	if h < 0 {
		h = -h
	}
	return h
}

func (s *Store) hitungEvaluasi(c CP, d *DetailExtra) EvaluasiDetail {
	h := hashKode(c.Kode)
	epSesudah := int(float64(c.JumlahEpisode) * 0.42)
	var temuan []string
	mutu := []MutuSeverity{}
	tarifBase := 4_000_000 + (h%40)*100_000

	for i, sv := range d.Severity {
		targetMedian := sv.LosMedian
		targetP75 := targetMedian + 2
		delta := (h / (i + 1)) % 4 // 0..3
		median := targetMedian + delta - 1
		if median < 1 {
			median = 1
		}
		p75 := median + 2
		ep := int(float64(epSesudah) * porsiSeverity[i])
		persenLewat := float64((h + i*7) % 40) // 0..39%
		m := MutuSeverity{
			Severity: sv.Severity, Episode: ep, LosMedian: median, LosP75: p75,
			TargetLos: targetMedian, TargetP75: targetP75,
			PersenLewatP75: persenLewat, EpisodeIcu: ep / 8,
		}
		mutu = append(mutu, m)
		if ep > 0 && median > targetP75 {
			temuan = append(temuan, fmt.Sprintf("Severity %s: median LOS %d hari melebihi p75 target (%d hari).", sv.Severity, median, targetP75))
		}
		if ep > 0 && persenLewat >= 30 {
			temuan = append(temuan, fmt.Sprintf("Severity %s: %.0f%% episode lama rawatnya melewati p75 target.", sv.Severity, persenLewat))
		}
	}

	baurSebelum := map[string]float64{"I": 50, "II": 35, "III": 15}
	naik := float64((h % 12))
	baurSesudah := map[string]float64{"I": 50 - naik/2, "II": 35 - naik/2, "III": 15 + naik}
	if naik >= 8 {
		temuan = append(temuan, fmt.Sprintf("Porsi severity III naik 15%% → %.0f%%. Periksa kesesuaian pengkodean.", baurSesudah["III"]))
	}

	return EvaluasiDetail{
		EvaluasiRow: EvaluasiRow{
			Kode: c.Kode, Nama: c.Nama, Kelompok: c.Kelompok, AktifSejak: c.StatusSejak,
			EpisodeSesudah: epSesudah, PerluDitinjau: len(temuan) > 0, Temuan: nz(temuan),
		},
		KendaliMutu: mutu,
		KendaliBiaya: KendaliBiaya{
			BaurSeveritySebelum: baurSebelum,
			BaurSeveritySesudah: baurSesudah,
			TarifRataRata:       tarifBase,
			TotalKlaim:          int64(tarifBase) * int64(epSesudah),
			PersenDiLuarPola:    float64(h % 12),
		},
	}
}

// Evaluasi mengembalikan ringkasan + baris untuk seluruh CP aktif.
func (s *Store) Evaluasi() EvaluasiResp {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := []EvaluasiRow{}
	var bisa, perlu, belum int
	for i := range s.cp {
		c := s.cp[i]
		if c.Status != "AKTIF" {
			belum++
			continue
		}
		det := s.hitungEvaluasi(c, s.detail[c.Kode])
		if det.EpisodeSesudah > 0 {
			bisa++
		}
		if det.PerluDitinjau {
			perlu++
		}
		rows = append(rows, det.EvaluasiRow)
	}
	return EvaluasiResp{
		Ringkasan: map[string]any{
			"cp_aktif": len(rows), "bisa_dievaluasi": bisa,
			"perlu_ditinjau": perlu, "belum_aktif": belum, "klaim_sampai": "2026-10-05",
		},
		Data: rows,
	}
}

// EvaluasiDetailCP mengembalikan evaluasi rinci satu CP aktif.
func (s *Store) EvaluasiDetailCP(kode string) (*EvaluasiDetail, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := s.cpByKode(kode)
	if cp == nil {
		return nil, fmt.Errorf("CP tidak ditemukan.")
	}
	if cp.Status != "AKTIF" {
		return nil, fmt.Errorf("Evaluasi hanya tersedia untuk CP yang sedang aktif.")
	}
	det := s.hitungEvaluasi(*cp, s.detail[kode])
	return &det, nil
}

func nz(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
