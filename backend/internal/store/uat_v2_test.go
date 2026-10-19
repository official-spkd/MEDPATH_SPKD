package store

import (
	"errors"
	"strings"
	"testing"
)

func TestUATV2ValidasiDanSiklus(t *testing.T) {
	s := New()
	tim := s.users[2]
	komite := s.users[5]
	dpjp := s.users[3]
	direktur := s.users[6]
	t.Run("TC-PGS-009", func(t *testing.T) {
		for _, reason := range []string{"", "   "} {
			if _, err := s.Transisi("I-4-14", "KEMBALIKAN", komite, reason); err == nil {
				t.Fatal("alasan kosong diterima")
			}
		}
		if s.cpByKode("I-4-14").Status != "REVIEW_KOMITE" {
			t.Fatal("penolakan mengubah status")
		}
	})
	t.Run("TC-ISI-010-016-API-015", func(t *testing.T) {
		if err := s.TambahKriteria("C-4-12", tim, "INKLUSI", "   "); err == nil {
			t.Fatal("spasi tersimpan")
		}
		for _, day := range []int{-3, 0} {
			if err := s.TambahRencana("C-4-12", tim, Rencana{Nama: "Uji", Severity: "I", Hari: day}); err == nil {
				t.Fatal("hari invalid tersimpan")
			}
		}
		if err := s.HapusAcuan("C-4-12", tim, 999); !errors.Is(err, ErrItemTidakDitemukan) {
			t.Fatal(err)
		}
	})
	t.Run("TC-ISI-005-006", func(t *testing.T) {
		if err := s.TambahAcuan("K-4-14", tim, 1, "Acuan seluruh CP"); err != nil {
			t.Fatal(err)
		}
		if err := s.TambahAcuan("K-4-14", tim, 1, ""); err == nil {
			t.Fatal("duplikat diterima")
		}
		d, _ := s.Detail("K-4-14", tim.Peran)
		if len(d.Acuan) != 1 || !d.SubCP[0].AdaAcuan {
			t.Fatal("acuan sub CP belum terpenuhi")
		}
		for _, sy := range d.Syarat {
			if sy.Kode == "ACUAN_SUB_CP" && !sy.Terpenuhi {
				t.Fatal(sy)
			}
		}
		if err := s.HapusAcuan("K-4-14", tim, d.Acuan[0].ID); err != nil {
			t.Fatal(err)
		}
		d, _ = s.Detail("K-4-14", tim.Peran)
		if d.SubCP[0].AdaAcuan {
			t.Fatal("acuan terhapus masih diwarisi")
		}
	})
	t.Run("TC-E2E-001", func(t *testing.T) {
		if err := s.TambahAcuan("J-4-09", dpjp, 10, "Acuan CP"); err != nil {
			t.Fatal(err)
		}
		for _, sv := range []string{"I", "II"} {
			if err := s.TambahRencana("J-4-09", dpjp, Rencana{Nama: "Uji rencana", Severity: sv, Hari: 1}); err != nil {
				t.Fatal(err)
			}
		}
		if err := s.TambahKriteria("J-4-09", dpjp, "INKLUSI", "Kriteria uji"); err != nil {
			t.Fatal(err)
		}
		d, _ := s.Detail("J-4-09", dpjp.Peran)
		for _, sy := range d.Syarat {
			if !sy.Terpenuhi {
				t.Fatal(sy)
			}
		}
		for _, step := range []struct {
			aksi string
			u    *User
		}{{"AJUKAN", dpjp}, {"SETUJUI", tim}, {"SETUJUI", komite}, {"SAHKAN", direktur}} {
			if _, err := s.Transisi("J-4-09", step.aksi, step.u, ""); err != nil {
				t.Fatal(err)
			}
		}
		if s.cpByKode("J-4-09").Status != "AKTIF" || len(s.Riwayat("J-4-09")) != 4 {
			t.Fatal("siklus belum lengkap")
		}
		k := s.Dashboard().Kartu
		if k.CpAktif != 13 || k.CpTanpaAcuan != 2 {
			t.Fatal(k)
		}
	})
	t.Run("TC-PGS-010-E2E-002", func(t *testing.T) {
		c, err := s.Transisi("J-4-03", "KEMBALIKAN", direktur, "Perbaiki rencana.")
		if err != nil || c.MenungguPeran != "Tim CP" {
			t.Fatal(c, err)
		}
		if _, err := s.Transisi("J-4-03", "AJUKAN", dpjp, ""); err != nil {
			t.Fatal(err)
		}
	})
}

func TestUATV2DashboardDanAntrean(t *testing.T) {
	s := New()
	d := s.Dashboard()
	if d.Kartu.DalamPengesahan != 8 || d.Kartu.CpTanpaAcuan != 3 || !strings.HasPrefix(d.Tindakan[0].Judul, "8 CP") || !strings.HasPrefix(d.Tindakan[1].Judul, "3 CP") {
		t.Fatal(d.Tindakan)
	}
	for i, p := range d.Tren {
		if i > 0 && p.CpAktif < d.Tren[i-1].CpAktif {
			t.Fatal("tren turun tanpa revisi", d.Tren)
		}
	}
	last := d.Tren[len(d.Tren)-1]
	if last.CpAktif != 12 || last.Episode != 2287 {
		t.Fatal(last)
	}
	rw := s.Riwayat("G-4-11")
	if len(rw) != 2 || rw[0].Aksi != "KEMBALIKAN" || rw[0].Alasan == "" {
		t.Fatal(rw)
	}
	for i := range s.cp {
		s.cp[i].Status = "AKTIF"
	}
	if q := s.Approval(); q == nil || len(q) != 0 {
		t.Fatal("antrean kosong harus []")
	}
	if s.Dashboard("RSUD Demo SPKD", "Rawat Jalan").Kartu.JumlahCP != 0 {
		t.Fatal("filter layanan diabaikan")
	}
	if len(s.ListCPFiltered("RSUD Demo SPKD", "Rawat Inap")) != 24 {
		t.Fatal("filter rawat inap")
	}
}

func TestUATV2Padanan(t *testing.T) {
	s := New()
	tim := s.users[2]
	if err := s.SimpanKptl(tim, "00.00", "KPTL-01"); err == nil {
		t.Fatal("prosedur asing diterima")
	}
	if err := s.HapusKptl(tim, "89.52"); err != nil {
		t.Fatal(err)
	}
	d, _ := s.Detail("I-4-01", tim.Peran)
	found := false
	for _, sy := range d.Syarat {
		if sy.Kode == "PADANAN_KPTL" {
			found = !sy.Terpenuhi && strings.Contains(strings.Join(sy.Kekurangan, " "), "89.52")
		}
	}
	if !found {
		t.Fatal("syarat tidak mengikuti padanan")
	}
	for _, x := range s.UsulanSnomed("I63.9") {
		if x.Kode == "22298006" {
			t.Fatal("infark serebral disarankan menjadi miokard")
		}
	}
	found = false
	for _, x := range s.UsulanSnomed("J44.0") {
		if x.Kode == "13645005" {
			found = true
		}
	}
	if !found {
		t.Fatal("singkatan PPOK")
	}
	pairs := map[string]string{"39.95": "KPTL-01", "88.72": "KPTL-02", "87.44": "KPTL-03", "99.04": "KPTL-04", "47.09": "KPTL-05", "89.52": "KPTL-06", "96.04": "KPTL-07", "93.94": "KPTL-08", "38.93": "KPTL-09", "87.03": "KPTL-10", "99.15": "KPTL-11", "45.23": "KPTL-12"}
	for k, v := range pairs {
		if err := s.SimpanKptl(tim, k, v); err != nil {
			t.Fatal(err)
		}
	}
	r := s.DaftarKptl(tim.Peran)
	if r.Ringkasan["sudah"] != 12 || r.Ringkasan["cakupan_episode"] != float64(100) {
		t.Fatal(r.Ringkasan)
	}
}

func TestUATV2EvaluasiTanggalSumber(t *testing.T) {
	s := New()
	c := s.cpByKode("I-4-01")
	c.StatusSejak = "2026-10-01"
	s.klaim = []Klaim{
		{"before", "I-4-01", "2026-09-30", "I", 2, 100, false, false},
		{"on", "I-4-01", "2026-10-01", "I", 4, 200, true, false},
		{"after", "I-4-01", "2026-10-02", "I", 6, 400, false, true},
		{"different", "I-4-08", "2026-10-02", "I", 99, 900, false, false},
	}
	d, err := s.EvaluasiDetailCP(c.Kode)
	if err != nil {
		t.Fatal(err)
	}
	if d.EpisodeSesudah != 2 || d.KendaliBiaya.TotalKlaim != 600 || d.KendaliBiaya.TarifRataRata != 300 || d.KendaliBiaya.PersenDiLuarPola != 50 {
		t.Fatal(d)
	}
	if d.KendaliMutu[0].Episode != 2 || d.KendaliMutu[0].EpisodeIcu != 1 || d.KendaliMutu[0].PersenLewatP75 != 100 {
		t.Fatal(d.KendaliMutu)
	}
	c.StatusSejak = "2026-10-03"
	d, _ = s.EvaluasiDetailCP(c.Kode)
	if d.EpisodeSesudah != 0 || d.KendaliBiaya.TotalKlaim != 0 || d.PerluDitinjau {
		t.Fatal("klaim sebelum aktif ikut dihitung")
	}
}

func TestUATV2AdministrasiDanHash(t *testing.T) {
	s := New()
	if s.users[0].Password == DemoPassword || !strings.HasPrefix(s.users[0].Password, "pbkdf2-sha256$") {
		t.Fatal("kata sandi belum dihash")
	}
	if _, ok := s.Autentikasi(s.users[0].Email, "salah"); ok {
		t.Fatal("kata sandi salah diterima")
	}
	d, err := s.TambahDokumen(Dokumen{Nama: "Dokumen uji", Penerbit: "QA", Sumber: "PNPK", Tahun: 2026, IsiButir: []string{"Satu", "Dua"}})
	if err != nil {
		t.Fatal(err)
	}
	saved, ok := s.DetailDokumen(d.ID)
	if !ok || len(saved.IsiButir) != 2 {
		t.Fatal(saved)
	}
	key, err := s.BuatKunci("QA")
	if err != nil || !s.KunciValid(key) {
		t.Fatal("kunci belum aktif")
	}
	a := s.Administrasi()
	if a.Kunci[0].Hash != "" {
		t.Fatal("hash bocor")
	}
	if err := s.CabutKunci(a.Kunci[0].ID); err != nil || s.KunciValid(key) {
		t.Fatal("kunci dicabut masih berlaku")
	}
}
