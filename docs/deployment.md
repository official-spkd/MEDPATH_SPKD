# MedPath di Vercel dan Supabase

Repository: `official-spkd/MEDPATH_SPKD`. Proyek Vercel: `medpath-spkd`.
Frontend Vue dan backend Go dibangun dari repository yang sama, memakai Vercel Services.
Rute `/api/*` menuju Go tanpa proksi Railway. Supabase hanya menyediakan PostgreSQL;
Supabase Auth/Edge Functions tidak diperlukan untuk arsitektur ini.

## Pengaturan

1. Vercel → Settings → Build and Deployment: Root Directory dikosongkan (root repository).
   Jangan gunakan override build frontend di tingkat proyek; build per service diatur
   dalam `vercel.json`. Region fungsi `sin1` mengikuti database Singapura.
2. Supabase → Connect → Transaction pooler: ambil connection string PostgreSQL port 6543.
   Gunakan TLS, misalnya `sslmode=require`. Password yang mengandung karakter khusus
   harus di-URL-encode. Jangan simpan connection string di GitHub atau kirim di chat.
3. Vercel → Environment Variables: simpan `DATABASE_URL` sebagai sensitive variable
   untuk environment deployment yang sesuai. Jangan beri awalan `VITE_`.
   Preview sebaiknya memakai database uji terpisah agar tidak mengubah state produksi.
4. Deploy `main`, lalu periksa halaman utama, health API, login, dan CP Library.
   Perubahan environment variable baru berlaku pada deployment baru; lakukan Redeploy
   setelah memperbarui `DATABASE_URL`.

`PORT` disediakan Vercel. `DEMO_IN_MEMORY` tidak berlaku ketika `VERCEL` tersedia.
Jika database gagal, aplikasi tidak diam-diam beralih ke penyimpanan sementara.
Tidak perlu menaruh Supabase service-role key atau database password di frontend.

## Database

Saat cold start pertama, backend membuat tabel yang belum ada (tanpa drop tabel),
menambahkan `auth_sessions`, `app_users`, `app_state`, dan `klaim_episode`, dan mengaktifkan RLS untuk tabel aplikasi. Tidak dibuat
policy untuk akses browser langsung; seluruh akses lewat Go dengan role database
server yang memiliki izin tabel. State CP yang sudah ada dipertahankan. Database kosong
akan diisi data demo, di dalam transaksi dan lock agar startup bersamaan tidak menggandakan data.

Pastikan `DATABASE_URL` menunjuk database MedPath yang benar. Jika database sebelumnya
berada di Railway, salin melalui backup/restore terverifikasi sebelum beralih; konfigurasi
ini tidak memindahkan data dari database lain secara otomatis.

## Verifikasi

- `GET /api/v1/kesehatan` → 200 ketika tabel database dapat dibaca.
- Login demo → dashboard, CP Library, dan detail CP dapat dibuka.
- Login di instance A tetap berlaku di instance B; logout dan expiry berlaku di keduanya.
- Edit CP terlihat di instance lain, tanpa kehilangan perubahan bersamaan.
- Galat penyimpanan mengembalikan 503 dan membatalkan status serta audit sekaligus.
- Backend `go test -race ./...` dengan `TEST_DATABASE_URL` database uji.
- Frontend `npm ci && npm run build`.

Hash sandi memakai PBKDF2-HMAC-SHA256 (600.000 iterasi, salt acak). Form login produksi
menyembunyikan bantuan akun demo. Super Admin harus membuat akun non-demo melalui
Pengaturan → Manajemen Pengguna, masuk dengan akun tersebut, lalu memilih Nonaktifkan
Akun Demo. Jangan menonaktifkan demo sebelum akses pengganti terverifikasi.

Evaluasi CP menghitung episode dari `klaim_episode` berdasarkan tanggal aktivasi; database
tanpa episode menampilkan keadaan kosong. Pengisian sumber klaim harus melalui proses
impor tervalidasi, bukan angka buatan. Statistik pustaka awal masih berasal dari data demo.
`ALLOWED_ORIGINS` berisi origin lengkap dipisah koma jika frontend berada di origin lain;
frontend dan API pada domain Vercel yang sama tidak memerlukan CORS lintas origin.
Deployment hosting bukan bukti kesiapan mengelola data klinis nyata.

## Rollback

Deployment lama yang sudah Ready dapat dipulihkan lewat Instant Rollback Vercel.
Jika opsi itu tidak tersedia, gunakan Promote pada deployment lama untuk mengembalikan
domain produksi tanpa membangun ulang kode lama.
Jika akan membangun ulang kode lama, kembalikan Root Directory menjadi `frontend`.
Tidak menghapus layanan Railway saat migrasi; penghapusan/penutupan layanan dilakukan
terpisah setelah koneksi database, API, dan penggunaan aplikasi terverifikasi.

Referensi: [Go runtime](https://vercel.com/docs/functions/runtimes/go),
[Services](https://vercel.com/docs/services),
[Supabase connection pooling](https://supabase.com/docs/guides/database/connecting-to-postgres).
