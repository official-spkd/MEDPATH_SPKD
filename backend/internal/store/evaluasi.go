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

// Seluruh nilai observasi dihitung dari episode sumber dengan tanggal masuk
// sesudah/tepat tanggal aktif. Tanpa episode, tampilkan keadaan belum ada data.
func (s *Store) hitungEvaluasi(c CP, d *DetailExtra) EvaluasiDetail {
	before, after := []Klaim{}, []Klaim{}
	for _, k := range s.klaim {
		if k.KodeCP != c.Kode {
			continue
		}
		if k.TanggalMasuk >= c.StatusSejak {
			after = append(after, k)
		} else {
			before = append(before, k)
		}
	}
	temuan := []string{}
	mutu := []MutuSeverity{}
	baurBefore, baurAfter := map[string]float64{}, map[string]float64{}
	var total int64
	outside := 0
	for _, k := range after {
		total += k.Tarif
		if k.DiLuarPola {
			outside++
		}
	}
	for _, sv := range d.Severity {
		los := []int{}
		prev := []int{}
		icu, lewat := 0, 0
		for _, k := range before {
			if k.Severity == sv.Severity {
				prev = append(prev, k.LOS)
			}
		}
		target, targetP75 := sv.LosMedian, sv.LosMedian+2
		if len(prev) > 0 {
			target = kuantil(prev, 0.5)
			targetP75 = kuantil(prev, 0.75)
		}
		for _, k := range after {
			if k.Severity == sv.Severity {
				los = append(los, k.LOS)
				if k.ICU {
					icu++
				}
				if k.LOS > targetP75 {
					lewat++
				}
			}
		}
		m := MutuSeverity{sv.Severity, len(los), kuantil(los, 0.5), kuantil(los, 0.75), target, targetP75, persentase(lewat, len(los)), icu}
		mutu = append(mutu, m)
		baurBefore[sv.Severity] = persentase(len(prev), len(before))
		baurAfter[sv.Severity] = persentase(len(los), len(after))
		if len(los) > 0 && m.LosMedian > targetP75 {
			temuan = append(temuan, fmt.Sprintf("Tingkat keparahan %s: median LOS %d hari melebihi p75 target (%d hari).", sv.Severity, m.LosMedian, targetP75))
		}
		if len(los) > 0 && m.PersenLewatP75 >= 30 {
			temuan = append(temuan, fmt.Sprintf("Tingkat keparahan %s: sedikitnya 30%% episode melewati p75 target.", sv.Severity))
		}
	}
	if len(before) > 0 && baurAfter["III"]-baurBefore["III"] >= 8 {
		temuan = append(temuan, "Porsi tingkat keparahan III naik setidaknya 8 poin persentase. Periksa kesesuaian pengkodean.")
	}
	avg := 0
	if len(after) > 0 {
		avg = int(total / int64(len(after)))
	}
	return EvaluasiDetail{EvaluasiRow: EvaluasiRow{c.Kode, c.Nama, c.Kelompok, c.StatusSejak, len(after), len(temuan) > 0, temuan}, KendaliMutu: mutu, KendaliBiaya: KendaliBiaya{baurBefore, baurAfter, avg, total, persentase(outside, len(after))}}
}

// Evaluasi mengembalikan ringkasan + baris untuk seluruh CP aktif.
func (s *Store) Evaluasi(filter ...string) EvaluasiResp {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := []EvaluasiRow{}
	var bisa, perlu, belum int
	for _, c := range filterCP(s.cp, filter) {
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
	sampai := "Belum ada klaim"
	for _, k := range s.klaim {
		if sampai == "Belum ada klaim" || k.TanggalMasuk > sampai {
			sampai = k.TanggalMasuk
		}
	}
	return EvaluasiResp{
		Ringkasan: map[string]any{
			"cp_aktif": len(rows), "bisa_dievaluasi": bisa,
			"perlu_ditinjau": perlu, "belum_aktif": belum, "klaim_sampai": sampai,
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
