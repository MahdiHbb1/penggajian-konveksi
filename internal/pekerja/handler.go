// Pemilik: Mahdi | Modul 1 | Aturan: 8, 9, 16
package pekerja

import (
	"net/http"

	"penggajian-konveksi/internal/web"
)

// Handler menangani halaman pekerja dan riwayat upah.
type Handler struct {
	renderer web.Renderer
	service  *Service
}

// NewHandler membuat handler modul pekerja.
func NewHandler(renderer web.Renderer, service *Service) *Handler {
	return &Handler{renderer: renderer, service: service}
}

// DaftarkanRute mendaftarkan empat halaman Modul 1.
func DaftarkanRute(mux *http.ServeMux, renderer web.Renderer, service *Service) {
	h := NewHandler(renderer, service)
	mux.HandleFunc("GET /pekerja", h.Daftar)
	mux.HandleFunc("GET /pekerja/baru", h.FormBaru)
	mux.HandleFunc("POST /pekerja/baru", h.SimpanBaru)
	mux.HandleFunc("GET /pekerja/{id}/ubah", h.FormUbah)
	mux.HandleFunc("POST /pekerja/{id}/ubah", h.SimpanUbah)
	mux.HandleFunc("POST /pekerja/{id}/nonaktif", h.Nonaktifkan)
	mux.HandleFunc("GET /pekerja/{id}/upah", h.RiwayatUpah)
	mux.HandleFunc("POST /pekerja/{id}/upah", h.SimpanUpah)
}

// Daftar menampilkan tabel pekerja.
func (h *Handler) Daftar(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "pekerja_daftar.html", web.Data{Judul: "Pekerja", Aktif: "pekerja"})
}

// FormBaru menampilkan form pekerja baru.
func (h *Handler) FormBaru(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "pekerja_form.html", web.Data{Judul: "Pekerja Baru", Aktif: "pekerja"})
}

// SimpanBaru adalah placeholder aksi simpan pekerja.
func (h *Handler) SimpanBaru(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "pekerja_form.html", web.Data{
		Judul: "Pekerja Baru",
		Aktif: "pekerja",
		Info:  "Aksi simpan belum diimplementasikan pada kerangka.",
	})
}

// FormUbah menampilkan form ubah pekerja.
func (h *Handler) FormUbah(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "pekerja_form.html", web.Data{Judul: "Ubah Pekerja", Aktif: "pekerja"})
}

// SimpanUbah adalah placeholder aksi ubah pekerja.
func (h *Handler) SimpanUbah(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "pekerja_form.html", web.Data{
		Judul: "Ubah Pekerja",
		Aktif: "pekerja",
		Info:  "Aksi ubah belum diimplementasikan pada kerangka.",
	})
}

// Nonaktifkan adalah placeholder aksi nonaktifkan pekerja.
func (h *Handler) Nonaktifkan(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/pekerja", http.StatusSeeOther)
}

// RiwayatUpah menampilkan riwayat upah pekerja.
func (h *Handler) RiwayatUpah(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "upah_riwayat.html", web.Data{Judul: "Riwayat Upah", Aktif: "pekerja"})
}

// SimpanUpah adalah placeholder aksi tambah upah.
func (h *Handler) SimpanUpah(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "upah_riwayat.html", web.Data{
		Judul: "Riwayat Upah",
		Aktif: "pekerja",
		Info:  "Aksi tambah upah belum diimplementasikan pada kerangka.",
	})
}
