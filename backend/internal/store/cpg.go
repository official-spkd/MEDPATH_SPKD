package store

// Data domain Clinical Pathway Generator (CPG). Semua FIKTIF untuk demo.

// CP satu Clinical Pathway dalam library.
type CP struct {
	Kode          string `json:"kode"`
	Nama          string `json:"nama"`
	CMG           string `json:"cmg"`
	Kelompok      string `json:"kelompok_diagnosis"`
	Status        string `json:"status"`
	StatusLabel   string `json:"status_label"`
	StatusSejak   string `json:"status_sejak"`
	JumlahEpisode int    `json:"jumlah_episode"`
	JumlahAcuan   int    `json:"jumlah_acuan"`
	JumlahRencana int    `json:"jumlah_rencana"`
	LosMedian     int    `json:"los_median"`
	MenungguPeran string `json:"menunggu_peran"`
}

// Dokumen panduan klinis (PPK/PNPK/asosiasi).
type Dokumen struct {
	ID       int      `json:"id"`
	Nama     string   `json:"nama"`
	Sumber   string   `json:"sumber"`
	Penerbit string   `json:"penerbit"`
	Tahun    int      `json:"tahun"`
	Cakupan  []string `json:"icd10_cakupan"`
	Butir    int      `json:"butir_count"`
}

var statusLabel = map[string]string{
	"DRAFT":             "Draft",
	"REVISI":            "Perlu Revisi",
	"REVIEW_TIM_CP":     "Review Tim CP",
	"REVIEW_KOMITE":     "Review Komite Medik",
	"MENUNGGU_DIREKTUR": "Menunggu Direktur",
	"AKTIF":             "Aktif",
}

var kelompokCMG = map[string]string{
	"A": "Penyakit Infeksi & Parasit",
	"C": "Neoplasma",
	"D": "Darah & Imunologi",
	"E": "Endokrin & Metabolik",
	"G": "Sistem Saraf",
	"I": "Kardiovaskular",
	"J": "Sistem Pernapasan",
	"K": "Sistem Pencernaan",
	"N": "Ginjal & Saluran Kemih",
	"O": "Kehamilan & Persalinan",
}

// daftarCP data demo Clinical Pathway.
var daftarCP = []CP{
	{"I-4-01", "Infark Miokard Akut (STEMI)", "I", "Kardiovaskular", "AKTIF", "", "2026-09-28", 214, 3, 42, 4, "—"},
	{"I-4-08", "Gagal Jantung Kongestif", "I", "Kardiovaskular", "AKTIF", "", "2026-09-20", 198, 2, 36, 5, "—"},
	{"I-4-09", "Aritmia & Gangguan Konduksi", "I", "Kardiovaskular", "MENUNGGU_DIREKTUR", "", "2026-10-02", 134, 2, 28, 4, "Direktur"},
	{"K-4-06", "Appendektomi", "K", "Sistem Pencernaan", "AKTIF", "", "2026-08-15", 124, 2, 24, 3, "—"},
	{"K-4-08", "Gastritis & Duodenitis", "K", "Sistem Pencernaan", "REVIEW_KOMITE", "", "2026-10-03", 119, 2, 22, 4, "Komite Medik"},
	{"A-4-02", "Demam Tifoid", "A", "Penyakit Infeksi & Parasit", "AKTIF", "", "2026-07-30", 111, 3, 26, 4, "—"},
	{"A-4-06", "Gastroenteritis & Dehidrasi", "A", "Penyakit Infeksi & Parasit", "REVIEW_TIM_CP", "", "2026-10-04", 98, 1, 18, 3, "Tim CP"},
	{"A-4-09", "Dengue / DBD Dewasa", "A", "Penyakit Infeksi & Parasit", "AKTIF", "", "2026-09-05", 92, 2, 20, 4, "—"},
	{"E-4-02", "Diabetes Mellitus Tipe 2", "E", "Endokrin & Metabolik", "AKTIF", "", "2026-08-22", 88, 3, 30, 5, "—"},
	{"E-4-05", "Ketoasidosis Diabetik", "E", "Endokrin & Metabolik", "REVIEW_KOMITE", "", "2026-10-01", 54, 2, 24, 4, "Komite Medik"},
	{"J-4-07", "Pneumonia Komunitas", "J", "Sistem Pernapasan", "AKTIF", "", "2026-08-10", 86, 2, 28, 5, "—"},
	{"J-4-03", "PPOK Eksaserbasi Akut", "J", "Sistem Pernapasan", "MENUNGGU_DIREKTUR", "", "2026-10-02", 72, 2, 22, 4, "Direktur"},
	{"J-4-09", "Asma Bronkial Serangan Akut", "J", "Sistem Pernapasan", "DRAFT", "", "2026-09-30", 48, 0, 0, 0, "Tim CP"},
	{"N-4-05", "Penyakit Ginjal Kronik (HD)", "N", "Ginjal & Saluran Kemih", "AKTIF", "", "2026-07-18", 76, 2, 26, 6, "—"},
	{"N-4-09", "Infeksi Saluran Kemih Komplikata", "N", "Ginjal & Saluran Kemih", "REVIEW_TIM_CP", "", "2026-10-04", 44, 1, 14, 3, "Tim CP"},
	{"G-4-04", "Stroke Iskemik Akut", "G", "Sistem Saraf", "AKTIF", "", "2026-06-25", 102, 3, 34, 5, "—"},
	{"G-4-11", "Epilepsi & Status Epileptikus", "G", "Sistem Saraf", "REVISI", "", "2026-10-03", 38, 1, 16, 4, "Tim CP"},
	{"S-4-05", "Fiksasi Fraktur Ekstremitas", "S", "Muskuloskeletal", "AKTIF", "", "2026-08-01", 66, 2, 20, 3, "—"},
	{"C-4-12", "Kemoterapi Tumor Padat", "C", "Neoplasma", "DRAFT", "", "2026-09-29", 41, 0, 0, 0, "Tim CP"},
	{"D-4-03", "Anemia Memerlukan Transfusi", "D", "Darah & Imunologi", "REVIEW_TIM_CP", "", "2026-10-05", 35, 1, 12, 2, "Tim CP"},
	{"O-6-10", "Seksio Sesarea", "O", "Kehamilan & Persalinan", "AKTIF", "", "2026-07-12", 158, 2, 24, 3, "—"},
	{"O-6-13", "Persalinan Normal", "O", "Kehamilan & Persalinan", "AKTIF", "", "2026-07-12", 203, 1, 18, 2, "—"},
	{"K-4-14", "Sirosis Hepatis Dekompensata", "K", "Sistem Pencernaan", "DRAFT", "", "2026-10-01", 29, 0, 0, 0, "Tim CP"},
	{"I-4-14", "Hipertensi Krisis", "I", "Kardiovaskular", "REVIEW_KOMITE", "", "2026-10-02", 57, 2, 16, 4, "Komite Medik"},
}

var daftarDokumen = []Dokumen{
	{1, "PNPK Tata Laksana Sindrom Koroner Akut", "PNPK", "Kementerian Kesehatan", 2023, []string{"I21.0", "I21.4", "I21.9"}, 14},
	{2, "PNPK Tata Laksana Gagal Jantung", "PNPK", "Kementerian Kesehatan", 2021, []string{"I50.0", "I50.9"}, 11},
	{3, "PPK Neurologi RS — Stroke Iskemik", "PPK_RS", "Komite Medik RS SPKD", 2024, []string{"I63.9", "I63.5"}, 9},
	{4, "PPK Penyakit Dalam RS — Penyakit Ginjal", "PPK_RS", "Komite Medik RS SPKD", 2022, []string{"N18.5", "N17.9"}, 8},
	{5, "PPK Pulmonologi — Pneumonia Komunitas", "PPK_ASOSIASI", "Perhimpunan Profesi", 2023, []string{"J18.9", "J15.9"}, 10},
	{6, "PNPK Infeksi Dengue Dewasa", "PNPK", "Kementerian Kesehatan", 2020, []string{"A91", "A90"}, 7},
	{7, "PNPK Tata Laksana Demam Tifoid", "PNPK", "Kementerian Kesehatan", 2018, []string{"A01.0"}, 6},
	{8, "PPK Penyakit Dalam RS — Sepsis", "PPK_RS", "Komite Medik RS SPKD", 2023, []string{"A41.9"}, 9},
	{9, "PPK Endokrin — Diabetes Mellitus", "PPK_RS", "Komite Medik RS SPKD", 2024, []string{"E11.9", "E14.9"}, 12},
	{10, "PNPK Tata Laksana PPOK", "PNPK", "Kementerian Kesehatan", 2022, []string{"J44.0", "J44.9"}, 8},
}

var butirKategori = map[string]int{
	"KRITERIA_DIAGNOSIS": 28,
	"TATA_LAKSANA":       41,
	"DOSIS_OBAT":         33,
	"INDIKASI_TERAPI":    19,
	"KONTRAINDIKASI":     12,
	"MONITORING":         24,
	"KOMPLIKASI":         15,
	"LAINNYA":            8,
}

// ListCP mengembalikan seluruh CP dari state saat ini (status bisa berubah).
func (s *Store) ListCP() []CP {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]CP, len(s.cp))
	for i, c := range s.cp {
		c.StatusLabel = statusLabel[c.Status]
		out[i] = c
	}
	return out
}

// cpByKode mengembalikan pointer ke CP dalam state (butuh lock dari pemanggil).
func (s *Store) cpByKode(kode string) *CP {
	for i := range s.cp {
		if s.cp[i].Kode == kode {
			return &s.cp[i]
		}
	}
	return nil
}

// ListDokumen mengembalikan dokumen panduan.
func (s *Store) ListDokumen() []Dokumen { return daftarDokumen }

func sedangDiproses(status string) bool {
	return status == "REVIEW_TIM_CP" || status == "REVIEW_KOMITE" || status == "MENUNGGU_DIREKTUR"
}
