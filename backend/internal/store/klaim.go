package store

import (
	"math"
	"sort"
)

// Klaim tidak menyimpan identitas pasien. ID adalah identitas episode dari sumber.
type Klaim struct {
	ID           string `json:"id"`
	KodeCP       string `json:"kode_cp"`
	TanggalMasuk string `json:"tanggal_masuk"`
	Severity     string `json:"severity"`
	LOS          int    `json:"los"`
	Tarif        int64  `json:"tarif"`
	ICU          bool   `json:"icu"`
	DiLuarPola   bool   `json:"di_luar_pola"`
}

func (s *Store) loadKlaim() error {
	rows, err := s.db.Query(s.ctx, "SELECT id,kode_cp,tanggal_masuk::text,severity,los,tarif,icu,di_luar_pola FROM klaim_episode ORDER BY tanggal_masuk,id")
	if err != nil {
		return err
	}
	defer rows.Close()
	s.klaim = []Klaim{}
	for rows.Next() {
		var k Klaim
		if err := rows.Scan(&k.ID, &k.KodeCP, &k.TanggalMasuk, &k.Severity, &k.LOS, &k.Tarif, &k.ICU, &k.DiLuarPola); err != nil {
			return err
		}
		s.klaim = append(s.klaim, k)
	}
	return rows.Err()
}
func persentase(n, total int) float64 {
	if total == 0 {
		return 0
	}
	return math.Round(float64(n)*1000/float64(total)) / 10
}
func kuantil(x []int, p float64) int {
	if len(x) == 0 {
		return 0
	}
	v := append([]int{}, x...)
	sort.Ints(v)
	return v[int(math.Ceil(float64(len(v))*p))-1]
}
