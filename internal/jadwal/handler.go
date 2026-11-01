// Pemilik: Edgar | Modul 2 | Aturan: 3, 19
package jadwal

import (
	"net/http"

	"penggajian-konveksi/internal/web"
)

// Handler menangani halaman jadwal shift.
type Handler struct {
	renderer web.Renderer
	service  *Service
}

// NewHandler membuat handler jadwal.
func NewHandler(renderer web.Renderer, service *Service) *Handler {
	return &Handler{renderer: renderer, service: service}
}

// DaftarkanRute mendaftarkan halaman jadwal dan aksi salin.
func DaftarkanRute(mux *http.ServeMux, renderer web.Renderer, service *Service) {
	h := NewHandler(renderer, service)
	mux.HandleFunc("GET /jadwal", h.Tampilkan)
	mux.HandleFunc("POST /jadwal", h.Simpan)
	mux.HandleFunc("GET /jadwal/salin", h.Salin)
	mux.HandleFunc("POST /jadwal/salin", h.Salin)
}

// Tampilkan menampilkan grid jadwal mingguan.
func (h *Handler) Tampilkan(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "jadwal.html", web.Data{Judul: "Jadwal Shift", Aktif: "jadwal"})
}

// Simpan adalah placeholder aksi simpan jadwal.
func (h *Handler) Simpan(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "jadwal.html", web.Data{
		Judul: "Jadwal Shift",
		Aktif: "jadwal",
		Info:  "Aksi simpan jadwal belum diimplementasikan pada kerangka.",
	})
}

// Salin adalah placeholder aksi salin jadwal pekan lalu.
func (h *Handler) Salin(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "jadwal.html", web.Data{
		Judul: "Jadwal Shift",
		Aktif: "jadwal",
		Info:  "Aksi salin jadwal belum diimplementasikan pada kerangka.",
	})
}
