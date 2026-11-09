// Pemilik: Zahra | Modul 3 | Aturan: 2, 3
package absensi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"penggajian-konveksi/internal/domain"
	"penggajian-konveksi/internal/web"
)

type barisHalamanAbsensi struct {
	Jadwal          domain.JadwalShift
	Absensi         domain.Absensi
	AdaAbsensi      bool
	JamMasuk        string
	JamKeluar       string
	JamMulaiShift   string
	JamSelesaiShift string
}

type halamanAbsensi struct {
	Tanggal string
	Baris   []barisHalamanAbsensi
}

type halamanBelumLengkap struct {
	PeriodeID int64
	Periode   []domain.PeriodeGaji
	Absensi   []domain.AbsensiLengkap
}

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

// Tampilkan menampilkan jadwal dan form absensi pada tanggal terpilih.
func (h *Handler) Tampilkan(w http.ResponseWriter, r *http.Request) {
	tanggal, err := tanggalDariString(r.URL.Query().Get("tanggal"))
	if err != nil {
		http.Error(w, "Format tanggal harus YYYY-MM-DD.", http.StatusBadRequest)
		return
	}
	info := ""
	if r.URL.Query().Get("sukses") == "1" {
		info = "Absensi berhasil disimpan."
	}
	h.renderAbsensi(w, r, tanggal, info, "", nil)
}

// Simpan mencatat absensi baru atau memperbarui absensi yang sudah ada.
func (h *Handler) Simpan(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Form absensi tidak dapat dibaca.", http.StatusBadRequest)
		return
	}
	tanggal, err := tanggalDariString(r.FormValue("tanggal"))
	if err != nil {
		http.Error(w, "Tanggal absensi tidak valid.", http.StatusBadRequest)
		return
	}
	jadwalID, err := strconv.ParseInt(r.FormValue("jadwal_shift_id"), 10, 64)
	if err != nil || jadwalID <= 0 {
		h.renderAbsensi(w, r, tanggal, "", "Jadwal absensi tidak valid.", nil)
		return
	}

	jamMasuk, err := waktuPadaTanggal(tanggal, r.FormValue("jam_masuk"))
	if err != nil {
		h.renderAbsensi(w, r, tanggal, "", "Jam masuk wajib diisi dengan format HH:MM.", map[int64]jamInput{jadwalID: {
			JamMasuk: r.FormValue("jam_masuk"), JamKeluar: r.FormValue("jam_keluar"),
		}})
		return
	}

	var jamKeluar *time.Time
	if strings.TrimSpace(r.FormValue("jam_keluar")) != "" {
		waktuKeluar, err := waktuPadaTanggal(tanggal, r.FormValue("jam_keluar"))
		if err != nil {
			h.renderAbsensi(w, r, tanggal, "", "Format jam keluar harus HH:MM.", nil)
			return
		}
		if waktuKeluar.Before(jamMasuk) {
			waktuKeluar = waktuKeluar.Add(24 * time.Hour)
		}
		jamKeluar = &waktuKeluar
	}

	data := domain.Absensi{
		JadwalShiftID: jadwalID,
		JamMasuk:      jamMasuk,
		JamKeluar:     jamKeluar,
	}
	if idText := r.FormValue("absensi_id"); idText != "" {
		id, parseErr := strconv.ParseInt(idText, 10, 64)
		if parseErr != nil || id <= 0 {
			h.renderAbsensi(w, r, tanggal, "", "ID absensi tidak valid.", nil)
			return
		}
		data.ID = id
		err = h.service.UbahAbsensi(r.Context(), data)
	} else {
		_, err = h.service.CatatAbsensi(r.Context(), data)
	}
	if err != nil {
		h.renderAbsensi(w, r, tanggal, "", pesanKesalahanAbsensi(err), map[int64]jamInput{jadwalID: {
			JamMasuk: r.FormValue("jam_masuk"), JamKeluar: r.FormValue("jam_keluar"),
		}})
		return
	}
	http.Redirect(w, r, "/absensi?tanggal="+tanggal.Format("2006-01-02")+"&sukses=1", http.StatusSeeOther)
}

// BelumLengkap menampilkan absensi tanpa jam keluar berdasarkan periode.
func (h *Handler) BelumLengkap(w http.ResponseWriter, r *http.Request) {
	periode, err := h.service.DaftarPeriode(r.Context())
	if err != nil {
		h.renderer.Render(w, r, "absensi_belum_lengkap.html", web.Data{
			Judul: "Absensi Belum Lengkap", Aktif: "absensi",
			Galat: "Gagal mengambil daftar periode: " + err.Error(),
			Isi:   halamanBelumLengkap{},
		})
		return
	}

	view := halamanBelumLengkap{Periode: periode}
	periodeText := r.URL.Query().Get("periode_id")
	if periodeText != "" {
		id, parseErr := strconv.ParseInt(periodeText, 10, 64)
		if parseErr != nil || id <= 0 {
			h.renderer.Render(w, r, "absensi_belum_lengkap.html", web.Data{
				Judul: "Absensi Belum Lengkap", Aktif: "absensi", Galat: "Pilihan periode tidak valid.",
				Isi: view,
			})
			return
		}
		view.PeriodeID = id
		view.Absensi, err = h.service.DaftarAbsensiBelumLengkap(r.Context(), id)
		if err != nil {
			h.renderer.Render(w, r, "absensi_belum_lengkap.html", web.Data{
				Judul: "Absensi Belum Lengkap", Aktif: "absensi", Galat: "Gagal mengambil absensi: " + err.Error(),
				Isi: view,
			})
			return
		}
	}
	h.renderer.Render(w, r, "absensi_belum_lengkap.html", web.Data{
		Judul: "Absensi Belum Lengkap", Aktif: "absensi", Isi: view,
	})
}

type jamInput struct {
	JamMasuk  string
	JamKeluar string
}

func (h *Handler) renderAbsensi(w http.ResponseWriter, r *http.Request, tanggal time.Time, info, galat string, override map[int64]jamInput) {
	daftar, err := h.service.DaftarJadwalAbsensi(r.Context(), tanggal)
	if err != nil {
		galat = "Gagal mengambil jadwal absensi: " + err.Error()
		daftar = nil
	}

	view := halamanAbsensi{
		Tanggal: tanggal.Format("2006-01-02"),
		Baris:   make([]barisHalamanAbsensi, 0, len(daftar)),
	}
	for _, item := range daftar {
		baris := barisHalamanAbsensi{
			Jadwal:          item.Jadwal,
			Absensi:         item.Absensi,
			AdaAbsensi:      item.Absensi.ID > 0,
			JamMulaiShift:   item.Jadwal.Shift.JamMulai,
			JamSelesaiShift: item.Jadwal.Shift.JamSelesai,
		}
		if item.Absensi.ID > 0 {
			baris.JamMasuk = item.Absensi.JamMasuk.Format("15:04")
			if item.Absensi.JamKeluar != nil {
				baris.JamKeluar = item.Absensi.JamKeluar.Format("15:04")
			}
		}
		if nilai, ok := override[item.Jadwal.ID]; ok {
			baris.JamMasuk = nilai.JamMasuk
			baris.JamKeluar = nilai.JamKeluar
		}
		view.Baris = append(view.Baris, baris)
	}
	h.renderer.Render(w, r, "absensi.html", web.Data{
		Judul: "Input Absensi", Aktif: "absensi", Info: info, Galat: galat, Isi: view,
	})
}

func tanggalDariString(nilai string) (time.Time, error) {
	if strings.TrimSpace(nilai) == "" {
		now := time.Now().In(time.Local)
		y, m, d := now.Date()
		return time.Date(y, m, d, 0, 0, 0, 0, now.Location()), nil
	}
	return time.ParseInLocation("2006-01-02", nilai, time.Local)
}

func waktuPadaTanggal(tanggal time.Time, jam string) (time.Time, error) {
	jam = strings.TrimSpace(jam)
	if jam == "" {
		return time.Time{}, ErrJamMasukTidakValid
	}
	var parsed time.Time
	var err error
	parsed, err = time.ParseInLocation("15:04", jam, tanggal.Location())
	if err != nil {
		parsed, err = time.ParseInLocation("15:04:05", jam, tanggal.Location())
	}
	if err != nil {
		return time.Time{}, err
	}
	j, m, d := tanggal.Date()
	h, menit, detik := parsed.Clock()
	return time.Date(j, m, d, h, menit, detik, 0, tanggal.Location()), nil
}

func pesanKesalahanAbsensi(err error) string {
	switch err {
	case ErrJadwalTidakDitemukan:
		return "Jadwal shift tidak ditemukan."
	case ErrAbsensiGanda:
		return "Absensi untuk jadwal ini sudah pernah dicatat."
	case ErrAbsensiTidakDitemukan:
		return "Data absensi tidak ditemukan."
	case ErrPeriodeTerkunci:
		return "Periode absensi sudah ditutup. Data tidak dapat diubah."
	case ErrJamMasukTidakValid:
		return "Jam masuk wajib diisi."
	case ErrJamKeluarTidakValid:
		return "Jam keluar harus setelah jam masuk."
	case ErrTanggalTidakSesuai:
		return "Tanggal jam masuk harus sesuai dengan tanggal jadwal."
	default:
		return fmt.Sprintf("Absensi gagal disimpan: %v", err)
	}
}
