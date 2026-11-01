// Pemilik: Farah | Modul 4 | Aturan: 2, 5, 6, 7, 8, 9, 12
package penggajian

import (
	"net/http"

	"penggajian-konveksi/internal/web"
)

// Handler menangani halaman hitung gaji dan rinciannya.
type Handler struct {
	renderer web.Renderer
	service  *Service
}

// NewHandler membuat handler penggajian.
func NewHandler(renderer web.Renderer, service *Service) *Handler {
	return &Handler{renderer: renderer, service: service}
}

// DaftarkanRute mendaftarkan halaman penggajian.
func DaftarkanRute(mux *http.ServeMux, renderer web.Renderer, service *Service) {
	h := NewHandler(renderer, service)
	mux.HandleFunc("GET /penggajian", h.Tampilkan)
	mux.HandleFunc("POST /penggajian", h.Hitung)
	mux.HandleFunc("GET /penggajian/{periodeID}/rincian/{pekerjaID}", h.Rincian)
	mux.HandleFunc("POST /penggajian/{periodeID}/hitung-ulang", h.HitungUlang)
}

// Tampilkan menampilkan halaman proses hitung gaji.
func (h *Handler) Tampilkan(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "penggajian.html", web.Data{Judul: "Proses Hitung Gaji", Aktif: "penggajian"})
}

// Hitung adalah placeholder aksi hitung gaji.
func (h *Handler) Hitung(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "penggajian.html", web.Data{
		Judul: "Proses Hitung Gaji",
		Aktif: "penggajian",
		Info:  "Aksi hitung gaji belum diimplementasikan pada kerangka.",
	})
}

// Rincian menampilkan rincian hasil hitung satu pekerja.
func (h *Handler) Rincian(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "penggajian_rincian.html", web.Data{Judul: "Rincian Gaji", Aktif: "penggajian"})
}

// HitungUlang adalah placeholder aksi hitung ulang.
func (h *Handler) HitungUlang(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "penggajian_rincian.html", web.Data{
		Judul: "Rincian Gaji",
		Aktif: "penggajian",
		Info:  "Aksi hitung ulang belum diimplementasikan pada kerangka.",
	})
}
