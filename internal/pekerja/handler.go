// Pemilik: Mahdi | Modul 1 | Aturan: 8, 9, 16
package pekerja

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"penggajian-konveksi/internal/domain"
	"penggajian-konveksi/internal/web"
)

const layoutTanggalInput = "2006-01-02"

var namaBulanPendek = [...]string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}

// BarisPekerja adalah satu baris tabel daftar pekerja.
type BarisPekerja struct {
	ID           int64
	Kode         string
	Nama         string
	JenisLabel   string
	TanggalMulai string
	Status       string
	Aktif        bool
}

// DaftarView adalah data halaman daftar pekerja.
type DaftarView struct {
	Pekerja []BarisPekerja
	Cari    string
	Jenis   domain.JenisPekerja
	Aktif   string
}

// FormView adalah data halaman form pekerja baru atau ubah.
type FormView struct {
	Aksi            string
	Baru            bool
	Kode            string
	Nama            string
	Jenis           domain.JenisPekerja
	TanggalMulai    string
	TanggalBerhenti string
	UpahAwal        domain.Rupiah
}

// BarisUpah adalah satu baris tabel riwayat upah.
type BarisUpah struct {
	BerlakuMulai string
	Nominal      string
	Alasan       string
}

// UpahView adalah data halaman riwayat upah.
type UpahView struct {
	Pekerja      domain.Pekerja
	JenisLabel   string
	TanggalMulai string
	Riwayat      []BarisUpah
	Nominal      domain.Rupiah
	BerlakuMulai string
	Alasan       string
}

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
	tanya := r.URL.Query()
	filter := Filter{Cari: tanya.Get("cari")}

	switch domain.JenisPekerja(strings.ToUpper(strings.TrimSpace(tanya.Get("jenis")))) {
	case domain.PekerjaHarian:
		filter.Jenis = domain.PekerjaHarian
	case domain.PekerjaBulanan:
		filter.Jenis = domain.PekerjaBulanan
	}

	switch strings.TrimSpace(tanya.Get("aktif")) {
	case "1":
		aktif := true
		filter.Aktif = &aktif
	case "0":
		aktif := false
		filter.Aktif = &aktif
	}

	h.renderDaftar(w, r, filter, "")
}

// FormBaru menampilkan form pekerja baru.
func (h *Handler) FormBaru(w http.ResponseWriter, r *http.Request) {
	h.renderForm(w, r, FormView{Aksi: "/pekerja/baru", Baru: true}, "")
}

// SimpanBaru menyimpan pekerja baru beserta upah pertama.
func (h *Handler) SimpanBaru(w http.ResponseWriter, r *http.Request) {
	pekerja, upahAwal, err := bacaFormPekerja(r)
	if err != nil {
		view := formDariMasukan(pekerja, upahAwal, true)
		view.TanggalMulai = strings.TrimSpace(r.PostFormValue("tanggal_mulai"))
		h.renderForm(w, r, view, err.Error())
		return
	}

	if _, err := h.service.TambahPekerja(r.Context(), pekerja, upahAwal); err != nil {
		h.renderForm(w, r, formDariMasukan(pekerja, upahAwal, true), err.Error())
		return
	}

	h.redirectInfo(w, r, "/pekerja", fmt.Sprintf("Pekerja %s berhasil disimpan.", strings.TrimSpace(pekerja.Kode)))
}

// FormUbah menampilkan form ubah pekerja.
func (h *Handler) FormUbah(w http.ResponseWriter, r *http.Request) {
	id, err := idDariPath(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	pekerja, err := h.service.CariPekerja(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	h.renderForm(w, r, formDariPekerja(pekerja), "")
}

// SimpanUbah menyimpan perubahan data pekerja.
func (h *Handler) SimpanUbah(w http.ResponseWriter, r *http.Request) {
	id, err := idDariPath(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	lama, err := h.service.CariPekerja(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	pekerja, _, err := bacaFormPekerja(r)
	pekerja.ID = id
	pekerja.Jenis = lama.Jenis
	if err != nil {
		view := formDariMasukan(pekerja, 0, false)
		view.TanggalMulai = strings.TrimSpace(r.PostFormValue("tanggal_mulai"))
		view.TanggalBerhenti = strings.TrimSpace(r.PostFormValue("tanggal_berhenti"))
		h.renderForm(w, r, view, err.Error())
		return
	}

	berhenti, err := parseTanggalInput(r.PostFormValue("tanggal_berhenti"))
	if err != nil {
		view := formDariMasukan(pekerja, 0, false)
		view.TanggalBerhenti = strings.TrimSpace(r.PostFormValue("tanggal_berhenti"))
		h.renderForm(w, r, view, err.Error())
		return
	}
	if berhenti.IsZero() {
		pekerja.TanggalBerhenti = lama.TanggalBerhenti
	} else {
		pekerja.TanggalBerhenti = &berhenti
	}

	if err := h.service.UbahPekerja(r.Context(), pekerja); err != nil {
		h.renderForm(w, r, formDariMasukan(pekerja, 0, false), err.Error())
		return
	}

	h.redirectInfo(w, r, "/pekerja", fmt.Sprintf("Pekerja %s berhasil diperbarui.", pekerja.Kode))
}

// Nonaktifkan menonaktifkan pekerja tanpa menghapus datanya.
func (h *Handler) Nonaktifkan(w http.ResponseWriter, r *http.Request) {
	id, err := idDariPath(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	tanggal, err := parseTanggalInput(r.PostFormValue("tanggal_berhenti"))
	if err != nil {
		h.renderDaftar(w, r, Filter{}, err.Error())
		return
	}
	if tanggal.IsZero() {
		tanggal = tanggalSaja(time.Now())
	}

	if err := h.service.NonaktifkanPekerja(r.Context(), id, tanggal); err != nil {
		h.renderDaftar(w, r, Filter{}, err.Error())
		return
	}

	h.redirectInfo(w, r, "/pekerja", "Pekerja berhasil dinonaktifkan.")
}

// RiwayatUpah menampilkan riwayat upah pekerja.
func (h *Handler) RiwayatUpah(w http.ResponseWriter, r *http.Request) {
	id, err := idDariPath(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	pekerja, err := h.service.CariPekerja(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	riwayat, err := h.service.DaftarRiwayatUpah(r.Context(), id)
	if err != nil {
		h.renderUpah(w, r, buatUpahView(pekerja, nil), err.Error())
		return
	}
	h.renderUpah(w, r, buatUpahView(pekerja, riwayat), "")
}

// SimpanUpah menambah satu baris riwayat upah.
func (h *Handler) SimpanUpah(w http.ResponseWriter, r *http.Request) {
	id, err := idDariPath(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	pekerja, err := h.service.CariPekerja(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	nominal, errNominal := parseNominalInput(r.PostFormValue("nominal"))
	berlaku, errTanggal := parseTanggalInput(r.PostFormValue("berlaku_mulai"))
	if errNominal != nil || errTanggal != nil {
		view := buatUpahView(pekerja, nil)
		view.Nominal = nominal
		view.BerlakuMulai = strings.TrimSpace(r.PostFormValue("berlaku_mulai"))
		view.Alasan = strings.TrimSpace(r.PostFormValue("alasan"))
		if errNominal != nil {
			h.renderUpah(w, r, view, errNominal.Error())
		} else {
			h.renderUpah(w, r, view, errTanggal.Error())
		}
		return
	}

	upah := domain.Upah{
		PekerjaID:    id,
		Nominal:      nominal,
		BerlakuMulai: berlaku,
		Alasan:       r.PostFormValue("alasan"),
	}
	if err := h.service.TambahRiwayatUpah(r.Context(), upah); err != nil {
		riwayat, _ := h.service.DaftarRiwayatUpah(r.Context(), id)
		view := buatUpahView(pekerja, riwayat)
		view.Nominal = nominal
		view.BerlakuMulai = strings.TrimSpace(r.PostFormValue("berlaku_mulai"))
		view.Alasan = strings.TrimSpace(r.PostFormValue("alasan"))
		h.renderUpah(w, r, view, err.Error())
		return
	}

	h.redirectInfo(w, r, fmt.Sprintf("/pekerja/%d/upah", id), "Upah berhasil disimpan.")
}

func (h *Handler) renderDaftar(w http.ResponseWriter, r *http.Request, filter Filter, galat string) {
	pekerja, err := h.service.DaftarPekerja(r.Context(), filter)
	if err != nil && galat == "" {
		galat = err.Error()
	}

	view := DaftarView{Cari: strings.TrimSpace(filter.Cari), Jenis: filter.Jenis}
	if filter.Aktif != nil {
		if *filter.Aktif {
			view.Aktif = "1"
		} else {
			view.Aktif = "0"
		}
	}
	for _, satu := range pekerja {
		view.Pekerja = append(view.Pekerja, BarisPekerja{
			ID:           satu.ID,
			Kode:         satu.Kode,
			Nama:         satu.Nama,
			JenisLabel:   labelJenis(satu.Jenis),
			TanggalMulai: formatTanggalTampilan(satu.TanggalMulai),
			Status:       satu.Status(),
			Aktif:        satu.Aktif,
		})
	}

	h.renderer.Render(w, r, "pekerja_daftar.html", web.Data{
		Judul: "Pekerja",
		Aktif: "pekerja",
		Galat: galat,
		Info:  r.URL.Query().Get("info"),
		Isi:   view,
	})
}

func (h *Handler) renderForm(w http.ResponseWriter, r *http.Request, view FormView, galat string) {
	judul := "Pekerja Baru"
	if !view.Baru {
		judul = "Ubah Pekerja"
	}
	h.renderer.Render(w, r, "pekerja_form.html", web.Data{
		Judul: judul,
		Aktif: "pekerja",
		Galat: galat,
		Isi:   view,
	})
}

func (h *Handler) renderUpah(w http.ResponseWriter, r *http.Request, view UpahView, galat string) {
	h.renderer.Render(w, r, "upah_riwayat.html", web.Data{
		Judul: "Riwayat Upah",
		Aktif: "pekerja",
		Galat: galat,
		Info:  r.URL.Query().Get("info"),
		Isi:   view,
	})
}

func (h *Handler) redirectInfo(w http.ResponseWriter, r *http.Request, path, pesan string) {
	tujuan := path + "?" + url.Values{"info": {pesan}}.Encode()
	http.Redirect(w, r, tujuan, http.StatusSeeOther)
}

func buatUpahView(pekerja domain.Pekerja, riwayat []domain.Upah) UpahView {
	view := UpahView{
		Pekerja:      pekerja,
		JenisLabel:   labelJenis(pekerja.Jenis),
		TanggalMulai: formatTanggalTampilan(pekerja.TanggalMulai),
	}
	for _, upah := range riwayat {
		view.Riwayat = append(view.Riwayat, BarisUpah{
			BerlakuMulai: formatTanggalTampilan(upah.BerlakuMulai),
			Nominal:      formatRupiah(upah.Nominal),
			Alasan:       upah.Alasan,
		})
	}
	return view
}

func formDariPekerja(pekerja domain.Pekerja) FormView {
	view := formDariMasukan(pekerja, 0, false)
	view.Aksi = fmt.Sprintf("/pekerja/%d/ubah", pekerja.ID)
	return view
}

func formDariMasukan(pekerja domain.Pekerja, upahAwal domain.Rupiah, baru bool) FormView {
	view := FormView{
		Baru:            baru,
		Kode:            strings.TrimSpace(pekerja.Kode),
		Nama:            strings.TrimSpace(pekerja.Nama),
		Jenis:           pekerja.Jenis,
		TanggalMulai:    formatTanggalInput(pekerja.TanggalMulai),
		TanggalBerhenti: formatTanggalInputPtr(pekerja.TanggalBerhenti),
		UpahAwal:        upahAwal,
	}
	if baru {
		view.Aksi = "/pekerja/baru"
	} else {
		view.Aksi = fmt.Sprintf("/pekerja/%d/ubah", pekerja.ID)
	}
	return view
}

func bacaFormPekerja(r *http.Request) (domain.Pekerja, domain.Rupiah, error) {
	if err := r.ParseForm(); err != nil {
		return domain.Pekerja{}, 0, err
	}

	pekerja := domain.Pekerja{
		Kode:  r.PostFormValue("kode"),
		Nama:  r.PostFormValue("nama"),
		Jenis: domain.JenisPekerja(strings.ToUpper(strings.TrimSpace(r.PostFormValue("jenis")))),
	}

	upahAwal, err := parseNominalInput(r.PostFormValue("upah_awal"))
	if err != nil {
		return pekerja, 0, err
	}
	tanggalMulai, err := parseTanggalInput(r.PostFormValue("tanggal_mulai"))
	if err != nil {
		return pekerja, upahAwal, err
	}
	pekerja.TanggalMulai = tanggalMulai
	return pekerja, upahAwal, nil
}

func idDariPath(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("ID pekerja tidak valid.")
	}
	return id, nil
}

func parseNominalInput(teks string) (domain.Rupiah, error) {
	bersih := strings.ReplaceAll(strings.TrimSpace(teks), ".", "")
	bersih = strings.ReplaceAll(bersih, " ", "")
	if bersih == "" {
		return 0, nil
	}
	nilai, err := strconv.ParseInt(bersih, 10, 64)
	if err != nil {
		return 0, errors.New("Nominal upah harus berupa angka.")
	}
	return domain.Rupiah(nilai), nil
}

func parseTanggalInput(teks string) (time.Time, error) {
	teks = strings.TrimSpace(teks)
	if teks == "" {
		return time.Time{}, nil
	}
	tanggal, err := time.ParseInLocation(layoutTanggalInput, teks, time.Local)
	if err != nil {
		return time.Time{}, errors.New("Format tanggal tidak valid.")
	}
	return tanggal, nil
}

func formatTanggalInput(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(layoutTanggalInput)
}

func formatTanggalInputPtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return formatTanggalInput(*t)
}

func formatTanggalTampilan(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return fmt.Sprintf("%02d %s %d", t.Day(), namaBulanPendek[int(t.Month())-1], t.Year())
}

func formatRupiah(nominal domain.Rupiah) string {
	angka := strconv.FormatInt(int64(nominal), 10)
	tanda := ""
	if strings.HasPrefix(angka, "-") {
		tanda = "-"
		angka = strings.TrimPrefix(angka, "-")
	}

	var hasil strings.Builder
	for i := 0; i < len(angka); i++ {
		if i > 0 && (len(angka)-i)%3 == 0 {
			hasil.WriteByte('.')
		}
		hasil.WriteByte(angka[i])
	}
	return "Rp " + tanda + hasil.String()
}

func labelJenis(jenis domain.JenisPekerja) string {
	switch jenis {
	case domain.PekerjaHarian:
		return "Harian"
	case domain.PekerjaBulanan:
		return "Bulanan"
	default:
		return string(jenis)
	}
}
