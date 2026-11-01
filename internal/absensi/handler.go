// Pemilik: Zahra | Modul 3 | Aturan: 2, 3
package absensi

import (
	"net/http"

	"penggajian-konveksi/internal/web"
)

// Handler menangani halaman absensi.
type Handler struct {
	renderer web.Renderer
	service  *Service
}

// NewHandler membuat handler absensi.
func NewHandler(renderer web.Renderer, service *Service) *Handler {
	return &Handler{renderer: renderer, service: service}
}

// DaftarkanRute mendaftarkan halaman input dan daftar absensi belum lengkap.
func DaftarkanRute(mux *http.ServeMux, renderer web.Renderer, service *Service) {
	h := NewHandler(renderer, service)
	mux.HandleFunc("GET /absensi", h.Tampilkan)
	mux.HandleFunc("POST /absensi", h.Simpan)
	mux.HandleFunc("GET /absensi/belum-lengkap", h.BelumLengkap)
}

// Tampilkan menampilkan form input absensi.
func (h *Handler) Tampilkan(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "absensi.html", web.Data{Judul: "Input Absensi", Aktif: "absensi"})
}

// Simpan adalah placeholder aksi simpan absensi.
func (h *Handler) Simpan(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "absensi.html", web.Data{
		Judul: "Input Absensi",
		Aktif: "absensi",
		Info:  "Aksi simpan absensi belum diimplementasikan pada kerangka.",
	})
}

// BelumLengkap menampilkan absensi tanpa jam keluar.
func (h *Handler) BelumLengkap(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "absensi_belum_lengkap.html", web.Data{Judul: "Absensi Belum Lengkap", Aktif: "absensi"})
}
