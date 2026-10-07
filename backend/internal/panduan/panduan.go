// Package panduan menyajikan Panduan Lengkap Super Admin: dokumentasi teknis MedPath
// yang ditulis sebagai Markdown di isi/*.md lalu di-embed ke binary saat build.
// Mengubah panduan = ubah file .md lalu build ulang backend.
package panduan

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"sync"
)

//go:embed isi/*.md
var berkas embed.FS

// Bagian satu bab panduan.
type Bagian struct {
	ID        string `json:"id"`
	Urutan    int    `json:"urutan"`
	Judul     string `json:"judul"`
	Ringkasan string `json:"ringkasan"`
	Isi       string `json:"isi"` // Markdown tanpa judul & ringkasan
}

var (
	sekali sync.Once
	cache  []Bagian
	galat  error
)

// Daftar mengembalikan seluruh bab, urut sesuai prefiks nama file (01-, 02-, …).
func Daftar() ([]Bagian, error) {
	sekali.Do(func() { cache, galat = muat() })
	return cache, galat
}

func muat() ([]Bagian, error) {
	nama, err := fs.Glob(berkas, "isi/*.md")
	if err != nil {
		return nil, err
	}
	sort.Strings(nama)
	out := make([]Bagian, 0, len(nama))
	for i, n := range nama {
		b, err := berkas.ReadFile(n)
		if err != nil {
			return nil, err
		}
		bag, err := urai(n, string(b))
		if err != nil {
			return nil, err
		}
		bag.Urutan = i + 1
		out = append(out, bag)
	}
	return out, nil
}

// urai memisahkan baris "# Judul" dan paragraf "> ringkasan" pertama dari isi.
func urai(path, teks string) (Bagian, error) {
	base := strings.TrimSuffix(path[strings.LastIndex(path, "/")+1:], ".md")
	if i := strings.Index(base, "-"); i > 0 {
		base = base[i+1:]
	}
	bag := Bagian{ID: base}

	baris := strings.Split(strings.ReplaceAll(teks, "\r\n", "\n"), "\n")
	i := 0
	for i < len(baris) && strings.TrimSpace(baris[i]) == "" {
		i++
	}
	if i >= len(baris) || !strings.HasPrefix(baris[i], "# ") {
		return bag, fmt.Errorf("panduan %s: baris pertama harus judul '# …'", path)
	}
	bag.Judul = strings.TrimSpace(strings.TrimPrefix(baris[i], "# "))
	i++
	for i < len(baris) && strings.TrimSpace(baris[i]) == "" {
		i++
	}
	var ringkas []string
	for i < len(baris) && strings.HasPrefix(baris[i], ">") {
		ringkas = append(ringkas, strings.TrimSpace(strings.TrimPrefix(baris[i], ">")))
		i++
	}
	bag.Ringkasan = strings.Join(ringkas, " ")
	bag.Isi = strings.TrimSpace(strings.Join(baris[i:], "\n"))
	return bag, nil
}
