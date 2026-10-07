package panduan

import (
	"strings"
	"testing"
)

func TestDaftarSemuaBabValid(t *testing.T) {
	daftar, err := Daftar()
	if err != nil {
		t.Fatalf("Daftar() galat: %v", err)
	}
	if len(daftar) < 10 {
		t.Fatalf("jumlah bab terlalu sedikit: %d", len(daftar))
	}
	ids := map[string]bool{}
	for i, b := range daftar {
		if b.Urutan != i+1 {
			t.Errorf("bab %q urutan %d, harusnya %d", b.ID, b.Urutan, i+1)
		}
		if b.Judul == "" || b.Ringkasan == "" || strings.TrimSpace(b.Isi) == "" {
			t.Errorf("bab %q tidak lengkap (judul/ringkasan/isi kosong)", b.ID)
		}
		if strings.HasPrefix(b.Isi, "# ") {
			t.Errorf("bab %q: judul seharusnya sudah dipisah dari isi", b.ID)
		}
		if ids[b.ID] {
			t.Errorf("id bab ganda: %q", b.ID)
		}
		ids[b.ID] = true
	}
}

func TestUrai(t *testing.T) {
	b, err := urai("isi/05-contoh-bab.md", "# Judul Bab\n\n> Ringkasan baris satu\n> baris dua\n\n## Sub\nIsi.")
	if err != nil {
		t.Fatal(err)
	}
	if b.ID != "contoh-bab" || b.Judul != "Judul Bab" || b.Ringkasan != "Ringkasan baris satu baris dua" {
		t.Fatalf("hasil urai salah: %+v", b)
	}
	if b.Isi != "## Sub\nIsi." {
		t.Fatalf("isi salah: %q", b.Isi)
	}
	if _, err := urai("isi/01-x.md", "tanpa judul"); err == nil {
		t.Fatal("harusnya galat bila baris pertama bukan judul")
	}
}
