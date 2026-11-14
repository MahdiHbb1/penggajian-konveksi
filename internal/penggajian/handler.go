// Pemilik: Farah | Modul 4 | Aturan: 2, 5, 6, 7, 8, 9, 12
package penggajian

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"penggajian-konveksi/internal/domain"
	"penggajian-konveksi/internal/web"
)

const galatServer = "Terjadi kesalahan pada server. Coba lagi atau hubungi pengelola."

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

// ---- Model tampilan ----

type opsiPeriode struct {
	ID       int64
	Label    string
	Terpilih bool
}

type prasyaratView struct {
	StatusOK       bool
	SebelumnyaOK   bool
	AbsensiLengkap bool
	JumlahKurang   int
	Pesan          string
}

type barisSlipView struct {
	PekerjaID      int64
	Kode           string
	Nama           string
	Jenis          string
	HariKerja      string
	UpahPokok      string
	Lembur         string
	PotonganTelat  string
	Koreksi        string
	Bruto          string
	PotonganKasbon string
	Bersih         string
}

type halamanPenggajian struct {
	DaftarPeriode []opsiPeriode
	PeriodeID     int64
	AdaPeriode    bool
	PeriodeLabel  string
	TipeLabel     string
	Rentang       string
	Status        string
	Prasyarat     prasyaratView
	Slip          []barisSlipView
	JumlahSlip    int
	TotalPokok    string
	TotalLembur   string
	TotalTelat    string
	TotalKoreksi  string
	TotalBruto    string
	TotalKasbon   string
	TotalBersih   string
}

type rincianHarianView struct {
	Tanggal       string
	Shift         string
	Masuk         string
	Keluar        string
	MenitTelat    int
	PotonganTelat string
	MenitLembur   int
	UpahLembur    string
	UpahPokok     string
}

type halamanRincian struct {
	PeriodeID          int64
	PekerjaID          int64
	Kode               string
	Nama               string
	Jenis              string
	Bulanan            bool
	PeriodeLabel       string
	Status             string
	UpahPokok          string
	UpahPerJam         string
	HariDibayar        int
	HariPeriode        int
	TotalLembur        string
	TotalTelat         string
	TotalKoreksi       string
	GajiBruto          string
	PotonganKasbon     string
	GajiBersih         string
	SisaKasbon         string
	Rincian            []rincianHarianView
	TotalUpahLembur    string
	TotalUpahPokok     string
	TotalPotonganTelat string
}

// ---- Handler ----

// Tampilkan menampilkan halaman proses hitung gaji.
func (h *Handler) Tampilkan(w http.ResponseWriter, r *http.Request) {
	periodeID := parseID(r.URL.Query().Get("periode_id"))
	info := ""
	if r.URL.Query().Get("hasil") == "1" {
		info = "Gaji periode berhasil dihitung."
	}
	h.tampilkan(w, r, periodeID, "", info)
}

// Hitung memproses aksi hitung gaji satu periode.
func (h *Handler) Hitung(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Form tidak valid.", http.StatusBadRequest)
		return
	}
	periodeID := parseID(r.PostFormValue("periode_id"))
	if periodeID <= 0 {
		h.tampilkan(w, r, 0, "Periode wajib dipilih.", "")
		return
	}

	if err := h.service.HitungGajiPeriode(r.Context(), periodeID); err != nil {
		pesan := PesanGalatHitung(err)
		if pesan == "" {
			log.Printf("penggajian: hitung periode %d: %v", periodeID, err)
			pesan = galatServer
		}
		h.tampilkan(w, r, periodeID, pesan, "")
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/penggajian?periode_id=%d&hasil=1", periodeID), http.StatusSeeOther)
}

// Rincian menampilkan rincian hasil hitung satu pekerja.
func (h *Handler) Rincian(w http.ResponseWriter, r *http.Request) {
	periodeID := parseID(r.PathValue("periodeID"))
	pekerjaID := parseID(r.PathValue("pekerjaID"))
	if periodeID <= 0 || pekerjaID <= 0 {
		http.NotFound(w, r)
		return
	}
	info := ""
	if r.URL.Query().Get("hasil") == "1" {
		info = "Gaji periode berhasil dihitung ulang."
	}
	h.tampilkanRincian(w, r, periodeID, pekerjaID, "", info)
}

// HitungUlang memproses hitung ulang satu periode.
func (h *Handler) HitungUlang(w http.ResponseWriter, r *http.Request) {
	periodeID := parseID(r.PathValue("periodeID"))
	if periodeID <= 0 {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Form tidak valid.", http.StatusBadRequest)
		return
	}
	pekerjaID := parseID(r.PostFormValue("pekerja_id"))

	if err := h.service.HitungGajiPeriode(r.Context(), periodeID); err != nil {
		pesan := PesanGalatHitung(err)
		if pesan == "" {
			log.Printf("penggajian: hitung ulang periode %d: %v", periodeID, err)
			pesan = galatServer
		}
		if pekerjaID > 0 {
			h.tampilkanRincian(w, r, periodeID, pekerjaID, pesan, "")
		} else {
			h.tampilkan(w, r, periodeID, pesan, "")
		}
		return
	}

	if pekerjaID > 0 {
		http.Redirect(w, r, fmt.Sprintf(
			"/penggajian/%d/rincian/%d?hasil=1", periodeID, pekerjaID), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/penggajian?periode_id=%d&hasil=1", periodeID), http.StatusSeeOther)
}

func (h *Handler) tampilkan(w http.ResponseWriter, r *http.Request, periodeID int64, galat, info string) {
	data := web.Data{Judul: "Proses Hitung Gaji", Aktif: "penggajian", Galat: galat, Info: info}
	hal := halamanPenggajian{PeriodeID: periodeID}

	daftar, err := h.service.DaftarPeriode(r.Context())
	if err != nil {
		log.Printf("penggajian: daftar periode: %v", err)
		if data.Galat == "" {
			data.Galat = galatServer
		}
	} else {
		for _, p := range daftar {
			hal.DaftarPeriode = append(hal.DaftarPeriode, opsiPeriode{
				ID:    p.ID,
				Label: labelPeriode(p),
			})
		}
		if hal.PeriodeID == 0 && len(daftar) > 0 {
			hal.PeriodeID = daftar[0].ID
		}
	}

	var periode *domain.PeriodeGaji
	if hal.PeriodeID > 0 {
		for i := range daftar {
			if daftar[i].ID == hal.PeriodeID {
				periode = &daftar[i]
				break
			}
		}
	}
	if periode == nil && hal.PeriodeID > 0 {
		if data.Galat == "" {
			data.Galat = "Periode yang dipilih tidak ditemukan."
		}
		hal.PeriodeID = 0
	}
	if periode != nil {
		hal.AdaPeriode = true
		hal.PeriodeLabel = labelPeriode(*periode)
		hal.TipeLabel = labelTipe(periode.Tipe)
		hal.Rentang = periode.TanggalMulai.Format("02/01/2006") + " - " +
			periode.TanggalSelesai.Format("02/01/2006")
		hal.Status = labelStatus(periode.Status)

		prasyarat, err := h.service.PeriksaPrasyaratHitung(r.Context(), periode.ID)
		if err != nil {
			log.Printf("penggajian: periksa prasyarat periode %d: %v", periode.ID, err)
			if data.Galat == "" {
				data.Galat = galatServer
			}
		} else {
			hal.Prasyarat = prasyaratView{
				StatusOK:       prasyarat.StatusPeriodeOK,
				SebelumnyaOK:   prasyarat.PeriodeSebelumnyaOK,
				AbsensiLengkap: prasyarat.AbsensiLengkap,
				JumlahKurang:   prasyarat.JumlahAbsensiKurang,
				Pesan:          prasyarat.Pesan,
			}
		}

		slip, err := h.service.DaftarSlip(r.Context(), periode.ID)
		if err != nil {
			log.Printf("penggajian: daftar slip periode %d: %v", periode.ID, err)
			if data.Galat == "" {
				data.Galat = galatServer
			}
		} else {
			var totalPokok, totalLembur, totalTelat, totalKoreksi domain.Rupiah
			var totalBruto, totalKasbon, totalBersih domain.Rupiah
			for _, s := range slip {
				baris := barisSlipView{
					PekerjaID:      s.PekerjaID,
					HariKerja:      strconv.Itoa(s.HariKerjaDibayar),
					UpahPokok:      formatRupiah(s.UpahPokok),
					Lembur:         formatRupiah(s.TotalLembur),
					PotonganTelat:  formatRupiah(s.PotonganTelat),
					Koreksi:        formatRupiah(s.TotalKoreksi),
					Bruto:          formatRupiah(s.GajiBruto),
					PotonganKasbon: formatRupiah(s.PotonganKasbon),
					Bersih:         formatRupiah(s.GajiBersih),
				}
				if s.Pekerja != nil {
					baris.Kode = s.Pekerja.Kode
					baris.Nama = s.Pekerja.Nama
					baris.Jenis = labelJenis(s.Pekerja.Jenis)
				}
				hal.Slip = append(hal.Slip, baris)
				totalPokok += s.UpahPokok
				totalLembur += s.TotalLembur
				totalTelat += s.PotonganTelat
				totalKoreksi += s.TotalKoreksi
				totalBruto += s.GajiBruto
				totalKasbon += s.PotonganKasbon
				totalBersih += s.GajiBersih
			}
			hal.JumlahSlip = len(slip)
			hal.TotalPokok = formatRupiah(totalPokok)
			hal.TotalLembur = formatRupiah(totalLembur)
			hal.TotalTelat = formatRupiah(totalTelat)
			hal.TotalKoreksi = formatRupiah(totalKoreksi)
			hal.TotalBruto = formatRupiah(totalBruto)
			hal.TotalKasbon = formatRupiah(totalKasbon)
			hal.TotalBersih = formatRupiah(totalBersih)
		}
	}

	for i := range hal.DaftarPeriode {
		hal.DaftarPeriode[i].Terpilih = hal.DaftarPeriode[i].ID == hal.PeriodeID
	}
	if !hal.AdaPeriode && data.Galat == "" && len(daftar) == 0 {
		data.Info = "Belum ada periode gaji. Buat periode terlebih dahulu di halaman Periode."
	}

	data.Isi = hal
	h.renderer.Render(w, r, "penggajian.html", data)
}

func (h *Handler) tampilkanRincian(w http.ResponseWriter, r *http.Request, periodeID, pekerjaID int64, galat, info string) {
	slip, rincian, err := h.service.AmbilRincianTampil(r.Context(), periodeID, pekerjaID)
	if errors.Is(err, domain.ErrTidakDitemukan) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Printf("penggajian: ambil rincian %d/%d: %v", periodeID, pekerjaID, err)
		http.Error(w, galatServer, http.StatusInternalServerError)
		return
	}

	hal := halamanRincian{
		PeriodeID:      periodeID,
		PekerjaID:      pekerjaID,
		UpahPokok:      formatRupiah(slip.UpahPokok),
		HariDibayar:    slip.HariKerjaDibayar,
		HariPeriode:    slip.HariKerjaPeriode,
		TotalLembur:    formatRupiah(slip.TotalLembur),
		TotalTelat:     formatRupiah(slip.PotonganTelat),
		TotalKoreksi:   formatRupiah(slip.TotalKoreksi),
		GajiBruto:      formatRupiah(slip.GajiBruto),
		PotonganKasbon: formatRupiah(slip.PotonganKasbon),
		GajiBersih:     formatRupiah(slip.GajiBersih),
	}
	if slip.Pekerja != nil {
		hal.Kode = slip.Pekerja.Kode
		hal.Nama = slip.Pekerja.Nama
		hal.Jenis = labelJenis(slip.Pekerja.Jenis)
		hal.Bulanan = slip.Pekerja.Jenis == domain.PekerjaBulanan
	}
	if slip.Periode != nil {
		hal.PeriodeLabel = labelPeriode(*slip.Periode)
		hal.Status = labelStatus(slip.Periode.Status)
	}

	if sisa, err := h.service.SisaKasbon(r.Context(), pekerjaID); err != nil {
		log.Printf("penggajian: sisa kasbon pekerja %d: %v", pekerjaID, err)
		hal.SisaKasbon = "-"
	} else {
		hal.SisaKasbon = formatRupiah(sisa)
	}

	var totalLembur, totalPokok, totalTelat domain.Rupiah
	for _, d := range rincian {
		upahPokok := "-"
		if !hal.Bulanan {
			upahPokok = formatRupiah(d.UpahPokokHari)
		}
		baris := rincianHarianView{
			Tanggal:       d.Tanggal.Format("02/01/2006"),
			Shift:         d.NamaShift,
			Masuk:         d.JamMasuk.Format("02/01 15:04"),
			Keluar:        "-",
			MenitTelat:    d.MenitTelat,
			PotonganTelat: formatRupiah(d.PotonganTelat),
			MenitLembur:   d.MenitLembur,
			UpahLembur:    formatRupiah(d.UpahLembur),
			UpahPokok:     upahPokok,
		}
		if d.JamKeluar != nil {
			baris.Keluar = d.JamKeluar.Format("02/01 15:04")
		}
		if hal.UpahPerJam == "" && d.UpahPerJam > 0 {
			hal.UpahPerJam = formatRupiah(d.UpahPerJam)
		}
		hal.Rincian = append(hal.Rincian, baris)
		totalLembur += d.UpahLembur
		totalPokok += d.UpahPokokHari
		totalTelat += d.PotonganTelat
	}
	if hal.UpahPerJam == "" {
		hal.UpahPerJam = "0"
	}
	hal.TotalUpahLembur = formatRupiah(totalLembur)
	hal.TotalUpahPokok = formatRupiah(totalPokok)
	hal.TotalPotonganTelat = formatRupiah(totalTelat)

	h.renderer.Render(w, r, "penggajian_rincian.html", web.Data{
		Judul: "Rincian Gaji",
		Aktif: "penggajian",
		Galat: galat,
		Info:  info,
		Isi:   hal,
	})
}

// ---- Pembantu ----

func parseID(teks string) int64 {
	id, err := strconv.ParseInt(strings.TrimSpace(teks), 10, 64)
	if err != nil || id <= 0 {
		return 0
	}
	return id
}

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

func labelTipe(tipe domain.TipePeriode) string {
	if tipe == domain.PeriodeBulanan {
		return "Bulanan"
	}
	return "Mingguan"
}

func labelPeriode(p domain.PeriodeGaji) string {
	return labelTipe(p.Tipe) + " " + p.TanggalMulai.Format("02/01/2006") +
		" - " + p.TanggalSelesai.Format("02/01/2006")
}

func labelStatus(status domain.StatusPeriode) string {
	switch status {
	case domain.PeriodeDitutup:
		return "Ditutup"
	case domain.PeriodeDihitung:
		return "Dihitung"
	default:
		return "Terbuka"
	}
}

func labelJenis(jenis domain.JenisPekerja) string {
	if jenis == domain.PekerjaBulanan {
		return "Bulanan"
	}
	return "Harian"
}
