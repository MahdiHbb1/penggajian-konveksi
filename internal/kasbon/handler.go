// Pemilik: Edgar | Modul 2 | Aturan: 12, 16
package kasbon

import (
	"net/http"

	"penggajian-konveksi/internal/web"
)

// Handler menangani halaman kasbon.
type Handler struct {
	renderer web.Renderer
	service  *Service
}

// NewHandler membuat handler kasbon.
func NewHandler(renderer web.Renderer, service *Service) *Handler {
	return &Handler{renderer: renderer, service: service}
}

// DaftarkanRute mendaftarkan halaman daftar, form, dan detail kasbon.
func DaftarkanRute(mux *http.ServeMux, renderer web.Renderer, service *Service) {
	h := NewHandler(renderer, service)
	mux.HandleFunc("GET /kasbon", h.Daftar)
	mux.HandleFunc("GET /kasbon/baru", h.FormBaru)
	mux.HandleFunc("POST /kasbon/baru", h.Simpan)
	mux.HandleFunc("GET /kasbon/{id}", h.Detail)
}

// Daftar menampilkan tabel kasbon.
func (h *Handler) Daftar(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "kasbon_daftar.html", web.Data{Judul: "Kasbon", Aktif: "kasbon"})
}

// FormBaru menampilkan form kasbon baru.
func (h *Handler) FormBaru(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "kasbon.html", web.Data{Judul: "Kasbon Baru", Aktif: "kasbon"})
}

// Simpan adalah placeholder aksi simpan kasbon.
func (h *Handler) Simpan(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "kasbon.html", web.Data{
		Judul: "Kasbon Baru",
		Aktif: "kasbon",
		Info:  "Aksi simpan kasbon belum diimplementasikan pada kerangka.",
	})
}

// Detail menampilkan kasbon dan riwayat potongannya.
func (h *Handler) Detail(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "kasbon.html", web.Data{Judul: "Detail Kasbon", Aktif: "kasbon"})
}
