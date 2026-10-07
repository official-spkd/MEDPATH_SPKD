package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"transcpg/internal/store"
)

func login(t *testing.T, base, email string) string {
	t.Helper()
	body := `{"email":"` + email + `","password":"` + store.DemoPassword + `"}`
	res, err := http.Post(base+"/api/v1/auth/login", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var out struct{ Token string }
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil || out.Token == "" {
		t.Fatalf("login %s gagal (status %d)", email, res.StatusCode)
	}
	return out.Token
}

func getPanduan(t *testing.T, base, token string) *http.Response {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, base+"/api/v1/panduan", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestPanduanHanyaSuperAdmin(t *testing.T) {
	ts := httptest.NewServer(New(store.New()).Handler())
	defer ts.Close()

	// Tanpa token → 401
	if res := getPanduan(t, ts.URL, ""); res.StatusCode != http.StatusUnauthorized {
		t.Errorf("tanpa token: status %d, harusnya 401", res.StatusCode)
	}

	// Peran lain (termasuk admin non-super) → 403
	for _, email := range []string{"direktur@demo.local", "tim_cp@demo.local", "komite_medik@demo.local"} {
		res := getPanduan(t, ts.URL, login(t, ts.URL, email))
		if res.StatusCode != http.StatusForbidden {
			t.Errorf("%s: status %d, harusnya 403", email, res.StatusCode)
		}
		res.Body.Close()
	}

	// Super Admin → 200 berisi bab
	res := getPanduan(t, ts.URL, login(t, ts.URL, "super_admin@demo.local"))
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("super admin: status %d, harusnya 200", res.StatusCode)
	}
	var out struct {
		Data []struct {
			ID    string `json:"id"`
			Judul string `json:"judul"`
			Isi   string `json:"isi"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if len(out.Data) == 0 || out.Data[0].Judul == "" || out.Data[0].Isi == "" {
		t.Fatalf("respons panduan kosong: %+v", out.Data)
	}
}
