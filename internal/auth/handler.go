// Pemilik: Fadli | Modul 6 | Aturan: akses Admin dan Pemilik
package auth

import (
	"errors"
	"log"
	"net/http"
	"os"

	"penggajian-konveksi/internal/web"
)

// secureCookie menentukan apakah cookie harus di-set Secure.
// Aktif jika variabel lingkungan SECURE_COOKIE=true (untuk production/HTTPS).
var secureCookie = os.Getenv("SECURE_COOKIE") == "true"

// buatCookieSesi membuat cookie sesi dengan pengaturan keamanan standar.
func buatCookieSesi(value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     namaCookie,
		Value:    value,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secureCookie,
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
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	// Bila sudah login, langsung ke dashboard.
	if pengguna := web.PenggunaDariContext(r.Context()); pengguna != nil {
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
		})
		return
	}

	// Validasi CSRF token.
	if !periksaCSRF(r) {
		log.Printf("login handler: CSRF token tidak valid dari %s", r.RemoteAddr)
		token := ambilTokenCSRF(w, r)
		h.renderer.Render(w, r, "login.html", web.Data{
			Judul:     "Masuk",
			Polos:     true,
			Galat:     "Sesi formulir tidak valid. Silakan coba lagi.",
			CSRFToken: token,
		})
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// Cek rate limiting.
	if cekRateLimit(username) {
		log.Printf("login handler: rate limit tercapai untuk %q", username)
		token := ambilTokenCSRF(w, r)
		h.renderer.Render(w, r, "login.html", web.Data{
			Judul:     "Masuk",
			Polos:     true,
			Galat:     "Terlalu banyak percobaan login. Silakan tunggu beberapa menit.",
			CSRFToken: token,
		})
		return
	}

	pengguna, err := h.service.Login(r.Context(), username, password)
	if err != nil {
		pesan := "Terjadi kesalahan sistem."
		if errors.Is(err, ErrCredentialSalah) {
			pesan = "Username atau password salah."
			catatGagalLogin(username)
		} else {
			log.Printf("login handler: %v", err)
		}
		token := ambilTokenCSRF(w, r)
		h.renderer.Render(w, r, "login.html", web.Data{
			Judul:     "Masuk",
			Polos:     true,
			Galat:     pesan,
			CSRFToken: token,
		})
		return
	}

	// Login berhasil — reset rate limit.
	resetRateLimit(username)

	sessionID, err := BuatSesi(pengguna)
	if err != nil {
		log.Printf("login handler: gagal membuat sesi: %v", err)
		token := ambilTokenCSRF(w, r)
		h.renderer.Render(w, r, "login.html", web.Data{
			Judul:     "Masuk",
			Polos:     true,
			Galat:     "Gagal membuat sesi. Coba lagi.",
			CSRFToken: token,
		})
		return
	}

	http.SetCookie(w, buatCookieSesi(sessionID, int(durasisesi.Seconds())))
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Logout menghapus sesi dan mengarahkan ke halaman login.
// Hanya menerima POST. CSRF divalidasi untuk mencegah logout paksa.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err == nil {
		if !periksaCSRF(r) {
			log.Printf("logout handler: CSRF token tidak valid dari %s", r.RemoteAddr)
			http.Error(w, "Permintaan tidak valid.", http.StatusForbidden)
			return
		}
	}
	if cookie, err := r.Cookie(namaCookie); err == nil {
		_ = h.service.Logout(r.Context(), cookie.Value)
	}
	http.SetCookie(w, buatCookieSesi("", -1))
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
