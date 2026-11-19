// Pemilik: Fadli | Modul 6 | Aturan: 13, 17
package laporan

import (
	"fmt"
	"net/http"
	"strconv"

	"penggajian-konveksi/internal/domain"
	"penggajian-konveksi/internal/web"
)

type IsiSlip struct {
	Slip    domain.SlipGaji
	Rincian []domain.SlipGajiDetail
}

type IsiRekap struct {
	PeriodeID int64
	Rekap     domain.RekapGaji
	Harian    []domain.SlipGaji
	Bulanan   []domain.SlipGaji
}

type Handler struct {
	renderer web.Renderer
	service  *Service
}

func NewHandler(renderer web.Renderer, service *Service) *Handler {
	return &Handler{renderer: renderer, service: service}
}

func DaftarkanRute(mux *http.ServeMux, renderer web.Renderer, service *Service) {
	h := NewHandler(renderer, service)
	mux.HandleFunc("GET /dashboard", h.Dashboard)
	mux.HandleFunc("GET /slip/{periodeID}/{pekerjaID}", h.Slip)
	mux.HandleFunc("GET /rekap/{periodeID}", h.Rekap)
}

// Dashboard menampilkan halaman ringkasan.
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "dashboard.html", web.Data{
		Judul: "Dashboard",
		Aktif: "dashboard",
	})
}

// Slip menampilkan slip beserta rincian penggajian.
func (h *Handler) Slip(w http.ResponseWriter, r *http.Request) {
	periodeID, err := strconv.ParseInt(r.PathValue("periodeID"), 10, 64)
	if err != nil || periodeID <= 0 {
		h.tampilkanGalatSlip(w, r, "ID periode tidak valid.")
		return
	}

	pekerjaID, err := strconv.ParseInt(r.PathValue("pekerjaID"), 10, 64)
	if err != nil || pekerjaID <= 0 {
		h.tampilkanGalatSlip(w, r, "ID pekerja tidak valid.")
		return
	}

	slip, err := h.service.AmbilSlipGaji(r.Context(), periodeID, pekerjaID)
	if err != nil {
		h.tampilkanGalatSlip(w, r, fmt.Sprintf("Slip tidak ditemukan: %v", err))
		return
	}

	rincian, err := h.service.DaftarRincianSlip(r.Context(), slip.ID)
	if err != nil {
		h.tampilkanGalatSlip(w, r, fmt.Sprintf("Gagal mengambil rincian slip: %v", err))
		return
	}

	h.renderer.Render(w, r, "slip.html", web.Data{
		Judul: "Slip Gaji",
		Aktif: "slip",
		Isi: IsiSlip{
			Slip:    slip,
			Rincian: rincian,
		},
	})
}

func (h *Handler) tampilkanGalatSlip(w http.ResponseWriter, r *http.Request, pesan string) {
	h.renderer.Render(w, r, "slip.html", web.Data{
		Judul: "Slip Gaji",
		Aktif: "slip",
		Galat: pesan,
	})
}

// Rekap menampilkan rekap gaji per periode.
func (h *Handler) Rekap(w http.ResponseWriter, r *http.Request) {
	periodeID, err := strconv.ParseInt(r.PathValue("periodeID"), 10, 64)

	if err != nil || periodeID <= 0 {
		h.renderer.Render(w, r, "rekap.html", web.Data{
			Judul: "Rekap Gaji",
			Aktif: "rekap",
			Isi:   IsiRekap{},
		})
		return
	}

	rekap, err := h.service.RekapGajiPerPeriode(r.Context(), periodeID)
	if err != nil {
		h.renderer.Render(w, r, "rekap.html", web.Data{
			Judul: "Rekap Gaji",
			Aktif: "rekap",
			Galat: fmt.Sprintf("Gagal mengambil rekap: %v", err),
			Isi:   IsiRekap{PeriodeID: periodeID},
		})
		return
	}

	slips, err := h.service.DaftarSlipPeriode(r.Context(), periodeID)
	if err != nil {
		h.renderer.Render(w, r, "rekap.html", web.Data{
			Judul: "Rekap Gaji",
			Aktif: "rekap",
			Galat: fmt.Sprintf("Gagal mengambil daftar slip: %v", err),
			Isi:   IsiRekap{PeriodeID: periodeID, Rekap: rekap},
		})
		return
	}

	var harian, bulanan []domain.SlipGaji

	for _, slip := range slips {
		if slip.Pekerja != nil && slip.Pekerja.Jenis == domain.PekerjaHarian {
			harian = append(harian, slip)
		} else {
			bulanan = append(bulanan, slip)
		}
	}

	h.renderer.Render(w, r, "rekap.html", web.Data{
		Judul: "Rekap Gaji",
		Aktif: "rekap",
		Isi: IsiRekap{
			PeriodeID: periodeID,
			Rekap:     rekap,
			Harian:    harian,
			Bulanan:   bulanan,
		},
	})
}
