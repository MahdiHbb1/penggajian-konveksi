// Pemilik: Fadli | Modul 6 | Aturan: 13, 17
package laporan

import (
	"net/http"

	"penggajian-konveksi/internal/web"
)

// Handler menangani dashboard, slip, dan rekap.
type Handler struct {
	renderer web.Renderer
	service  *Service
}

// NewHandler membuat handler laporan.
func NewHandler(renderer web.Renderer, service *Service) *Handler {
	return &Handler{renderer: renderer, service: service}
}

// DaftarkanRute mendaftarkan dashboard, slip, dan rekap.
func DaftarkanRute(mux *http.ServeMux, renderer web.Renderer, service *Service) {
	h := NewHandler(renderer, service)
	mux.HandleFunc("GET /dashboard", h.Dashboard)
	mux.HandleFunc("GET /slip/{periodeID}/{pekerjaID}", h.Slip)
	mux.HandleFunc("GET /rekap/{periodeID}", h.Rekap)
}

// Dashboard menampilkan ringkasan untuk Admin dan Pemilik.
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "dashboard.html", web.Data{Judul: "Dashboard", Aktif: "dashboard"})
}

// Slip menampilkan slip siap cetak.
func (h *Handler) Slip(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "slip.html", web.Data{Judul: "Slip Gaji", Aktif: "slip"})
}

// Rekap menampilkan rekap gaji per periode.
func (h *Handler) Rekap(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "rekap.html", web.Data{Judul: "Rekap Gaji", Aktif: "rekap"})
}
