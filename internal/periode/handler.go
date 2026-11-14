// Pemilik: Anya | Modul 5 | Aturan: 13, 15
package periode

import (
	"net/http"
	"penggajian-konveksi/internal/domain"
	"penggajian-konveksi/internal/web"
	"strconv"
	"time"
)

type Handler struct {
	renderer web.Renderer
	service  *Service
}

func NewHandler(r web.Renderer, s *Service) *Handler { return &Handler{renderer: r, service: s} }
func DaftarkanRute(m *http.ServeMux, r web.Renderer, s *Service) {
	h := NewHandler(r, s)
	m.HandleFunc("GET /periode", h.Daftar)
	m.HandleFunc("GET /periode/baru", h.FormBaru)
	m.HandleFunc("POST /periode/baru", h.Simpan)
	m.HandleFunc("POST /periode/{id}/hitung", h.Hitung)
	m.HandleFunc("GET /periode/{id}/tutup", h.Tutup)
	m.HandleFunc("POST /periode/{id}/tutup", h.SimpanTutup)
	m.HandleFunc("GET /koreksi", h.DaftarKoreksi)
	m.HandleFunc("GET /koreksi/baru", h.FormKoreksi)
	m.HandleFunc("POST /koreksi/baru", h.SimpanKoreksi)
}

type periodePage struct{ Periode []domain.PeriodeGaji }
type tutupPage struct {
	Periode   domain.PeriodeGaji
	Ringkasan RingkasanTutup
}
type koreksiPage struct {
	Koreksi []domain.KoreksiGaji
	Periode []domain.PeriodeGaji
	Pekerja []domain.Pekerja
}

func (h *Handler) Daftar(w http.ResponseWriter, r *http.Request) {
	ps, e := h.service.DaftarPeriode(r.Context())
	if e != nil {
		http.Error(w, "Gagal mengambil periode: "+e.Error(), 500)
		return
	}
	h.renderer.Render(w, r, "periode.html", web.Data{Judul: "Periode Gaji", Aktif: "periode", Isi: periodePage{ps}})
}
func (h *Handler) FormBaru(w http.ResponseWriter, r *http.Request) { h.Daftar(w, r) }
func (h *Handler) Simpan(w http.ResponseWriter, r *http.Request) {
	if e := r.ParseForm(); e != nil {
		http.Error(w, "Form tidak valid", 400)
		return
	}
	t := domain.TipePeriode(r.FormValue("tipe"))
	d, e := time.Parse("2006-01-02", r.FormValue("tanggal_mulai"))
	if e != nil {
		http.Error(w, "Tanggal mulai tidak valid", 400)
		return
	}
	if _, e = h.service.BuatPeriode(r.Context(), t, d); e != nil {
		http.Error(w, e.Error(), 400)
		return
	}
	http.Redirect(w, r, "/periode", http.StatusSeeOther)
}
func (h *Handler) Hitung(w http.ResponseWriter, r *http.Request) {
	_, e := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if e != nil {
		http.Error(w, "ID periode tidak valid", 400)
		return
	}
	http.Redirect(w, r, "/penggajian", http.StatusSeeOther)
}
func (h *Handler) Tutup(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if e != nil {
		http.Error(w, "ID periode tidak valid", 400)
		return
	}
	p, e := h.service.AmbilPeriode(r.Context(), id)
	if e != nil {
		http.Error(w, "Periode tidak ditemukan", 404)
		return
	}
	ring, e := h.service.RingkasanTutupPeriode(r.Context(), id)
	if e != nil {
		http.Error(w, "Gagal mengambil ringkasan: "+e.Error(), 500)
		return
	}
	h.renderer.Render(w, r, "periode_tutup.html", web.Data{Judul: "Tutup Periode", Aktif: "periode", Isi: tutupPage{p, ring}})
}
func (h *Handler) SimpanTutup(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("sudah_dibayar") != "1" {
		http.Error(w, "Konfirmasi pembayaran wajib dicentang", 400)
		return
	}
	id, e := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if e != nil {
		http.Error(w, "ID periode tidak valid", 400)
		return
	}
	if e = h.service.TutupPeriode(r.Context(), id); e != nil {
		http.Error(w, e.Error(), 400)
		return
	}
	http.Redirect(w, r, "/periode", http.StatusSeeOther)
}
func (h *Handler) DaftarKoreksi(w http.ResponseWriter, r *http.Request) {
	ks, e := h.service.DaftarKoreksi(r.Context(), nil)
	if e != nil {
		http.Error(w, "Gagal mengambil koreksi: "+e.Error(), 500)
		return
	}
	ps, e := h.service.DaftarPeriode(r.Context())
	if e != nil {
		http.Error(w, "Gagal mengambil periode: "+e.Error(), 500)
		return
	}
	pekerja, e := h.service.DaftarPekerjaKoreksi(r.Context())
	if e != nil {
		http.Error(w, "Gagal mengambil pekerja: "+e.Error(), 500)
		return
	}
	h.renderer.Render(w, r, "koreksi.html", web.Data{Judul: "Koreksi Gaji", Aktif: "koreksi", Isi: koreksiPage{Koreksi: ks, Periode: ps, Pekerja: pekerja}})
}
func (h *Handler) FormKoreksi(w http.ResponseWriter, r *http.Request) { h.DaftarKoreksi(w, r) }
func (h *Handler) SimpanKoreksi(w http.ResponseWriter, r *http.Request) {
	if e := r.ParseForm(); e != nil {
		http.Error(w, "Form tidak valid", 400)
		return
	}
	pid, e := strconv.ParseInt(r.FormValue("pekerja_id"), 10, 64)
	if e != nil {
		http.Error(w, "ID pekerja tidak valid", 400)
		return
	}
	asal, e := strconv.ParseInt(r.FormValue("periode_asal_id"), 10, 64)
	if e != nil {
		http.Error(w, "ID periode asal tidak valid", 400)
		return
	}
	jumlah, e := strconv.ParseInt(r.FormValue("jumlah"), 10, 64)
	if e != nil {
		http.Error(w, "Nominal tidak valid", 400)
		return
	}
	_, e = h.service.CatatKoreksi(r.Context(), domain.KoreksiGaji{PekerjaID: pid, PeriodeAsalID: asal, Jumlah: domain.Rupiah(jumlah), Keterangan: r.FormValue("keterangan")})
	if e != nil {
		http.Error(w, e.Error(), 400)
		return
	}
	http.Redirect(w, r, "/koreksi", http.StatusSeeOther)
}
