// Pemilik: Fadli | Modul 6 | Aturan: 13, 17
package laporan

import (
	"log"
	"net/http"
	"strconv"

	"penggajian-konveksi/internal/domain"
	"penggajian-konveksi/internal/web"
)

// MiddlewarePeranFunc adalah tipe middleware peran yang diterima oleh DaftarkanRute.
type MiddlewarePeranFunc func(peran ...domain.Peran) func(http.Handler) http.Handler

// IsiSlip adalah data yang dikirim ke template slip.html.
type IsiSlip struct {
	Slip       domain.SlipGaji
	Rincian    []domain.SlipGajiDetail
	SisaKasbon domain.Rupiah
}

// IsiRekap adalah data yang dikirim ke template rekap.html.
type IsiRekap struct {
	PeriodeID     int64
	Rekap         domain.RekapGaji
	Harian        []domain.SlipGaji
	Bulanan       []domain.SlipGaji
	DaftarPeriode []domain.PeriodeGaji
}

// IsiDashboard adalah data yang dikirim ke template dashboard.html.
type IsiDashboard struct {
	Ringkasan RingkasanDashboard
}

// Handler menangani dashboard, slip, dan rekap.
type Handler struct {
	renderer web.Renderer
	service  *Service
}

// NewHandler membuat handler laporan.
func NewHandler(renderer web.Renderer, service *Service) *Handler {
	return &Handler{renderer: renderer, service: service}
}

// DaftarkanRute mendaftarkan dashboard, slip, dan rekap dengan middleware peran.
//
// Aturan akses:
//   - /dashboard   : Admin dan Pemilik
//   - /slip/...    : Admin dan Pemilik
//   - /rekap       : Admin dan Pemilik (pemilih periode)
//   - /rekap/...   : Admin dan Pemilik (detail rekap)
func DaftarkanRute(mux *http.ServeMux, renderer web.Renderer, service *Service, middlewarePeran MiddlewarePeranFunc) {
	h := NewHandler(renderer, service)

	izinkanDua := middlewarePeran(domain.PeranAdmin, domain.PeranPemilik)

	mux.Handle("GET /dashboard", izinkanDua(http.HandlerFunc(h.Dashboard)))
	mux.Handle("GET /slip/{periodeID}/{pekerjaID}", izinkanDua(http.HandlerFunc(h.Slip)))
	mux.Handle("GET /rekap", izinkanDua(http.HandlerFunc(h.RekapPilihPeriode)))
	mux.Handle("GET /rekap/{periodeID}", izinkanDua(http.HandlerFunc(h.Rekap)))
}

// Dashboard menampilkan halaman ringkasan dengan data nyata dari database.
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	ringkasan := h.service.AmbilRingkasanDashboard(r.Context())
	h.renderer.Render(w, r, "dashboard.html", web.Data{
		Judul: "Dashboard",
		Aktif: "dashboard",
		Isi:   IsiDashboard{Ringkasan: ringkasan},
	})
}

// Slip menampilkan slip beserta rincian penggajian dan sisa kasbon pekerja.
func (h *Handler) Slip(w http.ResponseWriter, r *http.Request) {
	periodeID, err := strconv.ParseInt(r.PathValue("periodeID"), 10, 64)
	if err != nil || periodeID <= 0 {
		h.tampilkanGalatSlip(w, r, "ID periode tidak valid.", http.StatusBadRequest)
		return
	}

	pekerjaID, err := strconv.ParseInt(r.PathValue("pekerjaID"), 10, 64)
	if err != nil || pekerjaID <= 0 {
		h.tampilkanGalatSlip(w, r, "ID pekerja tidak valid.", http.StatusBadRequest)
		return
	}

	slip, err := h.service.AmbilSlipGaji(r.Context(), periodeID, pekerjaID)
	if err != nil {
		log.Printf("slip handler: ambil slip periode=%d pekerja=%d: %v", periodeID, pekerjaID, err)
		h.tampilkanGalatSlip(w, r, "Slip tidak ditemukan atau periode belum siap.", http.StatusNotFound)
		return
	}

	rincian, err := h.service.DaftarRincianSlip(r.Context(), slip.ID)
	if err != nil {
		log.Printf("slip handler: rincian slip=%d: %v", slip.ID, err)
		h.tampilkanGalatSlip(w, r, "Gagal mengambil rincian slip. Silakan coba lagi.", http.StatusInternalServerError)
		return
	}

	sisaKasbon, err := h.service.SisaKasbonPekerja(r.Context(), pekerjaID)
	if err != nil {
		log.Printf("slip handler: sisa kasbon pekerja=%d: %v", pekerjaID, err)
	}

	h.renderer.Render(w, r, "slip.html", web.Data{
		Judul: "Slip Gaji",
		Aktif: "slip",
		Isi: IsiSlip{
			Slip:       slip,
			Rincian:    rincian,
			SisaKasbon: sisaKasbon,
		},
	})
}

func (h *Handler) tampilkanGalatSlip(w http.ResponseWriter, r *http.Request, pesan string, status int) {
	h.renderer.Render(w, r, "slip.html", web.Data{
		Judul:  "Slip Gaji",
		Aktif:  "slip",
		Galat:  pesan,
		Status: status,
	})
}

// RekapPilihPeriode menampilkan halaman rekap dengan pemilih periode.
func (h *Handler) RekapPilihPeriode(w http.ResponseWriter, r *http.Request) {
	periodes, err := h.service.DaftarPeriodeTersedia(r.Context())
	if err != nil {
		log.Printf("rekap handler: daftar periode: %v", err)
	}
	h.renderer.Render(w, r, "rekap.html", web.Data{
		Judul: "Rekap Gaji",
		Aktif: "rekap",
		Isi:   IsiRekap{DaftarPeriode: periodes},
	})
}

// Rekap menampilkan rekap gaji per periode dari snapshot slip yang konsisten.
func (h *Handler) Rekap(w http.ResponseWriter, r *http.Request) {
	periodeID, err := strconv.ParseInt(r.PathValue("periodeID"), 10, 64)
	if err != nil || periodeID <= 0 {
		http.Redirect(w, r, "/rekap", http.StatusSeeOther)
		return
	}

	periodes, err := h.service.DaftarPeriodeTersedia(r.Context())
	if err != nil {
		log.Printf("rekap handler: daftar periode: %v", err)
	}

	rekap, slips, err := h.service.RekapDanDaftarSlip(r.Context(), periodeID)
	if err != nil {
		log.Printf("rekap handler: rekap periode=%d: %v", periodeID, err)
		h.renderer.Render(w, r, "rekap.html", web.Data{
			Judul: "Rekap Gaji",
			Aktif: "rekap",
			Galat: "Gagal mengambil data rekap. Pastikan periode sudah dihitung atau ditutup.",
			Isi:   IsiRekap{PeriodeID: 0, DaftarPeriode: periodes},
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
			PeriodeID:     periodeID,
			Rekap:         rekap,
			Harian:        harian,
			Bulanan:       bulanan,
			DaftarPeriode: periodes,
		},
	})
}
