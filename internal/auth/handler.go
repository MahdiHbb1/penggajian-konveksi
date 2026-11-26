// Pemilik: Fadli | Modul 6 | Aturan: akses Admin dan Pemilik
package auth

import (
	"errors"
	"log"
	"net/http"
	"sync/atomic"

	"penggajian-konveksi/internal/web"
)

// secureCookieFlag menentukan apakah cookie harus di-set Secure (HTTPS/production).
// Diatur setelah config.Muat() memanggil godotenv.Load() di main().
var secureCookieFlag atomic.Bool

// SetSecureCookie mengatur flag Secure untuk cookie sesi dan CSRF secara thread-safe.
func SetSecureCookie(aktif bool) {
	secureCookieFlag.Store(aktif)
}

func isSecureCookie() bool {
	return secureCookieFlag.Load()
}

// buatCookieSesi membuat cookie sesi dengan pengaturan keamanan standar.
func buatCookieSesi(value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     namaCookie,
		Value:    value,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isSecureCookie(),
		Path:     "/",
		MaxAge:   maxAge,
	}
}

// Handler menangani halaman login dan logout.
type Handler struct {
	renderer web.Renderer
	service  *Service
}

// NewHandler membuat handler autentikasi.
func NewHandler(renderer web.Renderer, service *Service) *Handler {
	return &Handler{renderer: renderer, service: service}
}

// DaftarkanRute mendaftarkan login dan logout.
// Logout hanya menerima POST untuk mencegah CSRF via link/image.
func DaftarkanRute(mux *http.ServeMux, renderer web.Renderer, service *Service) {
	h := NewHandler(renderer, service)
	mux.HandleFunc("GET /login", h.Login)
	mux.HandleFunc("POST /login", h.ProsesLogin)
	mux.HandleFunc("POST /logout", h.Logout)
}

// Login menampilkan kartu masuk tanpa navigasi.
// Bila pengguna sudah memiliki sesi sah, langsung arahkan ke dashboard.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if pengguna := web.PenggunaDariContext(r.Context()); pengguna != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	if cookie, err := r.Cookie(namaCookie); err == nil && AmbilSesi(cookie.Value) != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	token := ambilTokenCSRF(w, r)
	h.renderer.Render(w, r, "login.html", web.Data{
		Judul:     "Masuk",
		Polos:     true,
		CSRFToken: token,
	})
}

// ProsesLogin memvalidasi credential dan membuat sesi.
func (h *Handler) ProsesLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		token := ambilTokenCSRF(w, r)
		h.renderer.Render(w, r, "login.html", web.Data{
			Judul:     "Masuk",
			Polos:     true,
			Galat:     "Gagal membaca formulir.",
			CSRFToken: token,
			Status:    http.StatusBadRequest,
		})
		return
	}

	// Validasi CSRF Double-Submit Cookie (fail-closed).
	if !periksaCSRF(r) {
		log.Printf("login handler: CSRF token tidak valid dari %s", r.RemoteAddr)
		token := buatTokenCSRF(w)
		h.renderer.Render(w, r, "login.html", web.Data{
			Judul:     "Masuk",
			Polos:     true,
			Galat:     "Sesi formulir tidak valid. Silakan coba lagi.",
			CSRFToken: token,
			Status:    http.StatusForbidden,
		})
		return
	}

	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	// Reservasi slot rate limiting secara atomik sebelum menjalankan bcrypt (anti-TOCTOU).
	if !mulaiPercobaanLogin(username) {
		log.Printf("login handler: rate limit tercapai untuk %q", kunciRateLimit(username))
		token := ambilTokenCSRF(w, r)
		h.renderer.Render(w, r, "login.html", web.Data{
			Judul:     "Masuk",
			Polos:     true,
			Galat:     "Terlalu banyak percobaan login. Silakan tunggu beberapa menit.",
			CSRFToken: token,
			Status:    http.StatusTooManyRequests,
		})
		return
	}

	berhasil := false
	catatGagal := false
	defer func() {
		selesaiPercobaanLogin(username, berhasil, catatGagal)
	}()

	pengguna, err := h.service.Login(r.Context(), username, password)
	if err != nil {
		pesan := "Terjadi kesalahan sistem."
		status := http.StatusInternalServerError
		if errors.Is(err, ErrCredentialSalah) {
			pesan = "Username atau password salah."
			status = http.StatusUnauthorized
			catatGagal = true
		} else {
			log.Printf("login handler: %v", err)
		}
		token := ambilTokenCSRF(w, r)
		h.renderer.Render(w, r, "login.html", web.Data{
			Judul:     "Masuk",
			Polos:     true,
			Galat:     pesan,
			CSRFToken: token,
			Status:    status,
		})
		return
	}

	sessionID, err := BuatSesi(pengguna)
	if err != nil {
		log.Printf("login handler: gagal membuat sesi: %v", err)
		token := ambilTokenCSRF(w, r)
		h.renderer.Render(w, r, "login.html", web.Data{
			Judul:     "Masuk",
			Polos:     true,
			Galat:     "Gagal membuat sesi. Coba lagi.",
			CSRFToken: token,
			Status:    http.StatusInternalServerError,
		})
		return
	}

	berhasil = true
	// Rotasi sesi dan hapus cookie CSRF pra-login.
	hapusCookieCSRF(w)
	http.SetCookie(w, buatCookieSesi(sessionID, int(durasisesi.Seconds())))
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Logout menghapus sesi dan mengarahkan ke halaman login.
// Hanya menerima POST. CSRF divalidasi secara fail-closed.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("logout handler: gagal membaca form dari %s: %v", r.RemoteAddr, err)
		http.Error(w, "Permintaan tidak valid.", http.StatusForbidden)
		return
	}
	// Defense-in-depth bila handler dipanggil langsung tanpa MiddlewareAuth.
	expectedCSRF := web.CSRFTokenDariContext(r.Context())
	if expectedCSRF != "" {
		if !cocokTokenCSRF(expectedCSRF, r.PostFormValue(namaFieldCSRF)) {
			log.Printf("logout handler: CSRF token tidak valid dari %s", r.RemoteAddr)
			http.Error(w, "Permintaan tidak valid.", http.StatusForbidden)
			return
		}
	} else if !periksaCSRF(r) {
		log.Printf("logout handler: CSRF token tidak valid dari %s", r.RemoteAddr)
		http.Error(w, "Permintaan tidak valid.", http.StatusForbidden)
		return
	}

	if cookie, err := r.Cookie(namaCookie); err == nil {
		_ = h.service.Logout(r.Context(), cookie.Value)
	}
	http.SetCookie(w, buatCookieSesi("", -1))
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
