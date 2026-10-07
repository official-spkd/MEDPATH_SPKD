package store

import "fmt"

// Isi klinis sebuah CP (digenerasi dari ringkasan untuk demo).
type DetailExtra struct {
	Severity []SeverityStat `json:"severity"`
	SubCP    []SubCP        `json:"sub_cp"`
	Acuan    []Acuan        `json:"acuan"`
	Kriteria []Kriteria     `json:"kriteria"`
	Rencana  []Rencana      `json:"rencana"`
}

type SeverityStat struct {
	Severity  string `json:"severity"` // I | II | III
	Episode   int    `json:"episode"`
	LosMedian int    `json:"los_median"`
}

type SubCP struct {
	ICD10    string `json:"icd10"`
	Diagnosa string `json:"diagnosa"`
	Episode  int    `json:"jumlah_episode"`
	Porsi    string `json:"porsi"`
	AdaAcuan bool   `json:"ada_acuan"`
}

type Acuan struct {
	ID      int    `json:"id"`
	Dokumen string `json:"dokumen"`
	Sumber  string `json:"sumber"`
	Tahun   int    `json:"tahun"`
	Catatan string `json:"catatan"`
}

type Kriteria struct {
	ID     int    `json:"id"`
	Jenis  string `json:"jenis"` // INKLUSI | EKSKLUSI
	Uraian string `json:"uraian"`
}

type Rencana struct {
	ID        int    `json:"id"`
	Jenis     string `json:"jenis"` // OBAT | LAB | PROSEDUR | TINDAKAN
	Severity  string `json:"severity"`
	Hari      int    `json:"hari"`
	Sifat     string `json:"sifat"` // WAJIB | KONDISIONAL
	Kode      string `json:"kode"`
	Nama      string `json:"nama"`
	Dosis     string `json:"dosis"`
	Rute      string `json:"rute"`
	Frekuensi string `json:"frekuensi"`
}

type RiwayatItem struct {
	ID         int    `json:"id"`
	DariStatus string `json:"dari_status"`
	KeStatus   string `json:"ke_status"`
	Aksi       string `json:"aksi"`
	Peran      string `json:"peran"`
	User       string `json:"user"`
	Alasan     string `json:"alasan"`
	Waktu      string `json:"waktu"`
}

// Diagnosa representatif per CP (untuk sub-CP). Fallback generik bila tak ada.
var diagnosaCP = map[string][][2]string{
	"I-4-01": {{"I21.0", "STEMI dinding anterior"}, {"I21.9", "Infark miokard akut"}},
	"I-4-08": {{"I50.0", "Gagal jantung kongestif"}, {"I50.9", "Gagal jantung"}},
	"I-4-09": {{"I49.9", "Aritmia jantung"}, {"I44.2", "Blok AV total"}},
	"K-4-06": {{"K35.80", "Apendisitis akut"}},
	"K-4-08": {{"K29.7", "Gastritis"}, {"K29.9", "Gastroduodenitis"}},
	"A-4-02": {{"A01.0", "Demam tifoid"}},
	"A-4-06": {{"A09", "Gastroenteritis"}},
	"A-4-09": {{"A91", "Demam berdarah dengue"}, {"A90", "Demam dengue"}},
	"E-4-02": {{"E11.9", "Diabetes mellitus tipe 2"}},
	"E-4-05": {{"E10.1", "Ketoasidosis diabetik"}},
	"J-4-07": {{"J18.9", "Pneumonia"}, {"J15.9", "Pneumonia bakterial"}},
	"J-4-03": {{"J44.0", "PPOK eksaserbasi"}, {"J44.9", "PPOK"}},
	"J-4-09": {{"J45.9", "Asma bronkial"}},
	"N-4-05": {{"N18.5", "Penyakit ginjal kronik V"}},
	"N-4-09": {{"N39.0", "Infeksi saluran kemih"}},
	"G-4-04": {{"I63.9", "Stroke iskemik"}, {"I63.5", "Infark serebral"}},
	"G-4-11": {{"G40.9", "Epilepsi"}, {"G41.9", "Status epileptikus"}},
	"S-4-05": {{"S82.2", "Fraktur tibia"}},
	"C-4-12": {{"C50.9", "Neoplasma payudara"}},
	"D-4-03": {{"D64.9", "Anemia"}},
	"O-6-10": {{"O82", "Seksio sesarea"}},
	"O-6-13": {{"O80", "Persalinan normal"}},
	"K-4-14": {{"K74.6", "Sirosis hepatis"}},
	"I-4-14": {{"I16.9", "Krisis hipertensi"}},
}

// Template rencana klinis per jenis (dipakai bergilir).
var templateRencana = []Rencana{
	{Jenis: "LAB", Kode: "LOINC-2160", Nama: "Darah lengkap", Frekuensi: "1x/hari", Rute: "—", Dosis: "—"},
	{Jenis: "OBAT", Kode: "KFA-0101", Nama: "Infus kristaloid NaCl 0,9%", Dosis: "500 mL", Rute: "IV", Frekuensi: "/8 jam"},
	{Jenis: "OBAT", Kode: "KFA-0220", Nama: "Analgetik parasetamol", Dosis: "1 g", Rute: "IV", Frekuensi: "/8 jam"},
	{Jenis: "PROSEDUR", Kode: "ICD9-8703", Nama: "Foto toraks", Dosis: "—", Rute: "—", Frekuensi: "1x"},
	{Jenis: "LAB", Kode: "LOINC-2345", Nama: "Elektrolit serum", Dosis: "—", Rute: "—", Frekuensi: "1x/hari"},
	{Jenis: "OBAT", Kode: "KFA-0455", Nama: "Antibiotik seftriakson", Dosis: "2 g", Rute: "IV", Frekuensi: "/24 jam"},
	{Jenis: "TINDAKAN", Kode: "ICD9-9604", Nama: "Pemasangan kateter IV", Dosis: "—", Rute: "—", Frekuensi: "1x"},
	{Jenis: "OBAT", Kode: "KFA-0612", Nama: "Antiemetik ondansetron", Dosis: "4 mg", Rute: "IV", Frekuensi: "/8 jam"},
}

var severitas = []string{"I", "II", "III"}
var porsiSeverity = []float64{0.5, 0.35, 0.15}

// bangunDetail menghasilkan isi klinis sebuah CP dari ringkasannya.
func bangunDetail(c CP) *DetailExtra {
	d := &DetailExtra{
		Severity: []SeverityStat{},
		SubCP:    []SubCP{},
		Acuan:    []Acuan{},
		Kriteria: []Kriteria{},
		Rencana:  []Rencana{},
	}
	rid := 1

	// Severity stats
	for i, sv := range severitas {
		ep := int(float64(c.JumlahEpisode) * porsiSeverity[i])
		los := c.LosMedian + i // makin berat makin panjang
		d.Severity = append(d.Severity, SeverityStat{Severity: sv, Episode: ep, LosMedian: los})
	}

	// Sub-CP (diagnosis)
	dx := diagnosaCP[c.Kode]
	if len(dx) == 0 {
		dx = [][2]string{{c.CMG + "00.0", c.Nama}}
	}
	for i, dd := range dx {
		ep := c.JumlahEpisode
		if len(dx) > 1 {
			ep = int(float64(c.JumlahEpisode) * (1.0 - float64(i)*0.35))
		}
		d.SubCP = append(d.SubCP, SubCP{
			ICD10: dd[0], Diagnosa: dd[1], Episode: ep,
			Porsi:    fmt.Sprintf("%.0f%%", 100.0/float64(len(dx))),
			AdaAcuan: c.JumlahAcuan > i,
		})
	}

	// Acuan (dari daftar dokumen)
	for i := 0; i < c.JumlahAcuan; i++ {
		dk := daftarDokumen[(i*3+len(c.Kode))%len(daftarDokumen)]
		d.Acuan = append(d.Acuan, Acuan{
			ID: rid, Dokumen: dk.Nama, Sumber: dk.Sumber, Tahun: dk.Tahun,
			Catatan: "Acuan utama tata laksana " + c.Nama,
		})
		rid++
	}

	// Kriteria
	if c.JumlahRencana > 0 {
		d.Kriteria = []Kriteria{
			{ID: 1, Jenis: "INKLUSI", Uraian: "Pasien dewasa terdiagnosis " + c.Nama + " sesuai kriteria klinis & penunjang."},
			{ID: 2, Jenis: "INKLUSI", Uraian: "Dirawat inap di " + c.Kelompok + " dengan indikasi rawat."},
			{ID: 3, Jenis: "EKSKLUSI", Uraian: "Komorbiditas berat yang memerlukan pathway tersendiri."},
		}
	}

	// Rencana klinis tersebar ke severity
	for i := 0; i < c.JumlahRencana; i++ {
		t := templateRencana[i%len(templateRencana)]
		t.ID = rid
		t.Severity = severitas[i%3]
		t.Hari = (i % 4) + 1
		if t.Jenis == "OBAT" || t.Jenis == "LAB" {
			t.Sifat = "WAJIB"
		} else {
			t.Sifat = "KONDISIONAL"
		}
		d.Rencana = append(d.Rencana, t)
		rid++
	}

	return d
}
