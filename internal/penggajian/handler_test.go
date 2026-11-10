// Pemilik: Farah | Modul 4 | Aturan: 17
package penggajian

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"penggajian-konveksi/internal/domain"
	"penggajian-konveksi/internal/web"
)

func newHandlerUji(t *testing.T) (*http.ServeMux, *fakeRepo) {
	t.Helper()
	repo := newFakeRepo()
	renderer := &web.Render{Dir: filepath.Join("..", "..", "web", "templates")}
	mux := http.NewServeMux()
	DaftarkanRute(mux, renderer, serviceUji(repo))
	return mux, repo
}

func mintaHalaman(mux *http.ServeMux, target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rekaman := httptest.NewRecorder()
	mux.ServeHTTP(rekaman, req)
	return rekaman
}

func seedSlip(repo *fakeRepo) {
	periode := domain.PeriodeGaji{
		ID: 1, Tipe: domain.PeriodeMingguan,
		TanggalMulai: tanggal(2026, 11, 2), TanggalSelesai: tanggal(2026, 11, 7),
		Status: domain.PeriodeDihitung,
	}
	pekerja := domain.Pekerja{
		ID: 1, Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, 11, 1), Aktif: true,
	}
	slip := domain.SlipGaji{
		ID: 1, PeriodeID: periode.ID, PekerjaID: pekerja.ID,
		UpahPokok: 600000, TotalLembur: 56250, PotonganTelat: 12500,
		GajiBruto: 643750, PotonganKasbon: 193125, GajiBersih: 450625,
		HariKerjaDibayar: 6, Pekerja: &pekerja, Periode: &periode,
	}
	masuk := time.Date(2026, 11, 2, 7, 0, 0, 0, time.Local)
	keluar := time.Date(2026, 11, 2, 15, 0, 0, 0, time.Local)
	repo.periode = periode
	repo.slips[pekerja.ID] = slip
	repo.rincian[pekerja.ID] = []RincianHarian{{
		SlipGajiDetail: domain.SlipGajiDetail{
			ID: 1, SlipID: slip.ID, AbsensiID: 1, Tanggal: tanggal(2026, 11, 2),
			UpahPokokHari: 100000, UpahPerJam: 12500,
			MenitTelat: 0, PotonganTelat: 0, MenitLembur: 60, UpahLembur: 18750,
		},
		NamaShift: "Pagi",
		JamMasuk:  masuk,
		JamKeluar: &keluar,
	}}
}

func TestHalamanPenggajianMenampilkanHasil(t *testing.T) {
	mux, repo := newHandlerUji(t)
	seedSlip(repo)

	hasil := mintaHalaman(mux, "/penggajian")
	if hasil.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", hasil.Code)
	}
	body := hasil.Body.String()
	for _, teks := range []string{"Budi", "Rp 643.750", "Dihitung", "Lolos", "Hitung gaji"} {
		if !strings.Contains(body, teks) {
			t.Fatalf("halaman penggajian tidak memuat %q", teks)
		}
	}
}

func TestHalamanRincianMenampilkanData(t *testing.T) {
	mux, repo := newHandlerUji(t)
	seedSlip(repo)

	hasil := mintaHalaman(mux, "/penggajian/1/rincian/1")
	if hasil.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", hasil.Code)
	}
	body := hasil.Body.String()
	for _, teks := range []string{"Budi", "P-001", "Rp 600.000", "Pagi", "07:00", "Hitung ulang"} {
		if !strings.Contains(body, teks) {
			t.Fatalf("halaman rincian tidak memuat %q", teks)
		}
	}
}

func TestHalamanPenggajianTanpaPeriode(t *testing.T) {
	mux, _ := newHandlerUji(t)

	hasil := mintaHalaman(mux, "/penggajian")
	if hasil.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", hasil.Code)
	}
	body := hasil.Body.String()
	for _, teks := range []string{"Proses Hitung Gaji", "Belum ada periode gaji"} {
		if !strings.Contains(body, teks) {
			t.Fatalf("halaman penggajian tidak memuat %q", teks)
		}
	}
}

func TestHitungGajiMenampilkanGalatPrasyarat(t *testing.T) {
	mux, repo := newHandlerUji(t)
	repo.periode = domain.PeriodeGaji{
		ID: 1, Tipe: domain.PeriodeMingguan,
		TanggalMulai: tanggal(2026, 11, 2), TanggalSelesai: tanggal(2026, 11, 7),
		Status: domain.PeriodeDitutup,
	}

	req := httptest.NewRequest(http.MethodPost, "/penggajian", strings.NewReader("periode_id=1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rekaman := httptest.NewRecorder()
	mux.ServeHTTP(rekaman, req)

	if rekaman.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", rekaman.Code)
	}
	if !strings.Contains(rekaman.Body.String(), "Periode yang sudah ditutup tidak dapat dihitung.") {
		t.Fatal("halaman tidak menampilkan pesan periode tertutup")
	}
}

func TestHalamanRincianTidakDitemukan(t *testing.T) {
	mux, _ := newHandlerUji(t)
	hasil := mintaHalaman(mux, "/penggajian/1/rincian/99")
	if hasil.Code != http.StatusNotFound {
		t.Fatalf("status = %d, ingin 404", hasil.Code)
	}
}
