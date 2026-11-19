// Pemilik: Fadli | Modul 6 | Aturan: akses Admin dan Pemilik
package auth

import (
	"errors"
	"net/http"

	"penggajian-konveksi/internal/web"
)

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
func DaftarkanRute(mux *http.ServeMux, renderer web.Renderer, service *Service) {
	h := NewHandler(renderer, service)
	mux.HandleFunc("GET /login", h.Login)
	mux.HandleFunc("POST /login", h.ProsesLogin)
	mux.HandleFunc("POST /logout", h.Logout)
	mux.HandleFunc("GET /logout", h.Logout)
}

// Login menampilkan kartu masuk tanpa navigasi.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	// Bila sudah login, langsung ke dashboard.
	if pengguna := web.PenggunaDariContext(r.Context()); pengguna != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	h.renderer.Render(w, r, "login.html", web.Data{Judul: "Masuk", Polos: true})
}

// ProsesLogin memvalidasi credential dan membuat sesi.
func (h *Handler) ProsesLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.renderer.Render(w, r, "login.html", web.Data{
			Judul: "Masuk",
			Polos: true,
			Galat: "Gagal membaca formulir.",
		})
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

	pengguna, err := h.service.Login(r.Context(), username, password)
	if err != nil {
		pesan := "Terjadi kesalahan sistem."
		if errors.Is(err, ErrCredentialSalah) {
			pesan = "Username atau password salah."
		}
		h.renderer.Render(w, r, "login.html", web.Data{
			Judul: "Masuk",
			Polos: true,
			Galat: pesan,
		})
		return
	}

	sessionID, err := BuatSesi(pengguna)
	if err != nil {
		h.renderer.Render(w, r, "login.html", web.Data{
			Judul: "Masuk",
			Polos: true,
			Galat: "Gagal membuat sesi. Coba lagi.",
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     namaCookie,
		Value:    sessionID,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   int(durasisesi.Seconds()),
	})
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Logout menghapus sesi dan mengarahkan ke halaman login.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(namaCookie); err == nil {
		_ = h.service.Logout(r.Context(), cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     namaCookie,
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
