package store

import (
	"errors"
	"math"
	"sort"
	"strings"
	"time"
)

// Padanan kode (KPTL & SNOMED-CT) — meja kerja Tim Koding. Data demo FIKTIF.

type Prosedur struct {
	ICD9CM    string
	Deskripsi string
	Episode   int
	JumlahCP  int
}

type Diagnosis struct {
	ICD10     string
	Deskripsi string
	Episode   int
	JumlahCP  int
}

type kodeMaster struct {
	Kode      string
	Deskripsi string
	Aktif     bool
}

var prosedurKlaim = []Prosedur{
	{"39.95", "Hemodialisis", 420, 2},
	{"88.72", "Ekokardiografi (USG jantung)", 318, 3},
	{"87.44", "Foto toraks rutin", 286, 5},
	{"99.04", "Transfusi packed red cells", 164, 3},
	{"47.09", "Apendektomi", 124, 1},
	{"89.52", "Elektrokardiografi (EKG)", 410, 6},
	{"96.04", "Pemasangan pipa endotrakeal", 88, 4},
	{"93.94", "Nebulisasi/terapi inhalasi", 150, 2},
	{"38.93", "Kateterisasi vena sentral", 72, 3},
	{"99.15", "Infus parenteral nutrisi", 96, 2},
	{"45.23", "Kolonoskopi", 54, 1},
	{"87.03", "CT scan kepala", 102, 2},
}

var diagnosisKlaim = []Diagnosis{
	{"I21.9", "Infark miokard akut, tidak dirinci", 214, 1},
	{"I50.0", "Gagal jantung kongestif", 198, 1},
	{"E11.9", "Diabetes mellitus tipe 2 tanpa komplikasi", 142, 1},
	{"A01.0", "Demam tifoid", 111, 1},
	{"A91", "Demam berdarah dengue", 92, 1},
	{"J18.9", "Pneumonia, organisme tidak dirinci", 86, 1},
	{"N18.5", "Penyakit ginjal kronik stadium 5", 76, 1},
	{"I63.9", "Infark serebral, tidak dirinci", 102, 1},
	{"K35.80", "Apendisitis akut, tidak dirinci", 124, 1},
	{"J44.0", "PPOK dengan infeksi saluran napas bawah akut", 72, 1},
	{"E14.9", "Diabetes mellitus tidak dirinci", 54, 1},
	{"A41.9", "Sepsis, tidak dirinci", 44, 1},
}

var kptlMaster = []kodeMaster{
	{"KPTL-01", "Hemodialisis rutin", true},
	{"KPTL-02", "Ekokardiografi transtorakal", true},
	{"KPTL-03", "Pemeriksaan radiologi toraks", true},
	{"KPTL-04", "Transfusi komponen darah", true},
	{"KPTL-05", "Apendektomi laparoskopik/terbuka", true},
	{"KPTL-06", "Elektrokardiografi 12 sadapan", true},
	{"KPTL-07", "Intubasi & ventilasi mekanik", true},
	{"KPTL-08", "Terapi inhalasi/nebulisasi", true},
	{"KPTL-09", "Pemasangan kateter vena sentral", true},
	{"KPTL-10", "CT scan kepala tanpa kontras", true},
}

var snomedMaster = []kodeMaster{
	{"22298006", "Infark miokard (kelainan)", true},
	{"42343007", "Gagal jantung kongestif (kelainan)", true},
	{"44054006", "Diabetes mellitus tipe 2 (kelainan)", true},
	{"4834000", "Demam tifoid (kelainan)", true},
	{"1002005", "Demam berdarah dengue (kelainan)", true},
	{"233604007", "Pneumonia (kelainan)", true},
	{"433146000", "Penyakit ginjal kronik stadium 5 (kelainan)", true},
	{"422504002", "Stroke iskemik (kelainan)", true},
	{"74400008", "Apendisitis (kelainan)", true},
	{"13645005", "Penyakit paru obstruktif kronik (kelainan)", true},
	{"999000099", "Konsep lama diabetes (pensiun)", false},
}

type penetapan struct {
	kode string
	oleh string
	pada string
	ragu bool // khusus SNOMED: status RAGU (usulan otomatis belum dikonfirmasi)
}

// Baris meja kerja padanan (KPTL atau SNOMED).
type PadananRow struct {
	Kode           string `json:"kode"` // ICD-9-CM atau ICD-10
	Deskripsi      string `json:"deskripsi"`
	Episode        int    `json:"episode"`
	JumlahCP       int    `json:"jumlah_cp"`
	Padanan        string `json:"padanan"` // kode KPTL/SNOMED, "" bila belum
	PadananDesk    string `json:"padanan_deskripsi"`
	Status         string `json:"status"` // SUDAH | BELUM | RAGU
	DitetapkanOleh string `json:"ditetapkan_oleh"`
	DitetapkanPada string `json:"ditetapkan_pada"`
}

type PadananResp struct {
	Ringkasan map[string]any   `json:"ringkasan"`
	BolehUbah bool             `json:"boleh_ubah"`
	Master    []kodeMasterJSON `json:"master"`
	Data      []PadananRow     `json:"data"`
}

type kodeMasterJSON struct {
	Kode      string `json:"kode"`
	Deskripsi string `json:"deskripsi"`
}

func masterJSON(m []kodeMaster, hanyaAktif bool) []kodeMasterJSON {
	out := []kodeMasterJSON{}
	for _, k := range m {
		if hanyaAktif && !k.Aktif {
			continue
		}
		out = append(out, kodeMasterJSON{k.Kode, k.Deskripsi})
	}
	return out
}

func deskMaster(m []kodeMaster, kode string) string {
	for _, k := range m {
		if k.Kode == kode {
			return k.Deskripsi
		}
	}
	return ""
}

func bolehPetaKode(peran string) bool { return berperan(peran, peranTimCp()) }

// DaftarKptl menyusun meja kerja KPTL.
func (s *Store) DaftarKptl(peran string) PadananResp {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := []PadananRow{}
	var sudah, epTotal, epSudah int
	for _, p := range prosedurKlaim {
		r := PadananRow{Kode: p.ICD9CM, Deskripsi: p.Deskripsi, Episode: p.Episode, JumlahCP: p.JumlahCP, Status: "BELUM"}
		epTotal += p.Episode
		if pt, ok := s.padananKptl[p.ICD9CM]; ok {
			r.Padanan = pt.kode
			r.PadananDesk = deskMaster(kptlMaster, pt.kode)
			r.Status = "SUDAH"
			r.DitetapkanOleh = pt.oleh
			r.DitetapkanPada = pt.pada
			sudah++
			epSudah += p.Episode
		}
		rows = append(rows, r)
	}
	cakupan := 0.0
	if epTotal > 0 {
		cakupan = math.Round(float64(epSudah)/float64(epTotal)*1000) / 10
	}
	return PadananResp{
		Ringkasan: map[string]any{
			"jumlah_prosedur": len(prosedurKlaim), "sudah": sudah,
			"belum": len(prosedurKlaim) - sudah, "cakupan_episode": cakupan,
		},
		BolehUbah: bolehPetaKode(peran),
		Master:    masterJSON(kptlMaster, true),
		Data:      rows,
	}
}

// DaftarSnomed menyusun meja kerja SNOMED-CT.
func (s *Store) DaftarSnomed(peran string) PadananResp {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := []PadananRow{}
	var sudah, ragu int
	for _, dx := range diagnosisKlaim {
		r := PadananRow{Kode: dx.ICD10, Deskripsi: dx.Deskripsi, Episode: dx.Episode, JumlahCP: dx.JumlahCP, Status: "BELUM"}
		if pt, ok := s.padananSnomed[dx.ICD10]; ok {
			r.Padanan = pt.kode
			r.PadananDesk = deskMaster(snomedMaster, pt.kode)
			r.DitetapkanOleh = pt.oleh
			r.DitetapkanPada = pt.pada
			if pt.ragu {
				r.Status = "RAGU"
				ragu++
			} else {
				r.Status = "SUDAH"
				sudah++
			}
		}
		rows = append(rows, r)
	}
	return PadananResp{
		Ringkasan: map[string]any{
			"jumlah_diagnosis": len(diagnosisKlaim), "sudah": sudah, "ragu": ragu,
			"belum": len(diagnosisKlaim) - sudah - ragu,
		},
		BolehUbah: bolehPetaKode(peran),
		Master:    masterJSON(snomedMaster, true),
		Data:      rows,
	}
}

func (s *Store) SimpanKptl(u *User, icd9cm, kptl string) error {
	if !bolehPetaKode(u.Peran) {
		return errors.New("Peran Anda hanya dapat melihat padanan kode.")
	}
	if deskMaster(kptlMaster, kptl) == "" {
		return errors.New("Kode KPTL tidak ada di master.")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p := penetapan{kode: kptl, oleh: u.Nama, pada: time.Now().Format("2006-01-02")}
	s.padananKptl[icd9cm] = p
	s.simpanPadananKptlDB(icd9cm, p)
	return nil
}

func (s *Store) HapusKptl(u *User, icd9cm string) error {
	if !bolehPetaKode(u.Peran) {
		return errors.New("Peran Anda hanya dapat melihat padanan kode.")
	}
	s.mu.Lock()
	delete(s.padananKptl, icd9cm)
	s.hapusPadananKptlDB(icd9cm)
	s.mu.Unlock()
	return nil
}

func (s *Store) SimpanSnomed(u *User, icd10, snomed string) error {
	if !bolehPetaKode(u.Peran) {
		return errors.New("Peran Anda hanya dapat melihat padanan kode.")
	}
	for _, k := range snomedMaster {
		if k.Kode == snomed && !k.Aktif {
			return errors.New("Konsep SNOMED-CT tidak ada atau sudah pensiun.")
		}
	}
	if deskMaster(snomedMaster, snomed) == "" {
		return errors.New("Konsep SNOMED-CT tidak ada di master.")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p := penetapan{kode: snomed, oleh: u.Nama, pada: time.Now().Format("2006-01-02")}
	s.padananSnomed[icd10] = p
	s.simpanPadananSnomedDB(icd10, p)
	return nil
}

func (s *Store) HapusSnomed(u *User, icd10 string) error {
	if !bolehPetaKode(u.Peran) {
		return errors.New("Peran Anda hanya dapat melihat padanan kode.")
	}
	s.mu.Lock()
	delete(s.padananSnomed, icd10)
	s.hapusPadananSnomedDB(icd10)
	s.mu.Unlock()
	return nil
}

// seedPadanan mengisi sebagian padanan awal (sebagian sengaja dibiarkan kosong & satu RAGU).
func (s *Store) seedPadanan() {
	s.padananKptl = map[string]penetapan{
		"39.95": {"KPTL-01", "Tim CP (demo)", "2026-08-10", false},
		"88.72": {"KPTL-02", "Tim CP (demo)", "2026-08-12", false},
		"87.44": {"KPTL-03", "Tim CP (demo)", "2026-08-15", false},
		"89.52": {"KPTL-06", "Tim CP (demo)", "2026-08-15", false},
		"99.04": {"KPTL-04", "Tim CP (demo)", "2026-09-01", false},
	}
	s.padananSnomed = map[string]penetapan{
		"I21.9": {"22298006", "Tim CP (demo)", "2026-08-10", false},
		"I50.0": {"42343007", "Tim CP (demo)", "2026-08-11", false},
		"E11.9": {"44054006", "Tim CP (demo)", "2026-08-12", false},
		"A01.0": {"4834000", "Tim CP (demo)", "2026-08-14", false},
		"J18.9": {"233604007", "Tim CP (demo)", "2026-08-20", true}, // RAGU: usulan otomatis
	}
}

// usulanKptl memberi saran KPTL berdasar kemiripan kata (petunjuk saja).
func usulanKode(nama string, master []kodeMaster) []kodeMasterJSON {
	kata := map[string]bool{}
	for _, w := range strings.FieldsFunc(strings.ToLower(nama), pemisah) {
		if len(w) >= 4 {
			kata[w] = true
		}
	}
	type skor struct {
		m kodeMasterJSON
		s float64
	}
	var hasil []skor
	for _, k := range master {
		if !k.Aktif {
			continue
		}
		milik := map[string]bool{}
		for _, w := range strings.FieldsFunc(strings.ToLower(k.Deskripsi), pemisah) {
			if len(w) >= 4 {
				milik[w] = true
			}
		}
		sama, gab := 0, map[string]bool{}
		for w := range kata {
			gab[w] = true
			if milik[w] {
				sama++
			}
		}
		for w := range milik {
			gab[w] = true
		}
		if sama > 0 {
			hasil = append(hasil, skor{kodeMasterJSON{k.Kode, k.Deskripsi}, float64(sama) / float64(len(gab))})
		}
	}
	sort.Slice(hasil, func(i, j int) bool { return hasil[i].s > hasil[j].s })
	out := []kodeMasterJSON{}
	for i, h := range hasil {
		if i >= 5 {
			break
		}
		out = append(out, h.m)
	}
	return out
}

func pemisah(r rune) bool {
	return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
}

// UsulanKptl memberi saran KPTL untuk sebuah prosedur ICD-9-CM.
func (s *Store) UsulanKptl(icd9cm string) []kodeMasterJSON {
	for _, p := range prosedurKlaim {
		if p.ICD9CM == icd9cm {
			return usulanKode(p.Deskripsi, kptlMaster)
		}
	}
	return []kodeMasterJSON{}
}

// UsulanSnomed memberi saran konsep SNOMED untuk sebuah diagnosis ICD-10.
func (s *Store) UsulanSnomed(icd10 string) []kodeMasterJSON {
	for _, dx := range diagnosisKlaim {
		if dx.ICD10 == icd10 {
			return usulanKode(dx.Deskripsi, snomedMaster)
		}
	}
	return []kodeMasterJSON{}
}
