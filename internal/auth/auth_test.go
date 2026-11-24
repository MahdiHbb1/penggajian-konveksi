// Pemilik: Fadli | Modul 6 | Aturan: akses Admin dan Pemilik
package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
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

// --- Sesi & Secure Cookie ---

func TestSesiDibuatDanDihapus(t *testing.T) {
	p := domain.Pengguna{ID: 1, Username: "admin", Role: domain.PeranAdmin}
	id, err := BuatSesi(p)
	if err != nil {
		t.Fatalf("buat sesi: %v", err)
	}
	got, csrf := ambilEntriSesi(id)
	if got == nil {
		t.Fatal("sesi seharusnya ditemukan setelah dibuat")
	}
	if got.Username != "admin" {
		t.Fatalf("username sesi salah: %s", got.Username)
	}
	if !tokenHexValid(csrf) {
		t.Fatalf("sesi harus memiliki token CSRF terikat yang valid, dapat %q", csrf)
	}
	HapusSesi(id)
	if AmbilSesi(id) != nil {
		t.Fatal("sesi seharusnya kosong setelah dihapus")
	}
}

func TestSetSecureCookieBerlakuDinamis(t *testing.T) {
	defer SetSecureCookie(false)

	SetSecureCookie(true)
	c := buatCookieSesi("abc", 3600)
	if !c.Secure || !c.HttpOnly || c.SameSite != http.SameSiteLaxMode {
		t.Fatalf("cookie sesi tidak memenuhi atribut keamanan: %+v", c)
	}

	SetSecureCookie(false)
	c2 := buatCookieSesi("abc", 3600)
	if c2.Secure {
		t.Fatal("cookie seharusnya tidak Secure saat flag false")
	}
}

func TestBersihkanSesiKedaluwarsa(t *testing.T) {
	p := domain.Pengguna{ID: 10, Username: "expire-test", Role: domain.PeranAdmin}

	muSesi.Lock()
	storeSesi["test-expired"] = entriSesi{
		pengguna:    p,
		csrfToken:   "expired-token",
		kedaluwarsa: time.Now().Add(-1 * time.Hour),
	}
	muSesi.Unlock()

	activeID, err := BuatSesi(p)
	if err != nil {
		t.Fatalf("buat sesi aktif: %v", err)
	}
	defer HapusSesi(activeID)

	bersihkanSesiKedaluwarsa()

	muSesi.RLock()
	_, masihAda := storeSesi["test-expired"]
	muSesi.RUnlock()
	if masihAda {
		t.Fatal("sesi kedaluwarsa seharusnya dihapus oleh cleanup")
	}
	if AmbilSesi(activeID) == nil {
		t.Fatal("sesi aktif seharusnya tidak dihapus oleh cleanup")
	}
}

// --- Middleware & RBAC ---

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
	if !strings.Contains(rec.Body.String(), "Anda tidak memiliki akses ke halaman ini.") {
		t.Fatalf("pesan 403 tidak sesuai: %q", rec.Body.String())
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

	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/login" {
		t.Fatalf("tanpa sesi seharusnya 303 -> /login, dapat %d -> %s", rec.Code, rec.Header().Get("Location"))
	}
	if rec.Header().Get("X-Frame-Options") != "DENY" || rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("header keamanan harus terpasang pada respons")
	}
}

func TestMiddlewareAuthMenolakSesiPalsuAtauKedaluwarsa(t *testing.T) {
	dipanggil := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dipanggil = true
		w.WriteHeader(http.StatusOK)
	})
	handler := MiddlewareAuth(inner)

	req := httptest.NewRequest("GET", "/dashboard", nil)
	req.AddCookie(&http.Cookie{Name: namaCookie, Value: "sesi-tidak-ada"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if dipanggil {
		t.Fatal("handler dalam tidak boleh dipanggil untuk sesi palsu")
	}
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/login" {
		t.Fatalf("sesi palsu seharusnya redirect 303 -> /login, dapat %d", rec.Code)
	}
}

func TestMiddlewareAuthLoloskanSesiSahDanSediakanCSRF(t *testing.T) {
	p := domain.Pengguna{ID: 1, Username: "admin", Role: domain.PeranAdmin}
	sessionID, _ := BuatSesi(p)
	defer HapusSesi(sessionID)

	var tokenDariCtx string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if web.PenggunaDariContext(r.Context()) == nil {
			t.Error("pengguna seharusnya ada di context")
		}
		tokenDariCtx = web.CSRFTokenDariContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	handler := MiddlewareAuth(inner)

	req := httptest.NewRequest("GET", "/dashboard", nil)
	req.AddCookie(&http.Cookie{Name: namaCookie, Value: sessionID})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK || tokenDariCtx == "" {
		t.Fatalf("sesi sah seharusnya 200 dengan CSRF token, dapat %d token=%q", rec.Code, tokenDariCtx)
	}
}

func TestMiddlewareAuthMenolakPOSTLintasSesiDanTanpaCSRF(t *testing.T) {
	korbanID, _ := BuatSesi(domain.Pengguna{ID: 1, Username: "admin", Role: domain.PeranAdmin})
	defer HapusSesi(korbanID)
	penyerangID, _ := BuatSesi(domain.Pengguna{ID: 2, Username: "other", Role: domain.PeranAdmin})
	defer HapusSesi(penyerangID)

	_, csrfKorban := ambilEntriSesi(korbanID)
	_, csrfPenyerang := ambilEntriSesi(penyerangID)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := MiddlewareAuth(inner)

	// 1. Tanpa CSRF token -> 403
	req1 := httptest.NewRequest("POST", "/pekerja/baru", strings.NewReader("kode=X"))
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req1.AddCookie(&http.Cookie{Name: namaCookie, Value: korbanID})
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusForbidden {
		t.Fatalf("POST tanpa CSRF seharusnya 403, dapat %d", rec1.Code)
	}

	// 2. Pakai CSRF token sesi lain -> 403 (anti cross-session bypass)
	req2 := httptest.NewRequest("POST", "/pekerja/baru", strings.NewReader("csrf_token="+csrfPenyerang))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.AddCookie(&http.Cookie{Name: namaCookie, Value: korbanID})
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusForbidden {
		t.Fatalf("POST dengan CSRF sesi lain seharusnya 403, dapat %d", rec2.Code)
	}

	// 3. Body rusak (fail-closed) -> 403
	req3 := httptest.NewRequest("POST", "/logout", strings.NewReader("bad=%ZZ"))
	req3.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req3.AddCookie(&http.Cookie{Name: namaCookie, Value: korbanID})
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusForbidden {
		t.Fatalf("POST dengan form rusak seharusnya 403, dapat %d", rec3.Code)
	}
	if AmbilSesi(korbanID) == nil {
		t.Fatal("sesi korban tidak boleh terhapus oleh POST fail-open")
	}

	// 4. Token milik sesi sendiri -> 200
	req4 := httptest.NewRequest("POST", "/pekerja/baru", strings.NewReader("csrf_token="+csrfKorban))
	req4.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req4.AddCookie(&http.Cookie{Name: namaCookie, Value: korbanID})
	rec4 := httptest.NewRecorder()
	handler.ServeHTTP(rec4, req4)
	if rec4.Code != http.StatusOK {
		t.Fatalf("POST dengan CSRF sah seharusnya 200, dapat %d", rec4.Code)
	}
}

func TestStaticDapatDiaksesTanpaSesi(t *testing.T) {
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

// --- Handler Login / Logout ---

func TestLoginHandlerRedirectSaatSudahLogin(t *testing.T) {
	renderer := &web.Render{Dir: filepath.Join("..", "..", "web", "templates")}
	mux := http.NewServeMux()
	DaftarkanRute(mux, renderer, NewService(nil))
	handler := MiddlewareAuth(mux)

	sid, _ := BuatSesi(domain.Pengguna{ID: 1, Username: "admin", Role: domain.PeranAdmin})
	defer HapusSesi(sid)

	req := httptest.NewRequest("GET", "/login", nil)
	req.AddCookie(&http.Cookie{Name: namaCookie, Value: sid})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/dashboard" {
		t.Fatalf("user sudah login di GET /login seharusnya 303 -> /dashboard, dapat %d -> %s", rec.Code, rec.Header().Get("Location"))
	}
}

func TestProsesLoginMenolakCSRFInvalidDanMencatatRateLimit(t *testing.T) {
	renderer := &web.Render{Dir: filepath.Join("..", "..", "web", "templates")}
	h := NewHandler(renderer, NewService(nil))
	user := "uji-handler-rl"
	defer resetRateLimit(user)

	// 1. Tanpa cookie CSRF -> 403
	reqBad := httptest.NewRequest("POST", "/login", strings.NewReader("csrf_token=palsu&username="+user+"&password=x"))
	reqBad.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recBad := httptest.NewRecorder()
	h.ProsesLogin(recBad, reqBad)
	if recBad.Code != http.StatusForbidden {
		t.Fatalf("login dengan CSRF invalid seharusnya 403, dapat %d", recBad.Code)
	}

	// 2. Ambil cookie CSRF sah lewat GET /login lalu gagal 5x -> ke-6 harus 429
	recGet := httptest.NewRecorder()
	h.Login(recGet, httptest.NewRequest("GET", "/login", nil))
	var csrfCookie *http.Cookie
	for _, c := range recGet.Result().Cookies() {
		if c.Name == namaCookieCSRF {
			csrfCookie = c
		}
	}
	if csrfCookie == nil {
		t.Fatal("GET /login harus menyetel cookie CSRF")
	}

	for i := 0; i < maxPercobaanLogin; i++ {
		req := httptest.NewRequest("POST", "/login", strings.NewReader("csrf_token="+csrfCookie.Value+"&username="+user+"&password=salah"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(csrfCookie)
		rec := httptest.NewRecorder()
		h.ProsesLogin(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("percobaan %d seharusnya 401, dapat %d", i+1, rec.Code)
		}
	}

	// Percobaan ke-6 (bahkan dengan huruf kapital!) harus 429
	req6 := httptest.NewRequest("POST", "/login", strings.NewReader("csrf_token="+csrfCookie.Value+"&username="+strings.ToUpper(user)+"&password=salah"))
	req6.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req6.AddCookie(csrfCookie)
	rec6 := httptest.NewRecorder()
	h.ProsesLogin(rec6, req6)
	if rec6.Code != http.StatusTooManyRequests {
		t.Fatalf("percobaan ke-6 seharusnya 429, dapat %d", rec6.Code)
	}
}

// --- Rate Limiting (Case-insensitive, Anti-TOCTOU, Anti Early-Unlock) ---

func TestRateLimitMemblokirSetelahMaxPercobaanDanCaseInsensitive(t *testing.T) {
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
	if !cekRateLimit("TEST-RATELIMIT-USER") || !cekRateLimit("  Test-RateLimit-User ") {
		t.Fatal("rate limit harus case-insensitive agar tidak bisa dibypass kolasi MySQL")
	}
}

func TestRateLimitKonkurenAntiTOCTOU(t *testing.T) {
	user := "test-toctou-burst"
	defer resetRateLimit(user)

	const totalReq = 12
	var lolos atomic.Int32
	var wg sync.WaitGroup
	mulai := make(chan struct{})

	for i := 0; i < totalReq; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-mulai
			if mulaiPercobaanLogin(user) {
				lolos.Add(1)
				time.Sleep(15 * time.Millisecond)
				selesaiPercobaanLogin(user, false, true)
			}
		}()
	}
	close(mulai)
	wg.Wait()

	if got := int(lolos.Load()); got != maxPercobaanLogin {
		t.Fatalf("konkuren burst: lolos=%d, seharusnya tepat %d (anti-TOCTOU)", got, maxPercobaanLogin)
	}
	if !cekRateLimit(user) {
		t.Fatal("username harus terkunci setelah burst gagal")
	}
}

func TestRateLimitTidakBukaKunciLebihAwal(t *testing.T) {
	user := "test-no-early-unlock"
	defer resetRateLimit(user)

	muRL.Lock()
	storeRL[kunciRateLimit(user)] = &statusRateLimit{
		percobaan:   maxPercobaanLogin,
		pertama:     time.Now().Add(-10 * time.Minute), // lewat jendelaRateLimit (5m)
		dikunciSamp: time.Now().Add(10 * time.Minute),  // namun masih dikunci 10m lagi!
	}
	muRL.Unlock()

	catatGagalLogin(user)
	if !cekRateLimit(user) {
		t.Fatal("catatGagalLogin tidak boleh membatalkan kunci yang masih aktif")
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

	muRL.Lock()
	storeRL[kunciRateLimit(user)] = &statusRateLimit{
		percobaan: 1,
		pertama:   time.Now().Add(-10 * time.Minute),
	}
	muRL.Unlock()

	bersihkanRateLimitKedaluwarsa()

	muRL.Lock()
	_, masih := storeRL[kunciRateLimit(user)]
	muRL.Unlock()
	if masih {
		t.Fatal("entri rate limit kedaluwarsa seharusnya dihapus oleh cleanup")
	}
}

// --- CSRF Double-Submit Cookie (Pra-login) ---

func TestCSRFDoubleSubmitDibuatDanDivalidasi(t *testing.T) {
	rec := httptest.NewRecorder()
	token := buatTokenCSRF(rec)
	if !tokenHexValid(token) {
		t.Fatalf("CSRF token tidak valid: %q", token)
	}
	var cookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == namaCookieCSRF {
			cookie = c
		}
	}
	if cookie == nil || !cookie.HttpOnly {
		t.Fatal("cookie CSRF harus terpasang dan HttpOnly")
	}

	// Cocok antara cookie dan form -> valid
	req := httptest.NewRequest("POST", "/login", strings.NewReader("csrf_token="+token))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookie)
	_ = req.ParseForm()
	if !periksaCSRF(req) {
		t.Fatal("CSRF token yang cocok dengan cookie seharusnya diterima")
	}

	// Form valid tapi tanpa cookie (atau cookie berbeda) -> ditolak
	reqNoCookie := httptest.NewRequest("POST", "/login", strings.NewReader("csrf_token="+token))
	reqNoCookie.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = reqNoCookie.ParseForm()
	if periksaCSRF(reqNoCookie) {
		t.Fatal("CSRF token tanpa cookie pasangan seharusnya ditolak")
	}
}
