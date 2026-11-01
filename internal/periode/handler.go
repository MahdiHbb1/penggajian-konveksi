// Pemilik: Anya | Modul 5 | Aturan: 13, 15
package periode

import (
	"net/http"

	"penggajian-konveksi/internal/web"
)

// Handler menangani halaman periode dan koreksi.
type Handler struct {
	renderer web.Renderer
	service  *Service
}

// NewHandler membuat handler periode.
func NewHandler(renderer web.Renderer, service *Service) *Handler {
	return &Handler{renderer: renderer, service: service}
}

// DaftarkanRute mendaftarkan halaman periode, tutup periode, dan koreksi.
func DaftarkanRute(mux *http.ServeMux, renderer web.Renderer, service *Service) {
	h := NewHandler(renderer, service)
	mux.HandleFunc("GET /periode", h.Daftar)
	mux.HandleFunc("GET /periode/baru", h.FormBaru)
	mux.HandleFunc("POST /periode/baru", h.Simpan)
	mux.HandleFunc("POST /periode/{id}/hitung", h.Hitung)
	mux.HandleFunc("GET /periode/{id}/tutup", h.Tutup)
	mux.HandleFunc("POST /periode/{id}/tutup", h.SimpanTutup)
	mux.HandleFunc("GET /koreksi", h.DaftarKoreksi)
	mux.HandleFunc("GET /koreksi/baru", h.FormKoreksi)
	mux.HandleFunc("POST /koreksi/baru", h.SimpanKoreksi)
}

// Daftar menampilkan tabel periode dan form samping.
func (h *Handler) Daftar(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "periode.html", web.Data{Judul: "Periode Gaji", Aktif: "periode"})
}

// FormBaru menampilkan form periode baru.
func (h *Handler) FormBaru(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "periode.html", web.Data{Judul: "Periode Baru", Aktif: "periode"})
}

// Simpan adalah placeholder aksi buat periode.
func (h *Handler) Simpan(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "periode.html", web.Data{
		Judul: "Periode Gaji",
		Aktif: "periode",
		Info:  "Aksi buat periode belum diimplementasikan pada kerangka.",
	})
}

// Hitung adalah placeholder aksi hitung dari daftar periode.
func (h *Handler) Hitung(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/penggajian", http.StatusSeeOther)
}

// Tutup menampilkan konfirmasi tutup periode.
func (h *Handler) Tutup(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "periode_tutup.html", web.Data{Judul: "Tutup Periode", Aktif: "periode"})
}

// SimpanTutup adalah placeholder aksi tutup periode.
func (h *Handler) SimpanTutup(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "periode_tutup.html", web.Data{
		Judul: "Tutup Periode",
		Aktif: "periode",
		Info:  "Aksi tutup periode belum diimplementasikan pada kerangka.",
	})
}

// DaftarKoreksi menampilkan tabel koreksi.
func (h *Handler) DaftarKoreksi(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "koreksi.html", web.Data{Judul: "Koreksi Gaji", Aktif: "koreksi"})
}

// FormKoreksi menampilkan form koreksi baru.
func (h *Handler) FormKoreksi(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "koreksi.html", web.Data{Judul: "Koreksi Baru", Aktif: "koreksi"})
}

// SimpanKoreksi adalah placeholder aksi simpan koreksi.
func (h *Handler) SimpanKoreksi(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "koreksi.html", web.Data{
		Judul: "Koreksi Gaji",
		Aktif: "koreksi",
		Info:  "Aksi simpan koreksi belum diimplementasikan pada kerangka.",
	})
}
