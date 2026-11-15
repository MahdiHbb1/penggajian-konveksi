// Pemilik: Edgar | Modul 2 | Aturan: 12, 16
package kasbon

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"penggajian-konveksi/internal/domain"
	"penggajian-konveksi/internal/web"
)

const galatServer = "Terjadi kesalahan pada server. Coba lagi atau hubungi pengelola."

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

// ---- Model tampilan (sudah berupa teks supaya template tetap sederhana) ----

type barisDaftar struct {
	ID      int64
	Tanggal string
	Pekerja string
	Jumlah  string
	Sisa    string
	Status  string
}

type halamanDaftar struct {
	Baris           []barisDaftar
	Pekerja         []domain.Pekerja
	PekerjaTerpilih int64
	Status          string
}

type barisPotongan struct {
	Periode  string
	Status   string
	Dipotong string
	Sisa     string
}

type detailTampil struct {
	ID         int64
	Pekerja    string
	Tanggal    string
	Jumlah     string
	Sisa       string
	Status     string
	Keterangan string
	Potongan   []barisPotongan
}

type halamanForm struct {
	Pekerja    []domain.Pekerja
	PekerjaID  int64
	Tanggal    string
	Jumlah     string
	Keterangan string
	Detail     *detailTampil
}

// ---- Handler ----

// Daftar menampilkan tabel kasbon.
func (h *Handler) Daftar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()

	filter := Filter{Status: strings.ToUpper(strings.TrimSpace(q.Get("status")))}
	hal := halamanDaftar{Status: filter.Status}
	if v := q.Get("pekerja_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil && id > 0 {
			filter.PekerjaID = &id
			hal.PekerjaTerpilih = id
		}
	}

	data := web.Data{Judul: "Kasbon", Aktif: "kasbon"}

	if pekerja, err := h.service.DaftarPekerja(ctx, false); err != nil {
		log.Printf("kasbon: daftar pekerja: %v", err)
	} else {
		hal.Pekerja = pekerja
	}

	daftar, err := h.service.DaftarKasbon(ctx, filter)
	if err != nil {
		log.Printf("kasbon: daftar kasbon: %v", err)
		data.Galat = galatServer
	}
	for _, k := range daftar {
		hal.Baris = append(hal.Baris, barisDaftar{
			ID:      k.ID,
			Tanggal: k.Tanggal.Format("02/01/2006"),
			Pekerja: k.NamaPekerja,
			Jumlah:  formatRupiah(k.Jumlah),
			Sisa:    formatRupiah(k.Sisa),
			Status:  k.Status(),
		})
	}

	data.Isi = hal
	h.renderer.Render(w, r, "kasbon_daftar.html", data)
}

// FormBaru menampilkan form kasbon baru.
func (h *Handler) FormBaru(w http.ResponseWriter, r *http.Request) {
	hal := halamanForm{Tanggal: time.Now().Format("2006-01-02")}
	h.tampilkanForm(w, r, "Kasbon Baru", hal, "")
}

// Simpan memproses form kasbon baru.
func (h *Handler) Simpan(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Form tidak valid.", http.StatusBadRequest)
		return
	}

	hal := halamanForm{
		Tanggal:    r.PostFormValue("tanggal"),
		Jumlah:     strings.TrimSpace(r.PostFormValue("jumlah")),
		Keterangan: strings.TrimSpace(r.PostFormValue("keterangan")),
	}
	hal.PekerjaID, _ = strconv.ParseInt(r.PostFormValue("pekerja_id"), 10, 64)

	kasbon := domain.Kasbon{PekerjaID: hal.PekerjaID, Keterangan: hal.Keterangan}

	if t, err := time.ParseInLocation("2006-01-02", hal.Tanggal, time.Local); err == nil {
		kasbon.Tanggal = t
	}
	if jumlah, err := strconv.ParseInt(hal.Jumlah, 10, 64); err == nil {
		kasbon.Jumlah = domain.Rupiah(jumlah)
	} else {
		kasbon.Jumlah = 0 // tidak terbaca dianggap tidak valid
	}

	if _, err := h.service.CatatKasbon(r.Context(), kasbon); err != nil {
		pesan := PesanGalatInput(err)
		if pesan == "" {
			log.Printf("kasbon: catat kasbon: %v", err)
			pesan = galatServer
		}
		h.tampilkanForm(w, r, "Kasbon Baru", hal, pesan)
		return
	}

	http.Redirect(w, r, "/kasbon", http.StatusSeeOther)
}

// Detail menampilkan kasbon dan riwayat potongannya.
func (h *Handler) Detail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.NotFound(w, r)
		return
	}

	ctx := r.Context()
	k, err := h.service.AmbilKasbon(ctx, id)
	if errors.Is(err, ErrKasbonTidakAda) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Printf("kasbon: ambil kasbon %d: %v", id, err)
		h.tampilkanForm(w, r, "Detail Kasbon", halamanForm{}, galatServer)
		return
	}

	d := &detailTampil{
		ID:         k.ID,
		Pekerja:    k.NamaPekerja,
		Tanggal:    k.Tanggal.Format("02/01/2006"),
		Jumlah:     formatRupiah(k.Jumlah),
		Sisa:       formatRupiah(k.Sisa),
		Status:     k.Status(),
		Keterangan: k.Keterangan,
	}

	galat := ""
	potongan, err := h.service.RiwayatPotongan(ctx, id)
	if err != nil {
		log.Printf("kasbon: riwayat potongan %d: %v", id, err)
		galat = galatServer
	}
	// Sisa per baris dihitung berurutan dari jumlah awal.
	sisa := k.Jumlah
	for _, p := range potongan {
		sisa -= p.Jumlah
		d.Potongan = append(d.Potongan, barisPotongan{
			Periode:  labelPeriode(p),
			Status:   labelStatusPeriode(p.StatusPeriode),
			Dipotong: formatRupiah(p.Jumlah),
			Sisa:     formatRupiah(sisa),
		})
	}

	h.tampilkanForm(w, r, "Detail Kasbon", halamanForm{
		Tanggal: time.Now().Format("2006-01-02"),
		Detail:  d,
	}, galat)
}

func (h *Handler) tampilkanForm(w http.ResponseWriter, r *http.Request, judul string, hal halamanForm, galat string) {
	if pekerja, err := h.service.DaftarPekerja(r.Context(), true); err != nil {
		log.Printf("kasbon: daftar pekerja aktif: %v", err)
	} else {
		hal.Pekerja = pekerja
	}
	h.renderer.Render(w, r, "kasbon.html", web.Data{
		Judul: judul,
		Aktif: "kasbon",
		Galat: galat,
		Isi:   hal,
	})
}

// ---- Pembantu format ----

func formatRupiah(n domain.Rupiah) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := strconv.FormatInt(int64(n), 10)
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(c)
	}
	if neg {
		return "-" + b.String()
	}
	return b.String()
}

func labelPeriode(p PotonganTampil) string {
	tipe := "Mingguan"
	if p.Tipe == domain.PeriodeBulanan {
		tipe = "Bulanan"
	}
	return tipe + " " + p.Mulai.Format("02/01/2006") + " - " + p.Selesai.Format("02/01/2006")
}

func labelStatusPeriode(s domain.StatusPeriode) string {
	switch s {
	case domain.PeriodeDitutup:
		return "Sudah ditutup"
	case domain.PeriodeDihitung:
		return "Dihitung (belum ditutup)"
	default:
		return "Terbuka"
	}
}