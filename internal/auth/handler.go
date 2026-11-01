// Pemilik: Fadli | Modul 6 | Aturan: akses Admin dan Pemilik
package auth

import (
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
	h.renderer.Render(w, r, "login.html", web.Data{Judul: "Masuk", Polos: true})
}

// ProsesLogin adalah placeholder autentikasi.
func (h *Handler) ProsesLogin(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "login.html", web.Data{
		Judul: "Masuk",
		Polos: true,
		Info:  "Autentikasi belum diimplementasikan pada kerangka.",
	})
}

// Logout mengarahkan pengguna kembali ke halaman login.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
