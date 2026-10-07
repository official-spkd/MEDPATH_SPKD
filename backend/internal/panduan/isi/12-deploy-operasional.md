# Deploy & Operasional

> Cara membangun artefak produksi, menyusun server (reverse proxy + Go + PostgreSQL), memantau kesehatan, dan merawat sistem sehari-hari.

## Artefak build

| Bagian | Perintah | Hasil |
|---|---|---|
| Backend | `cd backend && CGO_ENABLED=0 go build -o transcpg-api .` | satu binary statis (panduan sudah ter-embed) |
| Backend untuk Linux dari Mac | `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o transcpg-api .` | binary Linux amd64 |
| Frontend | `cd frontend && npm ci && npm run build` | folder `frontend/dist/` (HTML + aset ber-hash) |

Binary Go tidak butuh runtime apa pun di server. Frontend hanyalah file statis.

## Topologi yang disarankan

```text
Internet ──HTTPS──▶ Caddy / Nginx  (TLS, satu domain)
                      ├── /          → file statis frontend/dist/
                      └── /api/*     → http://127.0.0.1:8080  (transcpg-api)
                                              │
                                              ▼
                                        PostgreSQL 17 (localhost / jaringan privat)
```

Dengan satu domain, frontend memanggil `/api/v1` secara relatif — tidak perlu CORS terbuka, dan `Access-Control-Allow-Origin: *` bisa dipersempit.

### Contoh Caddyfile

```text
medpath.contoh-rs.id {
    encode gzip
    handle /api/* {
        reverse_proxy 127.0.0.1:8080
    }
    handle {
        root * /srv/medpath/dist
        try_files {path} /index.html
        file_server
    }
}
```

Karena router memakai hash (`/#/…`), `try_files` sebenarnya opsional — tetapi aman untuk dipasang.

### Contoh unit systemd

```ini
[Unit]
Description=MedPath API
After=network.target postgresql.service

[Service]
ExecStart=/srv/medpath/transcpg-api
Environment=ALAMAT=127.0.0.1:8080
Environment=DATABASE_URL=postgres://medpath:GANTI_SANDI@127.0.0.1:5432/transcpg?sslmode=disable
Restart=always
User=medpath

[Install]
WantedBy=multi-user.target
```

Ikat backend ke `127.0.0.1` agar hanya bisa diakses lewat reverse proxy.

## Menyiapkan PostgreSQL produksi

```sql
CREATE ROLE medpath LOGIN PASSWORD 'GANTI_SANDI';
CREATE DATABASE transcpg OWNER medpath;
```

Saat pertama kali backend jalan, tabel dibuat dan data demo ditanam. **Untuk produksi sungguhan, nonaktifkan seed demo** (data CP demo, akun & sandi demo) — lihat daftar periksa di bab *Keamanan & RBAC*.

## Pemantauan

- **Health check:** `GET /api/v1/kesehatan` → `{"status":"ok","layanan":"transcpg-api"}` — pakai untuk uptime monitor / load balancer.
- **Log:** backend menulis ke stdout (`journalctl -u medpath -f` bila memakai systemd). Baris penting saat start: `Terhubung ke PostgreSQL.` dan `Memuat state dari PostgreSQL (N CP)…`. Bila muncul `memakai data in-memory`, **perubahan tidak tersimpan** — tangani segera.
- **Database:** pantau ukuran `cp_riwayat` (tumbuh terus, append-only).

## Pekerjaan rutin

| Frekuensi | Tugas |
|---|---|
| Harian | backup `pg_dump -Fc`, simpan di luar server |
| Mingguan | cek log galat & ruang disk |
| Bulanan | uji restore backup ke database uji |
| Saat rilis | `go test ./...`, `npm run build`, tag versi git |
| Saat token/akses tak dipakai | cabut Personal Access Token GitHub & kunci integrasi |

## Rilis versi baru

```bash
git pull origin main
cd backend && go test ./... && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o transcpg-api .
cd ../frontend && npm ci && npm run build
# salin transcpg-api & dist/ ke server, lalu:
sudo systemctl restart medpath
```

Restart backend **mengeluarkan semua pengguna** (sesi di memori). Jadwalkan rilis di luar jam sibuk dan umumkan sebelumnya — sampai sesi persisten dibangun.

## Rollback

Simpan binary dan `dist/` versi sebelumnya (mis. `transcpg-api.prev`). Bila rilis bermasalah: kembalikan file lama, restart service. Skema DB saat ini hanya bertambah (`IF NOT EXISTS`), jadi rollback binary aman selama rilis baru tidak menghapus/mengubah kolom.
