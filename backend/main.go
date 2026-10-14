// MedPath — API Go untuk Vue, Vercel, dan PostgreSQL/Supabase.
package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/user"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"transcpg/internal/api"
	"transcpg/internal/store"
)

func main() {
	alamat := os.Getenv("ALAMAT")
	if port := os.Getenv("PORT"); port != "" {
		alamat = ":" + port
	}
	if alamat == "" {
		alamat = ":8080"
	}
	var st *store.Store
	if os.Getenv("DEMO_IN_MEMORY") == "true" && os.Getenv("VERCEL") == "" {
		log.Println("Mode demo lokal: data tidak persisten.")
		st = store.New()
	} else {
		pool, err := sambungDB()
		if err != nil {
			log.Fatal(err)
		}
		defer pool.Close()
		st, err = store.NewDenganDB(pool)
		if err != nil {
			log.Fatal("Inisialisasi database gagal; API tidak dijalankan.")
		}
	}
	server := &http.Server{
		Addr: alamat, Handler: api.New(st).Handler(),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second,
		WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second,
	}
	log.Printf("MedPath API mendengarkan %s", alamat)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func sambungDB() (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		if os.Getenv("VERCEL") != "" {
			return nil, errors.New("DATABASE_URL wajib disetel di Vercel")
		}
		username := "postgres"
		if current, err := user.Current(); err == nil && current.Username != "" {
			username = current.Username
		}
		dsn = "postgres://" + username + "@localhost:5432/transcpg?sslmode=disable"
	}
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, errors.New("Konfigurasi DATABASE_URL tidak valid")
	}
	// Supavisor transaction mode tidak mendukung prepared statement bernama.
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	config.MaxConns, config.MinConns = 3, 0
	config.MaxConnIdleTime = 30 * time.Second
	config.MaxConnLifetime = 5 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, errors.New("Pool PostgreSQL tidak dapat dibuat")
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, databaseError(err)
	}
	return pool, nil
}

// Klasifikasi tanpa mencetak DSN, password, parameter, atau pesan server mentah.
func databaseError(err error) error {
	var postgres *pgconn.PgError
	if errors.As(err, &postgres) {
		switch postgres.Code {
		case "28P01", "28000":
			return errors.New("PostgreSQL menolak autentikasi; periksa user dan password pada DATABASE_URL")
		case "3D000":
			return errors.New("Nama database pada DATABASE_URL tidak ditemukan")
		case "53300":
			return errors.New("Batas koneksi PostgreSQL tercapai")
		default:
			return errors.New("PostgreSQL menolak koneksi; SQLSTATE=" + postgres.Code)
		}
	}
	var dns *net.DNSError
	if errors.As(err, &dns) {
		return errors.New("Hostname PostgreSQL tidak dapat di-resolve")
	}
	var network net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &network) && network.Timeout()) {
		return errors.New("Koneksi PostgreSQL timeout; periksa host, port, dan network restrictions")
	}
	return errors.New("Koneksi PostgreSQL gagal sebelum autentikasi; periksa TLS, host, dan jaringan")
}
