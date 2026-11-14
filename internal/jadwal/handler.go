// Pemilik: Edgar | Modul 2 | Aturan: 3, 19
package jadwal

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"penggajian-konveksi/internal/domain"
	"penggajian-konveksi/internal/web"
)

const (
	layoutTanggal = "2006-01-02"
	galatServer   = "Terjadi kesalahan pada server. Coba lagi atau hubungi pengelola."
)

var namaHari = []string{"Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}

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

// ---- Model tampilan ----

type hariView struct {
	Nama    string
	Tanggal string
}

type selView struct {
	Nama     string // nama field form
	ShiftID  int64
	Terkunci bool
}

type barisView struct {
	Pekerja domain.Pekerja
	Sel     []selView
}

type halamanJadwal struct {
	Mulai      string
	Rentang    string
	Sebelumnya string
	Berikutnya string
	Hari       []hariView
	Shift      []domain.Shift
	Baris      []barisView
}

func namaField(pekerjaID int64, tanggal time.Time) string {
	return "sel_" + kunciSel(pekerjaID, tanggal)
}

// pekanDefault adalah pekan yang ditampilkan bila tidak ada pilihan:
// pekan ini, atau pekan depan bila hari ini Minggu.
func pekanDefault(sekarang time.Time) time.Time {
	if sekarang.Weekday() == time.Sunday {
		sekarang = sekarang.AddDate(0, 0, 1)
	}
	return SeninPekan(sekarang)
}

func parseMulai(s string, sekarang time.Time) time.Time {
	if s == "" {
		return pekanDefault(sekarang)
	}
	t, err := time.ParseInLocation(layoutTanggal, s, time.Local)
	if err != nil {
		return pekanDefault(sekarang)
	}
	return SeninPekan(t)
}

func (h *Handler) susunHalaman(r *http.Request, mulai time.Time) (halamanJadwal, error) {
	hal := halamanJadwal{
		Mulai:      mulai.Format(layoutTanggal),
		Rentang:    mulai.Format("02/01/2006") + " - " + mulai.AddDate(0, 0, 5).Format("02/01/2006"),
		Sebelumnya: mulai.AddDate(0, 0, -7).Format(layoutTanggal),
		Berikutnya: mulai.AddDate(0, 0, 7).Format(layoutTanggal),
	}
	for i, nama := range namaHari {
		hal.Hari = append(hal.Hari, hariView{Nama: nama, Tanggal: mulai.AddDate(0, 0, i).Format("02/01")})
	}

	ctx := r.Context()
	shift, err := h.service.DaftarShift(ctx)
	if err != nil {
		return hal, err
	}
	hal.Shift = shift

	pekerja, err := h.service.PekerjaPekan(ctx, mulai)
	if err != nil {
		return hal, err
	}
	sel, err := h.service.JadwalPekan(ctx, mulai)
	if err != nil {
		return hal, err
	}
	peta := make(map[string]SelJadwal, len(sel))
	for _, c := range sel {
		peta[kunciSel(c.PekerjaID, c.Tanggal)] = c
	}

	var kunciTanggal [6]bool
	for i := range kunciTanggal {
		kunciTanggal[i], err = h.service.Terkunci(ctx, mulai.AddDate(0, 0, i))
		if err != nil {
			return hal, err
		}
	}

	for _, p := range pekerja {
		baris := barisView{Pekerja: p}
		for i := 0; i < 6; i++ {
			tgl := mulai.AddDate(0, 0, i)
			v := selView{Nama: namaField(p.ID, tgl), Terkunci: kunciTanggal[i]}
			if c, ada := peta[kunciSel(p.ID, tgl)]; ada {
				v.ShiftID = c.ShiftID
				if c.PunyaAbsensi {
					v.Terkunci = true
				}
			}
			baris.Sel = append(baris.Sel, v)
		}
		hal.Baris = append(hal.Baris, baris)
	}
	return hal, nil
}

func (h *Handler) tampilkan(w http.ResponseWriter, r *http.Request, mulai time.Time, galat, info string) {
	hal, err := h.susunHalaman(r, mulai)
	if err != nil {
		log.Printf("jadwal: susun halaman: %v", err)
		if galat == "" {
			galat = galatServer
		}
	}
	h.renderer.Render(w, r, "jadwal.html", web.Data{
		Judul: "Jadwal Shift",
		Aktif: "jadwal",
		Galat: galat,
		Info:  info,
		Isi:   hal,
	})
}

// Tampilkan menampilkan grid jadwal mingguan.
func (h *Handler) Tampilkan(w http.ResponseWriter, r *http.Request) {
	mulai := parseMulai(r.URL.Query().Get("mulai"), time.Now())
	h.tampilkan(w, r, mulai, "", "")
}

// Simpan menyimpan perubahan sel pada grid pekan terpilih.
func (h *Handler) Simpan(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Form tidak valid.", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	mulai := parseMulai(r.PostFormValue("mulai"), time.Now())

	pekerja, err := h.service.PekerjaPekan(ctx, mulai)
	if err != nil {
		log.Printf("jadwal: pekerja pekan: %v", err)
		h.tampilkan(w, r, mulai, galatServer, "")
		return
	}
	sel, err := h.service.JadwalPekan(ctx, mulai)
	if err != nil {
		log.Printf("jadwal: jadwal pekan: %v", err)
		h.tampilkan(w, r, mulai, galatServer, "")
		return
	}
	ada := make(map[string]SelJadwal, len(sel))
	for _, c := range sel {
		ada[kunciSel(c.PekerjaID, c.Tanggal)] = c
	}

	disimpan := 0
	var galat []string
	catatGalat := func(p domain.Pekerja, tgl time.Time, err error) {
		pesan := PesanGalatInput(err)
		if pesan == "" {
			log.Printf("jadwal: simpan sel %d %s: %v", p.ID, tgl.Format(layoutTanggal), err)
			pesan = galatServer
		}
		galat = append(galat, fmt.Sprintf("%s, %s: %s", p.Nama, tgl.Format("02/01"), pesan))
	}

	for _, p := range pekerja {
		for i := 0; i < 6; i++ {
			tgl := mulai.AddDate(0, 0, i)
			nilai, dikirim := r.PostForm[namaField(p.ID, tgl)]
			if !dikirim { // sel terkunci tidak dikirim browser
				continue
			}
			v := strings.TrimSpace(nilai[0])
			cur, sudah := ada[kunciSel(p.ID, tgl)]

			switch {
			case v == "" && !sudah:
				continue
			case v == "":
				if err := h.service.BatalkanJadwal(ctx, cur.ID); err != nil {
					catatGalat(p, tgl, err)
					continue
				}
			default:
				shiftID, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					catatGalat(p, tgl, ErrShiftTidakValid)
					continue
				}
				if sudah {
					if cur.ShiftID == shiftID {
						continue
					}
					err = h.service.UbahJadwal(ctx, domain.JadwalShift{ID: cur.ID, ShiftID: shiftID})
				} else {
					_, err = h.service.BuatJadwal(ctx, domain.JadwalShift{PekerjaID: p.ID, ShiftID: shiftID, Tanggal: tgl})
				}
				if err != nil {
					catatGalat(p, tgl, err)
					continue
				}
			}
			disimpan++
		}
	}

	info := ""
	if len(galat) == 0 {
		info = fmt.Sprintf("Jadwal disimpan (%d sel diubah).", disimpan)
	} else if disimpan > 0 {
		info = fmt.Sprintf("%d sel berhasil disimpan.", disimpan)
	}
	h.tampilkan(w, r, mulai, strings.Join(galat, " | "), info)
}

// Salin menyalin jadwal pekan sebelumnya ke pekan terpilih.
func (h *Handler) Salin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/jadwal", http.StatusSeeOther)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Form tidak valid.", http.StatusBadRequest)
		return
	}
	mulai := parseMulai(r.PostFormValue("mulai"), time.Now())

	disalin, dilewati, err := h.service.SalinJadwalMingguLalu(r.Context(), mulai)
	if err != nil {
		pesan := PesanGalatInput(err)
		if pesan == "" {
			log.Printf("jadwal: salin: %v", err)
			pesan = galatServer
		}
		h.tampilkan(w, r, mulai, pesan, "")
		return
	}
	h.tampilkan(w, r, mulai, "",
		fmt.Sprintf("Salin selesai: %d sel disalin, %d sel dilewati.", disalin, dilewati))
}