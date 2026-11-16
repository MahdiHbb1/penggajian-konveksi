// Pemilik: Fadli | Modul 6 | Aturan: akses Admin dan Pemilik
package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"penggajian-konveksi/internal/domain"
	"penggajian-konveksi/internal/web"
)

// TestLoginMenolakPasswordSalah memastikan password yang salah ditolak tanpa DB.
func TestLoginMenolakPasswordSalah(t *testing.T) {
	svc := NewService(nil)
	_, err := svc.Login(context.Background(), "admin", "salah")
	if err == nil {
		t.Fatal("password salah seharusnya ditolak")
	}
}

// TestLoginMenolakPenggunaTidakDikenal memastikan pengguna tak dikenal ditolak.
func TestLoginMenolakPenggunaTidakDikenal(t *testing.T) {
	svc := NewService(nil)
	_, err := svc.Login(context.Background(), "tidak-ada", "apa-saja")
	if err == nil {
		t.Fatal("pengguna tidak dikenal seharusnya ditolak")
	}
}

// TestPasswordDisimpanSebagaiHash memastikan hash tidak sama dengan teks biasa.
func TestPasswordDisimpanSebagaiHash(t *testing.T) {
	hash := HashPasswordBaru("admin123")
	if hash == "admin123" {
		t.Fatal("password tidak boleh disimpan sebagai teks biasa")
	}
	if !strings.HasPrefix(hash, "sha512:") {
		t.Fatalf("hash harus berformat sha512:salt:hash, dapat: %s", hash)
	}
}

// TestPeriksaPasswordBenar memastikan password yang cocok diterima.
func TestPeriksaPasswordBenar(t *testing.T) {
	hash := HashPasswordBaru("rahasia")
	if !periksaPassword("rahasia", hash) {
		t.Fatal("password yang benar seharusnya diterima")
	}
}

// TestPeriksaPasswordSalahDitolak memastikan password keliru ditolak.
func TestPeriksaPasswordSalahDitolak(t *testing.T) {
	hash := HashPasswordBaru("rahasia")
	if periksaPassword("keliru", hash) {
		t.Fatal("password yang salah seharusnya ditolak")
	}
}

// TestSesiDibuatDanDihapus memastikan store sesi berfungsi.
func TestSesiDibuatDanDihapus(t *testing.T) {
	p := domain.Pengguna{ID: 1, Username: "admin", Role: domain.PeranAdmin}
	id, err := BuatSesi(p)
	if err != nil {
		t.Fatalf("buat sesi: %v", err)
	}
	got := AmbilSesi(id)
	if got == nil {
		t.Fatal("sesi seharusnya ditemukan setelah dibuat")
	}
	if got.Username != "admin" {
		t.Fatalf("username sesi salah: %s", got.Username)
	}
	HapusSesi(id)
	if AmbilSesi(id) != nil {
		t.Fatal("sesi seharusnya kosong setelah dihapus")
	}
}

// TestMiddlewarePeranMenolakPemilikDiHalamanAdmin memastikan Pemilik tidak masuk rute Admin.
func TestMiddlewarePeranMenolakPemilikDiHalamanAdmin(t *testing.T) {
	pemilik := &domain.Pengguna{ID: 2, Username: "pemilik", Role: domain.PeranPemilik}
	handler := MiddlewarePeran(domain.PeranAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/pekerja", nil)
	req = req.WithContext(web.DenganPengguna(req.Context(), pemilik))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("pemilik di rute admin seharusnya 403, dapat %d", rec.Code)
	}
}

// TestMiddlewarePeranMenerimaPemilikDiHalamanRekap memastikan Pemilik bisa akses rekap.
func TestMiddlewarePeranMenerimaPemilikDiHalamanRekap(t *testing.T) {
	pemilik := &domain.Pengguna{ID: 2, Username: "pemilik", Role: domain.PeranPemilik}
	handler := MiddlewarePeran(domain.PeranAdmin, domain.PeranPemilik)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/rekap/1", nil)
	req = req.WithContext(web.DenganPengguna(req.Context(), pemilik))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("pemilik di rekap seharusnya 200, dapat %d", rec.Code)
	}
}

// TestMiddlewareAuthRedirectTanpaSesi memastikan request tanpa sesi diarahkan ke /login.
func TestMiddlewareAuthRedirectTanpaSesi(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := MiddlewareAuth(inner)

	req := httptest.NewRequest("GET", "/dashboard", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("tanpa sesi seharusnya redirect 303, dapat %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/login" {
		t.Fatalf("redirect seharusnya ke /login, dapat %s", loc)
	}
}

// TestMiddlewareAuthLoloskanSesiSah memastikan request bersesi valid diteruskan.
func TestMiddlewareAuthLoloskanSesiSah(t *testing.T) {
	p := domain.Pengguna{ID: 1, Username: "admin", Role: domain.PeranAdmin}
	sessionID, _ := BuatSesi(p)
	defer HapusSesi(sessionID)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pengguna := web.PenggunaDariContext(r.Context())
		if pengguna == nil {
			t.Error("pengguna seharusnya ada di context")
		}
		w.WriteHeader(http.StatusOK)
	})
	handler := MiddlewareAuth(inner)

	req := httptest.NewRequest("GET", "/dashboard", nil)
	req.AddCookie(&http.Cookie{Name: namaCookie, Value: sessionID})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("sesi sah seharusnya 200, dapat %d", rec.Code)
	}
}
