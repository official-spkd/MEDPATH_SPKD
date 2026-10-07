// MedPath — Clinical Pathway Generator (backend Go).
// Server API untuk frontend Vue. State disimpan di PostgreSQL (fallback in-memory bila DB mati).
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/user"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"transcpg/internal/api"
	"transcpg/internal/store"
)

func main() {
	alamat := os.Getenv("ALAMAT")
	if alamat == "" {
		alamat = ":8080"
	}

	st := store.NewDenganDB(sambungDB())

	srv := api.New(st)
	s := &http.Server{
		Addr:         alamat,
		Handler:      srv.Handler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("MedPath API berjalan di http://127.0.0.1%s/api/v1", alamat)
	log.Printf("akun demo: super_admin@demo.local · sandi %s", store.DemoPassword)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// sambungDB mencoba konek ke PostgreSQL. Mengembalikan nil bila gagal (server lanjut in-memory).
func sambungDB() *pgxpool.Pool {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		u := "postgres"
		if cur, err := user.Current(); err == nil && cur.Username != "" {
			u = cur.Username
		}
		dsn = "postgres://" + u + "@localhost:5432/transcpg?sslmode=disable"
	}

	ctx, batal := context.WithTimeout(context.Background(), 5*time.Second)
	defer batal()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Printf("PostgreSQL tidak dapat dikonfigurasi (%v) — memakai data in-memory.", err)
		return nil
	}
	if err := pool.Ping(ctx); err != nil {
		log.Printf("PostgreSQL tidak merespons (%v) — memakai data in-memory.", err)
		pool.Close()
		return nil
	}
	log.Println("Terhubung ke PostgreSQL.")
	return pool
}
