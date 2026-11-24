// Pemilik: Fadli | Modul 6 | Aturan: akses Admin dan Pemilik
package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"penggajian-konveksi/internal/domain"
	"penggajian-konveksi/internal/web"
)

// --- Password / bcrypt ---

func TestLoginMenolakPasswordSalah(t *testing.T) {
	svc := NewService(nil)
	_, err := svc.Login(context.Background(), "admin", "salah")
	if err == nil {
		t.Fatal("password salah seharusnya ditolak")
	}
}

func TestLoginMenolakPenggunaTidakDikenal(t *testing.T) {
	svc := NewService(nil)
	_, err := svc.Login(context.Background(), "tidak-ada", "apa-saja")
	if err == nil {
		t.Fatal("pengguna tidak dikenal seharusnya ditolak")
	}
}

func TestPasswordDisimpanSebagaiHash(t *testing.T) {
	hash, err := HashPasswordBaru("admin123")
	if err != nil {
		t.Fatalf("HashPasswordBaru gagal: %v", err)
	}
	if hash == "admin123" {
		t.Fatal("password tidak boleh disimpan sebagai teks biasa")
	}
	if !strings.HasPrefix(hash, "$2") {
		t.Fatalf("hash harus berformat bcrypt ($2a$ atau $2b$), dapat: %s", hash)
	}
}

func TestPeriksaPasswordBenar(t *testing.T) {
	hash, err := HashPasswordBaru("rahasia")
	if err != nil {
		t.Fatalf("HashPasswordBaru gagal: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("rahasia")); err != nil {
		t.Fatalf("password yang benar seharusnya diterima: %v", err)
	}
}

func TestPeriksaPasswordSalahDitolak(t *testing.T) {
	hash, err := HashPasswordBaru("rahasia")
	if err != nil {
		t.Fatalf("HashPasswordBaru gagal: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("keliru")); err == nil {
		t.Fatal("password yang salah seharusnya ditolak")
	}
}

// --- Sesi ---

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

func TestBersihkanSesiKedaluwarsa(t *testing.T) {
	p := domain.Pengguna{ID: 10, Username: "expire-test", Role: domain.PeranAdmin}

	// Buat sesi yang sudah kedaluwarsa langsung via store.
	muSesi.Lock()
	storeSesi["test-expired"] = entriSesi{
		pengguna:    p,
		kedaluwarsa: time.Now().Add(-1 * time.Hour),
	}
	muSesi.Unlock()

	// Buat sesi aktif.
	activeID, err := BuatSesi(p)
	if err != nil {
		t.Fatalf("buat sesi aktif: %v", err)
	}
	defer HapusSesi(activeID)

	bersihkanSesiKedaluwarsa()

	// Sesi kedaluwarsa harus sudah dihapus.
	muSesi.RLock()
	_, masihAda := storeSesi["test-expired"]
	muSesi.RUnlock()
	if masihAda {
		t.Fatal("sesi kedaluwarsa seharusnya dihapus oleh cleanup")
	}

	// Sesi aktif harus masih ada.
	if AmbilSesi(activeID) == nil {
		t.Fatal("sesi aktif seharusnya tidak dihapus oleh cleanup")
	}
}

func TestLogoutMenghapusSesi(t *testing.T) {
	p := domain.Pengguna{ID: 1, Username: "admin", Role: domain.PeranAdmin}
	sessionID, _ := BuatSesi(p)

	// Pastikan sesi ada.
	if AmbilSesi(sessionID) == nil {
		t.Fatal("sesi harus ada sebelum logout")
	}

	HapusSesi(sessionID)

	// Pastikan sesi sudah dihapus — pengguna tidak bisa pakai session lama.
	if AmbilSesi(sessionID) != nil {
		t.Fatal("sesi seharusnya sudah dihapus setelah logout")
	}
}

// --- Middleware ---

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

func TestStaticDapatDiaksesToanpaSesi(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := MiddlewareAuth(inner)

	req := httptest.NewRequest("GET", "/static/css/style.css", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("/static/ seharusnya bisa diakses tanpa login, dapat %d", rec.Code)
	}
}

func TestMiddlewareAuthMenyediakanCSRFToken(t *testing.T) {
	p := domain.Pengguna{ID: 1, Username: "admin", Role: domain.PeranAdmin}
	sessionID, _ := BuatSesi(p)
	defer HapusSesi(sessionID)

	var tokenDariCtx string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenDariCtx = web.CSRFTokenDariContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	handler := MiddlewareAuth(inner)

	req := httptest.NewRequest("GET", "/dashboard", nil)
	req.AddCookie(&http.Cookie{Name: namaCookie, Value: sessionID})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if tokenDariCtx == "" {
		t.Fatal("middleware auth seharusnya menyediakan CSRF token di context")
	}
}

// --- Rate Limiting ---

func TestRateLimitMemblokirSetelahMaxPercobaan(t *testing.T) {
	user := "test-ratelimit-user"
	defer resetRateLimit(user)

	for i := 0; i < maxPercobaanLogin; i++ {
		if cekRateLimit(user) {
			t.Fatalf("seharusnya belum dikunci setelah %d percobaan", i)
		}
		catatGagalLogin(user)
	}

	if !cekRateLimit(user) {
		t.Fatal("seharusnya dikunci setelah maxPercobaanLogin percobaan gagal")
	}
}

func TestRateLimitResetSetelahLoginBerhasil(t *testing.T) {
	user := "test-ratelimit-reset"
	for i := 0; i < 3; i++ {
		catatGagalLogin(user)
	}
	resetRateLimit(user)

	if cekRateLimit(user) {
		t.Fatal("seharusnya tidak dikunci setelah reset")
	}
}

func TestBersihkanRateLimitKedaluwarsa(t *testing.T) {
	user := "test-cleanup-rl"

	// Masukkan entri yang sudah lewat jendelanya.
	muRL.Lock()
	storeRL[user] = &statusRateLimit{
		percobaan: 1,
		pertama:   time.Now().Add(-10 * time.Minute),
	}
	muRL.Unlock()

	bersihkanRateLimitKedaluwarsa()

	muRL.Lock()
	_, masih := storeRL[user]
	muRL.Unlock()
	if masih {
		t.Fatal("entri rate limit kedaluwarsa seharusnya dihapus oleh cleanup")
	}
}

// --- CSRF ---

func TestCSRFTokenDibuatDanDivalidasi(t *testing.T) {
	rec := httptest.NewRecorder()
	token := buatTokenCSRF(rec)
	if token == "" {
		t.Fatal("CSRF token tidak boleh kosong")
	}

	// Simulasikan form POST dengan token.
	req := httptest.NewRequest("POST", "/login", strings.NewReader("csrf_token="+token+"&username=x&password=y"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatalf("parse form: %v", err)
	}

	if !periksaCSRF(req) {
		t.Fatal("CSRF token yang valid seharusnya diterima")
	}
}

func TestCSRFTokenInvalidDitolak(t *testing.T) {
	req := httptest.NewRequest("POST", "/login", strings.NewReader("csrf_token=token-palsu&username=x&password=y"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatalf("parse form: %v", err)
	}

	if periksaCSRF(req) {
		t.Fatal("CSRF token palsu seharusnya ditolak")
	}
}

func TestCSRFTokenKosongDitolak(t *testing.T) {
	req := httptest.NewRequest("POST", "/login", strings.NewReader("username=x&password=y"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatalf("parse form: %v", err)
	}

	if periksaCSRF(req) {
		t.Fatal("form tanpa CSRF token seharusnya ditolak")
	}
}

func TestHapusTokenCSRFKedaluwarsa(t *testing.T) {
	token := "test-csrf-expired"
	muCSRF.Lock()
	storeCSRF[token] = time.Now().Add(-1 * time.Hour)
	muCSRF.Unlock()

	hapusTokenCSRFKedaluwarsa()

	muCSRF.RLock()
	_, masih := storeCSRF[token]
	muCSRF.RUnlock()
	if masih {
		t.Fatal("CSRF token kedaluwarsa seharusnya dihapus oleh cleanup")
	}
}

func TestAmbilTokenCSRFMembuatBaruJikaTidakAda(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/login", nil)

	token := ambilTokenCSRF(rec, req)
	if token == "" {
		t.Fatal("ambilTokenCSRF harus mengembalikan token baru jika tidak ada cookie")
	}

	// Harus ada cookie CSRF di response.
	cookies := rec.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == namaCookieCSRF && c.Value == token {
			found = true
		}
	}
	if !found {
		t.Fatal("cookie CSRF harus di-set saat membuat token baru")
	}
}
