package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"transcpg/internal/store"
)

// Tests use an isolated schema, never existing MedPath tables.
func TestPostgresAcrossInstances(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to a disposable PostgreSQL database")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	schema := fmt.Sprintf("medpath_test_%d", time.Now().UnixNano())
	ident := pgx.Identifier{schema}.Sanitize()
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+ident); err != nil {
		t.Fatal(err)
	}
	defer admin.Exec(ctx, "DROP SCHEMA "+ident+" CASCADE")
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	// Simultaneous cold starts must seed exactly once.
	var stores [2]*store.Store
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for i := range stores {
		wg.Add(1)
		go func(i int) { defer wg.Done(); var e error; stores[i], e = store.NewDenganDB(pool); errs <- e }(i)
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	a := httptest.NewServer(New(stores[0]).Handler())
	defer a.Close()
	b := httptest.NewServer(New(stores[1]).Handler())
	defer b.Close()
	token := login(t, a.URL, "super_admin@demo.local")
	request := func(base, method, path, body string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, "/api/v1"+path, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		idx := 0
		if base == b.URL {
			idx = 1
		}
		New(stores[idx]).Handler().ServeHTTP(rec, req)
		return rec
	}
	if r := request(b.URL, "GET", "/auth/me", ""); r.Code != 200 {
		t.Fatalf("session missing on second instance: %d", r.Code)
	}
	var rawCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM auth_sessions WHERE token_hash=$1`, token).Scan(&rawCount); err != nil {
		t.Fatal(err)
	}
	if rawCount != 0 {
		t.Fatal("raw token persisted")
	}
	if r := request(b.URL, "GET", "/cp", ""); r.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("API response can be cached")
	}

	// Two simultaneous inserts into the same CP must both survive, with distinct IDs.
	var code string
	if err := pool.QueryRow(ctx, `SELECT kode FROM cp_state WHERE status='DRAFT' ORDER BY kode LIMIT 1`).Scan(&code); err != nil {
		t.Fatal(err)
	}
	var before int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM cp_kriteria WHERE kode=$1`, code).Scan(&before); err != nil {
		t.Fatal(err)
	}
	results := make(chan int, 2)
	for _, base := range []string{a.URL, b.URL} {
		wg.Add(1)
		go func(base string) {
			defer wg.Done()
			results <- request(base, "POST", "/cp/"+code+"/kriteria", `{"jenis":"INKLUSI","uraian":"Integration test"}`).Code
		}(base)
	}
	wg.Wait()
	close(results)
	for status := range results {
		if status != 200 {
			t.Fatalf("concurrent write failed: %d", status)
		}
	}
	var after int
	pool.QueryRow(ctx, `SELECT count(*) FROM cp_kriteria WHERE kode=$1`, code).Scan(&after)
	if after != before+2 {
		t.Fatalf("lost update: before=%d after=%d", before, after)
	}
	r := request(a.URL, "GET", "/cp/"+code, "")
	if r.Code != 200 || strings.Count(r.Body.String(), "Integration test") != 2 {
		t.Fatal("stale data on another instance")
	}

	// A failed audit INSERT must undo its accompanying status UPDATE.
	if _, err := pool.Exec(ctx, `CREATE FUNCTION reject_audit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected test failure'; END $$; CREATE TRIGGER reject_audit BEFORE INSERT ON cp_riwayat FOR EACH ROW EXECUTE FUNCTION reject_audit()`); err != nil {
		t.Fatal(err)
	}
	r = request(a.URL, "POST", "/cp/"+code+"/transisi", `{"aksi":"AJUKAN"}`)
	if r.Code != 503 {
		t.Fatalf("failed write reported status %d", r.Code)
	}
	var status string
	if err := pool.QueryRow(ctx, `SELECT status FROM cp_state WHERE kode=$1`, code).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "DRAFT" {
		t.Fatal("partial transaction committed")
	}
	if _, err := pool.Exec(ctx, `DROP TRIGGER reject_audit ON cp_riwayat`); err != nil {
		t.Fatal(err)
	}
	r = request(b.URL, "POST", "/cp/"+code+"/transisi", `{"aksi":"AJUKAN"}`)
	if r.Code != 200 {
		t.Fatalf("transition failed: %d %s", r.Code, r.Body.String())
	}
	var state struct {
		CP struct {
			Status string `json:"status"`
		} `json:"cp"`
	}
	if err := json.Unmarshal(r.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if state.CP.Status == "DRAFT" || state.CP.Status == "" {
		t.Fatal("transition did not advance")
	}

	// Administrative state and newly provisioned users survive instance boundaries.
	for _, tc := range []struct {
		method, path, body string
		status             int
	}{
		{"PUT", "/admin/profil", `{"nama":"RS Uji Persistensi","tipe":"B","kepemilikan":"Pemerintah","regional":"1","kelas":"1"}`, 200},
		{"POST", "/dokumen-panduan", `{"nama":"Dokumen Uji Persistensi","penerbit":"QA","tahun":2026,"sumber":"PPK_RS","butir":["Butir pengujian, bukan pedoman klinis"]}`, 201},
		{"POST", "/admin/pengguna", `{"nama":"Administrator Uji","email":"admin@example.test","peran":"SUPER_ADMIN","password":"integration-fixture-only"}`, 200},
	} {
		if r := request(a.URL, tc.method, tc.path, tc.body); r.Code != tc.status {
			t.Fatalf("administration %s: %d %s", tc.path, r.Code, r.Body.String())
		}
	}
	if r := request(b.URL, "GET", "/admin", ""); r.Code != 200 || !strings.Contains(r.Body.String(), "RS Uji Persistensi") || !strings.Contains(r.Body.String(), "admin@example.test") || strings.Contains(r.Body.String(), "password") {
		t.Fatal("administration not persistent or secret exposed")
	}
	if r := request(b.URL, "GET", "/dokumen-panduan", ""); r.Code != 200 || !strings.Contains(r.Body.String(), "Dokumen Uji Persistensi") {
		t.Fatal("document not persistent")
	}
	var encoded string
	if err := pool.QueryRow(ctx, "SELECT password_hash FROM app_users WHERE email='admin@example.test'").Scan(&encoded); err != nil || !strings.HasPrefix(encoded, "pbkdf2-sha256$") {
		t.Fatal("password not hashed")
	}
	loginRecord := request(b.URL, "POST", "/auth/login", `{"email":"admin@example.test","password":"integration-fixture-only"}`)
	if loginRecord.Code != 200 {
		t.Fatalf("new account login: %d", loginRecord.Code)
	}
	if r := request(a.URL, "POST", "/admin/nonaktifkan-demo", ""); r.Code != 422 {
		t.Fatal("demo can retire itself and lock owner out")
	}
	keyRecord := request(a.URL, "POST", "/admin/kunci", `{"nama":"Integration fixture"}`)
	var key struct {
		Kunci string `json:"kunci"`
	}
	if keyRecord.Code != 201 || json.Unmarshal(keyRecord.Body.Bytes(), &key) != nil || key.Kunci == "" {
		t.Fatal("key creation failed")
	}
	keyReq := httptest.NewRequest("GET", "/api/v1/integrasi/cp-aktif", nil)
	keyReq.Header.Set("Authorization", "Bearer "+key.Kunci)
	keyRes := httptest.NewRecorder()
	New(stores[1]).Handler().ServeHTTP(keyRes, keyReq)
	if keyRes.Code != 200 {
		t.Fatal("integration key not persistent")
	}
	if r := request(b.URL, "GET", "/admin", ""); strings.Contains(r.Body.String(), key.Kunci) || strings.Contains(r.Body.String(), `"hash"`) {
		t.Fatal("integration secret exposed")
	}

	// Logout and expiry apply to every warm or newly started instance.
	if r := request(b.URL, "POST", "/auth/logout", ""); r.Code != 200 {
		t.Fatal("logout failed")
	}
	if r := request(a.URL, "GET", "/auth/me", ""); r.Code != 401 {
		t.Fatal("logout not propagated")
	}
	token = login(t, a.URL, "super_admin@demo.local")
	if _, err := pool.Exec(ctx, `UPDATE auth_sessions SET expires_at=now()-interval '1 second'`); err != nil {
		t.Fatal(err)
	}
	if r := request(b.URL, "GET", "/auth/me", ""); r.Code != 401 {
		t.Fatal("expired session accepted")
	}

	// A read failure must not return partial/fabricated data or success health.
	if _, err := pool.Exec(ctx, `ALTER TABLE cp_acuan RENAME TO unavailable_cp_acuan`); err != nil {
		t.Fatal(err)
	}
	if r := request(a.URL, "GET", "/kesehatan", ""); r.Code != http.StatusServiceUnavailable {
		t.Fatalf("database failure masked: %d", r.Code)
	}
}
