package store

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Persistensi PostgreSQL untuk STATE yang berubah saat runtime (status CP, isi klinis,
// riwayat, padanan). Data statis (severity, sub-CP, dokumen, master) tetap digenerasi di Go.
// Bila pool nil, store berjalan murni di memori.

const skema = `
CREATE TABLE IF NOT EXISTS cp_state (
    kode TEXT PRIMARY KEY,
    status TEXT NOT NULL,
    status_sejak TEXT,
    menunggu_peran TEXT,
    jumlah_acuan INT NOT NULL DEFAULT 0,
    jumlah_rencana INT NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS cp_acuan (
    kode TEXT NOT NULL, id INT NOT NULL,
    dokumen TEXT, sumber TEXT, tahun INT, catatan TEXT,
    PRIMARY KEY (kode, id)
);
CREATE TABLE IF NOT EXISTS cp_kriteria (
    kode TEXT NOT NULL, id INT NOT NULL,
    jenis TEXT, uraian TEXT,
    PRIMARY KEY (kode, id)
);
CREATE TABLE IF NOT EXISTS cp_rencana (
    kode TEXT NOT NULL, id INT NOT NULL,
    jenis TEXT, severity TEXT, hari INT, sifat TEXT,
    kode_item TEXT, nama TEXT, dosis TEXT, rute TEXT, frekuensi TEXT,
    PRIMARY KEY (kode, id)
);
CREATE TABLE IF NOT EXISTS cp_riwayat (
    kode TEXT NOT NULL, id INT NOT NULL,
    dari_status TEXT, ke_status TEXT, aksi TEXT,
    peran TEXT, nama_user TEXT, alasan TEXT, waktu TEXT,
    PRIMARY KEY (kode, id)
);
CREATE TABLE IF NOT EXISTS padanan_kptl (
    icd9cm TEXT PRIMARY KEY, kptl TEXT, oleh TEXT, pada TEXT
);
CREATE TABLE IF NOT EXISTS padanan_snomed (
    icd10 TEXT PRIMARY KEY, snomed TEXT, oleh TEXT, pada TEXT, ragu BOOLEAN
);
`

var ctx = context.Background()

// pasangDB menyambungkan store ke PostgreSQL: migrasi, lalu seed (bila kosong) atau load.
func (s *Store) pasangDB(pool *pgxpool.Pool) error {
	s.pool = pool
	if _, err := pool.Exec(ctx, skema); err != nil {
		return err
	}
	var jml int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM cp_state").Scan(&jml); err != nil {
		return err
	}
	if jml == 0 {
		log.Println("DB kosong — menanam data demo awal ke PostgreSQL…")
		return s.seedDB()
	}
	log.Printf("Memuat state dari PostgreSQL (%d CP)…", jml)
	return s.loadDB()
}

// seedDB menulis state awal (dari memori) ke DB.
func (s *Store) seedDB() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b := &pgxBatchHelper{pool: s.pool}
	for i := range s.cp {
		c := s.cp[i]
		b.exec(`INSERT INTO cp_state(kode,status,status_sejak,menunggu_peran,jumlah_acuan,jumlah_rencana)
		        VALUES($1,$2,$3,$4,$5,$6)`, c.Kode, c.Status, c.StatusSejak, c.MenungguPeran, c.JumlahAcuan, c.JumlahRencana)
		d := s.detail[c.Kode]
		if d != nil {
			for _, a := range d.Acuan {
				b.exec(`INSERT INTO cp_acuan(kode,id,dokumen,sumber,tahun,catatan) VALUES($1,$2,$3,$4,$5,$6)`,
					c.Kode, a.ID, a.Dokumen, a.Sumber, a.Tahun, a.Catatan)
			}
			for _, k := range d.Kriteria {
				b.exec(`INSERT INTO cp_kriteria(kode,id,jenis,uraian) VALUES($1,$2,$3,$4)`, c.Kode, k.ID, k.Jenis, k.Uraian)
			}
			for _, r := range d.Rencana {
				b.exec(`INSERT INTO cp_rencana(kode,id,jenis,severity,hari,sifat,kode_item,nama,dosis,rute,frekuensi)
				        VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
					c.Kode, r.ID, r.Jenis, r.Severity, r.Hari, r.Sifat, r.Kode, r.Nama, r.Dosis, r.Rute, r.Frekuensi)
			}
		}
		for _, rw := range s.riwayat[c.Kode] {
			b.exec(`INSERT INTO cp_riwayat(id,kode,dari_status,ke_status,aksi,peran,nama_user,alasan,waktu)
			        VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
				rw.ID, c.Kode, rw.DariStatus, rw.KeStatus, rw.Aksi, rw.Peran, rw.User, rw.Alasan, rw.Waktu)
		}
	}
	for icd, p := range s.padananKptl {
		b.exec(`INSERT INTO padanan_kptl(icd9cm,kptl,oleh,pada) VALUES($1,$2,$3,$4)`, icd, p.kode, p.oleh, p.pada)
	}
	for icd, p := range s.padananSnomed {
		b.exec(`INSERT INTO padanan_snomed(icd10,snomed,oleh,pada,ragu) VALUES($1,$2,$3,$4,$5)`, icd, p.kode, p.oleh, p.pada, p.ragu)
	}
	return b.flush()
}

// loadDB memuat state dari DB ke memori (menimpa bagian yang mutable).
func (s *Store) loadDB() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// cp_state → status
	rows, err := s.pool.Query(ctx, "SELECT kode,status,status_sejak,menunggu_peran,jumlah_acuan,jumlah_rencana FROM cp_state")
	if err != nil {
		return err
	}
	for rows.Next() {
		var kode, status, sejak, menunggu string
		var ja, jr int
		if err := rows.Scan(&kode, &status, &sejak, &menunggu, &ja, &jr); err != nil {
			rows.Close()
			return err
		}
		if cp := s.cpByKode(kode); cp != nil {
			cp.Status, cp.StatusSejak, cp.MenungguPeran = status, sejak, menunggu
			cp.StatusLabel = statusLabel[status]
			cp.JumlahAcuan, cp.JumlahRencana = ja, jr
		}
	}
	rows.Close()

	// Kosongkan isi mutable lalu isi dari DB (severity & sub_cp tetap dari memori).
	for _, d := range s.detail {
		d.Acuan = []Acuan{}
		d.Kriteria = []Kriteria{}
		d.Rencana = []Rencana{}
	}
	ar, _ := s.pool.Query(ctx, "SELECT kode,id,dokumen,sumber,tahun,catatan FROM cp_acuan ORDER BY kode,id")
	for ar.Next() {
		var kode string
		var a Acuan
		ar.Scan(&kode, &a.ID, &a.Dokumen, &a.Sumber, &a.Tahun, &a.Catatan)
		if d := s.detail[kode]; d != nil {
			d.Acuan = append(d.Acuan, a)
		}
	}
	ar.Close()
	kr, _ := s.pool.Query(ctx, "SELECT kode,id,jenis,uraian FROM cp_kriteria ORDER BY kode,id")
	for kr.Next() {
		var kode string
		var k Kriteria
		kr.Scan(&kode, &k.ID, &k.Jenis, &k.Uraian)
		if d := s.detail[kode]; d != nil {
			d.Kriteria = append(d.Kriteria, k)
		}
	}
	kr.Close()
	rr, _ := s.pool.Query(ctx, "SELECT kode,id,jenis,severity,hari,sifat,kode_item,nama,dosis,rute,frekuensi FROM cp_rencana ORDER BY kode,id")
	for rr.Next() {
		var kode string
		var r Rencana
		rr.Scan(&kode, &r.ID, &r.Jenis, &r.Severity, &r.Hari, &r.Sifat, &r.Kode, &r.Nama, &r.Dosis, &r.Rute, &r.Frekuensi)
		if d := s.detail[kode]; d != nil {
			d.Rencana = append(d.Rencana, r)
		}
	}
	rr.Close()

	// riwayat
	for k := range s.riwayat {
		s.riwayat[k] = []RiwayatItem{}
	}
	wr, _ := s.pool.Query(ctx, "SELECT id,kode,dari_status,ke_status,aksi,peran,nama_user,alasan,waktu FROM cp_riwayat ORDER BY kode,id")
	for wr.Next() {
		var kode string
		var it RiwayatItem
		wr.Scan(&it.ID, &kode, &it.DariStatus, &it.KeStatus, &it.Aksi, &it.Peran, &it.User, &it.Alasan, &it.Waktu)
		s.riwayat[kode] = append(s.riwayat[kode], it)
		if it.ID >= s.nextRiwayat {
			s.nextRiwayat = it.ID + 1
		}
	}
	wr.Close()

	// padanan
	s.padananKptl = map[string]penetapan{}
	pk, _ := s.pool.Query(ctx, "SELECT icd9cm,kptl,oleh,pada FROM padanan_kptl")
	for pk.Next() {
		var icd string
		var p penetapan
		pk.Scan(&icd, &p.kode, &p.oleh, &p.pada)
		s.padananKptl[icd] = p
	}
	pk.Close()
	s.padananSnomed = map[string]penetapan{}
	ps, _ := s.pool.Query(ctx, "SELECT icd10,snomed,oleh,pada,ragu FROM padanan_snomed")
	for ps.Next() {
		var icd string
		var p penetapan
		ps.Scan(&icd, &p.kode, &p.oleh, &p.pada, &p.ragu)
		s.padananSnomed[icd] = p
	}
	ps.Close()
	return nil
}

// --- Write-through (dipanggil dari dalam lock store; no-op bila pool nil) ---

func (s *Store) simpanCPState(c *CP) {
	if s.pool == nil {
		return
	}
	s.pool.Exec(ctx, `UPDATE cp_state SET status=$2,status_sejak=$3,menunggu_peran=$4,jumlah_acuan=$5,jumlah_rencana=$6 WHERE kode=$1`,
		c.Kode, c.Status, c.StatusSejak, c.MenungguPeran, c.JumlahAcuan, c.JumlahRencana)
}

func (s *Store) simpanRiwayat(kode string, r RiwayatItem) {
	if s.pool == nil {
		return
	}
	s.pool.Exec(ctx, `INSERT INTO cp_riwayat(id,kode,dari_status,ke_status,aksi,peran,nama_user,alasan,waktu)
	        VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		r.ID, kode, r.DariStatus, r.KeStatus, r.Aksi, r.Peran, r.User, r.Alasan, r.Waktu)
}

func (s *Store) simpanAcuanDB(kode string, a Acuan) {
	if s.pool == nil {
		return
	}
	s.pool.Exec(ctx, `INSERT INTO cp_acuan(kode,id,dokumen,sumber,tahun,catatan) VALUES($1,$2,$3,$4,$5,$6)`,
		kode, a.ID, a.Dokumen, a.Sumber, a.Tahun, a.Catatan)
}

func (s *Store) hapusAcuanDB(kode string, id int) {
	if s.pool == nil {
		return
	}
	s.pool.Exec(ctx, `DELETE FROM cp_acuan WHERE kode=$1 AND id=$2`, kode, id)
}

func (s *Store) simpanKriteriaDB(kode string, k Kriteria) {
	if s.pool == nil {
		return
	}
	s.pool.Exec(ctx, `INSERT INTO cp_kriteria(kode,id,jenis,uraian) VALUES($1,$2,$3,$4)`, kode, k.ID, k.Jenis, k.Uraian)
}

func (s *Store) hapusKriteriaDB(kode string, id int) {
	if s.pool == nil {
		return
	}
	s.pool.Exec(ctx, `DELETE FROM cp_kriteria WHERE kode=$1 AND id=$2`, kode, id)
}

func (s *Store) simpanRencanaDB(kode string, r Rencana) {
	if s.pool == nil {
		return
	}
	s.pool.Exec(ctx, `INSERT INTO cp_rencana(kode,id,jenis,severity,hari,sifat,kode_item,nama,dosis,rute,frekuensi)
	        VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		kode, r.ID, r.Jenis, r.Severity, r.Hari, r.Sifat, r.Kode, r.Nama, r.Dosis, r.Rute, r.Frekuensi)
}

func (s *Store) hapusRencanaDB(kode string, id int) {
	if s.pool == nil {
		return
	}
	s.pool.Exec(ctx, `DELETE FROM cp_rencana WHERE kode=$1 AND id=$2`, kode, id)
}

func (s *Store) simpanPadananKptlDB(icd9cm string, p penetapan) {
	if s.pool == nil {
		return
	}
	s.pool.Exec(ctx, `INSERT INTO padanan_kptl(icd9cm,kptl,oleh,pada) VALUES($1,$2,$3,$4)
	        ON CONFLICT (icd9cm) DO UPDATE SET kptl=$2,oleh=$3,pada=$4`, icd9cm, p.kode, p.oleh, p.pada)
}

func (s *Store) hapusPadananKptlDB(icd9cm string) {
	if s.pool == nil {
		return
	}
	s.pool.Exec(ctx, `DELETE FROM padanan_kptl WHERE icd9cm=$1`, icd9cm)
}

func (s *Store) simpanPadananSnomedDB(icd10 string, p penetapan) {
	if s.pool == nil {
		return
	}
	s.pool.Exec(ctx, `INSERT INTO padanan_snomed(icd10,snomed,oleh,pada,ragu) VALUES($1,$2,$3,$4,$5)
	        ON CONFLICT (icd10) DO UPDATE SET snomed=$2,oleh=$3,pada=$4,ragu=$5`, icd10, p.kode, p.oleh, p.pada, p.ragu)
}

func (s *Store) hapusPadananSnomedDB(icd10 string) {
	if s.pool == nil {
		return
	}
	s.pool.Exec(ctx, `DELETE FROM padanan_snomed WHERE icd10=$1`, icd10)
}

// pgxBatchHelper menjalankan banyak INSERT sekaligus.
type pgxBatchHelper struct {
	pool  *pgxpool.Pool
	calls []func() error
	err   error
}

func (b *pgxBatchHelper) exec(sql string, args ...any) {
	_, err := b.pool.Exec(ctx, sql, args...)
	if err != nil && b.err == nil {
		b.err = err
	}
}

func (b *pgxBatchHelper) flush() error { return b.err }
