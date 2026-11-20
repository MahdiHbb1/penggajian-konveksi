// Pemilik: Anya | Modul 5 | Aturan: 13, 15
package periode

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"penggajian-konveksi/internal/domain"
	"penggajian-konveksi/internal/web"
)

// Handler menangani halaman periode dan koreksi.
type Handler struct {
	renderer web.Renderer
	service  *Service
}

// NewHandler membuat handler periode.
func NewHandler(r web.Renderer, s *Service) *Handler { return &Handler{renderer: r, service: s} }

// DaftarkanRute mendaftarkan halaman 13, 14, dan 15.
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
	Periode     domain.PeriodeGaji
	Ringkasan   RingkasanTutup
	BisaDitutup bool
}
type koreksiPage struct {
	Koreksi []domain.KoreksiGaji
	Periode []domain.PeriodeGaji
	Pekerja []domain.Pekerja
}

func galatAksi(w http.ResponseWriter, e error, pesan string) {
	switch {
	case errors.Is(e, ErrPeriodeTidakDitemukan):
		http.Error(w, ErrPeriodeTidakDitemukan.Error(), http.StatusNotFound)
	case errors.Is(e, ErrPeriodeSudahDitutup):
		http.Error(w, ErrPeriodeSudahDitutup.Error(), http.StatusConflict)
	case errors.Is(e, ErrDatabasePeriodeTidakTersedia):
		http.Error(w, pesan+": "+e.Error(), http.StatusInternalServerError)
	default:
		http.Error(w, pesan+": "+e.Error(), http.StatusBadRequest)
	}
}

func galatServer(w http.ResponseWriter, e error, pesan string) {
	if errors.Is(e, ErrPeriodeTidakDitemukan) {
		http.Error(w, ErrPeriodeTidakDitemukan.Error(), http.StatusNotFound)
		return
	}
	http.Error(w, pesan+": "+e.Error(), http.StatusInternalServerError)
}

func (h *Handler) Daftar(w http.ResponseWriter, r *http.Request) {
	ps, e := h.service.DaftarPeriode(r.Context())
	if e != nil {
		galatServer(w, e, "Gagal mengambil periode")
		return
	}
	h.renderer.Render(w, r, "periode.html", web.Data{Judul: "Periode Gaji", Aktif: "periode", Isi: periodePage{Periode: ps}})
}

func (h *Handler) FormBaru(w http.ResponseWriter, r *http.Request) { h.Daftar(w, r) }

func (h *Handler) Simpan(w http.ResponseWriter, r *http.Request) {
	if e := r.ParseForm(); e != nil {
		http.Error(w, "Form tidak valid", http.StatusBadRequest)
		return
	}
	tipe := domain.TipePeriode(r.FormValue("tipe"))
	mulai, e := time.Parse("2006-01-02", r.FormValue("tanggal_mulai"))
	if e != nil {
		http.Error(w, "Tanggal mulai tidak valid", http.StatusBadRequest)
		return
	}
	if _, e = h.service.BuatPeriode(r.Context(), tipe, mulai); e != nil {
		galatAksi(w, e, "Gagal membuat periode")
		return
	}
	http.Redirect(w, r, "/periode", http.StatusSeeOther)
}

// Hitung meneruskan pengguna ke halaman penggajian dengan periode yang sudah dipilih.
// Perhitungan gaji sendiri tetap menjadi tanggung jawab Modul 4.
func (h *Handler) Hitung(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if e != nil || id <= 0 {
		http.Error(w, "ID periode tidak valid", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/penggajian?periode_id="+strconv.FormatInt(id, 10), http.StatusSeeOther)
}

func (h *Handler) Tutup(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if e != nil {
		http.Error(w, "ID periode tidak valid", http.StatusBadRequest)
		return
	}
	p, e := h.service.AmbilPeriode(r.Context(), id)
	if e != nil {
		galatServer(w, e, "Gagal mengambil periode")
		return
	}
	ring, e := h.service.RingkasanTutupPeriode(r.Context(), id)
	if e != nil {
		galatServer(w, e, "Gagal mengambil ringkasan")
		return
	}
	h.renderer.Render(w, r, "periode_tutup.html", web.Data{
		Judul: "Tutup Periode",
		Aktif: "periode",
		Isi:   tutupPage{Periode: p, Ringkasan: ring, BisaDitutup: p.Status == domain.PeriodeDihitung},
	})
}

func (h *Handler) SimpanTutup(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("sudah_dibayar") != "1" {
		http.Error(w, "Konfirmasi pembayaran wajib dicentang", http.StatusBadRequest)
		return
	}
	id, e := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if e != nil {
		http.Error(w, "ID periode tidak valid", http.StatusBadRequest)
		return
	}
	if e = h.service.TutupPeriode(r.Context(), id); e != nil {
		galatAksi(w, e, "Gagal menutup periode")
		return
	}
	http.Redirect(w, r, "/periode", http.StatusSeeOther)
}

func (h *Handler) DaftarKoreksi(w http.ResponseWriter, r *http.Request) {
	ks, e := h.service.DaftarKoreksi(r.Context(), nil)
	if e != nil {
		galatServer(w, e, "Gagal mengambil koreksi")
		return
	}
	ps, e := h.service.DaftarPeriode(r.Context())
	if e != nil {
		galatServer(w, e, "Gagal mengambil periode")
		return
	}
	pekerja, e := h.service.DaftarPekerjaKoreksi(r.Context())
	if e != nil {
		galatServer(w, e, "Gagal mengambil pekerja")
		return
	}
	h.renderer.Render(w, r, "koreksi.html", web.Data{
		Judul: "Koreksi Gaji",
		Aktif: "koreksi",
		Isi:   koreksiPage{Koreksi: ks, Periode: ps, Pekerja: pekerja},
	})
}

func (h *Handler) FormKoreksi(w http.ResponseWriter, r *http.Request) { h.DaftarKoreksi(w, r) }

func (h *Handler) SimpanKoreksi(w http.ResponseWriter, r *http.Request) {
	if e := r.ParseForm(); e != nil {
		http.Error(w, "Form tidak valid", http.StatusBadRequest)
		return
	}
	pid, e := strconv.ParseInt(r.FormValue("pekerja_id"), 10, 64)
	if e != nil {
		http.Error(w, "ID pekerja tidak valid", http.StatusBadRequest)
		return
	}
	asal, e := strconv.ParseInt(r.FormValue("periode_asal_id"), 10, 64)
	if e != nil {
		http.Error(w, "ID periode asal tidak valid", http.StatusBadRequest)
		return
	}
	jumlah, e := strconv.ParseInt(r.FormValue("jumlah"), 10, 64)
	if e != nil {
		http.Error(w, "Nominal tidak valid", http.StatusBadRequest)
		return
	}
	_, e = h.service.CatatKoreksi(r.Context(), domain.KoreksiGaji{
		PekerjaID:     pid,
		PeriodeAsalID: asal,
		Jumlah:        domain.Rupiah(jumlah),
		Keterangan:    r.FormValue("keterangan"),
	})
	if e != nil {
		galatAksi(w, e, "Gagal menyimpan koreksi")
		return
	}
	http.Redirect(w, r, "/koreksi", http.StatusSeeOther)
}
