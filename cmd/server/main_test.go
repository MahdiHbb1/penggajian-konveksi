// Pemilik: Mahdi | Modul 1 | Aturan: akses Admin dan Pemilik, keamanan lintas modul
package main

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"penggajian-konveksi/internal/auth"
	"penggajian-konveksi/internal/domain"
	"penggajian-konveksi/internal/web"
)

func handlerIntegrasi(t *testing.T) http.Handler {
	t.Helper()
	renderer := &web.Render{Dir: filepath.Join("..", "..", "web", "templates")}
	staticDir := filepath.Join("..", "..", "web", "static")
	return bangunHandler(nil, renderer, staticDir)
}

func TestRBACPemilikDitolakDiSemuaRuteAdmin(t *testing.T) {
	h := handlerIntegrasi(t)

	sidPemilik, err := auth.BuatSesi(domain.Pengguna{ID: 2, Username: "pemilik", Role: domain.PeranPemilik})
	if err != nil {
		t.Fatalf("buat sesi pemilik: %v", err)
	}
	defer auth.HapusSesi(sidPemilik)

	ruteAdmin := []string{
		"/pekerja", "/pekerja/baru",
		"/jadwal", "/kasbon", "/kasbon/baru",
		"/absensi", "/absensi/belum-lengkap",
		"/penggajian", "/periode", "/periode/baru",
		"/koreksi", "/koreksi/baru",
	}
	for _, path := range ruteAdmin {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.AddCookie(&http.Cookie{Name: "sesi_penggajian", Value: sidPemilik})
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("PEMILIK GET %s: status=%d, ingin 403", path, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "Anda tidak memiliki akses ke halaman ini.") {
			t.Fatalf("PEMILIK GET %s: body tidak memuat pesan akses ditolak: %q", path, rec.Body.String())
		}
	}

	// Pemilik tetap bisa mengakses /dashboard dan /rekap (200 OK).
	for _, path := range []string{"/dashboard", "/rekap"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.AddCookie(&http.Cookie{Name: "sesi_penggajian", Value: sidPemilik})
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("PEMILIK GET %s: status=%d, ingin 200", path, rec.Code)
		}
	}
}

func TestPOSTAdminWajibCSRFCocokSesi(t *testing.T) {
	h := handlerIntegrasi(t)

	sidAdmin, err := auth.BuatSesi(domain.Pengguna{ID: 1, Username: "admin", Role: domain.PeranAdmin})
	if err != nil {
		t.Fatalf("buat sesi admin: %v", err)
	}
	defer auth.HapusSesi(sidAdmin)

	for _, path := range []string{"/pekerja/baru", "/jadwal", "/kasbon/baru", "/absensi", "/penggajian", "/periode/baru", "/koreksi/baru", "/logout"} {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader("foo=bar"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "sesi_penggajian", Value: sidAdmin})
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("ADMIN POST %s tanpa CSRF: status=%d, ingin 403", path, rec.Code)
		}
	}
}

func TestRutePublikStaticDanLogin(t *testing.T) {
	h := handlerIntegrasi(t)

	recStatic := httptest.NewRecorder()
	h.ServeHTTP(recStatic, httptest.NewRequest(http.MethodGet, "/static/css/style.css", nil))
	if recStatic.Code != http.StatusOK {
		t.Fatalf("GET /static/css/style.css tanpa sesi: status=%d, ingin 200", recStatic.Code)
	}

	recLogin := httptest.NewRecorder()
	h.ServeHTTP(recLogin, httptest.NewRequest(http.MethodGet, "/login", nil))
	if recLogin.Code != http.StatusOK {
		t.Fatalf("GET /login tanpa sesi: status=%d, ingin 200", recLogin.Code)
	}
}
