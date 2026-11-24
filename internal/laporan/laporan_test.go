// Pemilik: Fadli | Modul 6 | Aturan: 13, 17
package laporan

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"penggajian-konveksi/internal/domain"
	"penggajian-konveksi/internal/web"
)

// mockRepo mengimplementasikan repoAkses untuk pengujian unit dan handler.
type mockRepo struct {
	slip              domain.SlipGaji
	slips             []domain.SlipGaji
	rekap             domain.RekapGaji
	status            domain.StatusPeriode
	rincian           []domain.SlipGajiDetail
	periodes          []domain.PeriodeGaji
	periodesSiap      []domain.PeriodeGaji
	absensiBlmLengkap int
	totalGajiBersih   domain.Rupiah
	totalSisaKasbon   domain.Rupiah
	sisaKasbonPekerja domain.Rupiah
	errSlip           error
	errDaftarSlip     error
	errRekap          error
	errStatus         error
	errRincian        error
	errPeriode        error
	errAbsensi        error
	errTotalGaji      error
	errTotalKasbon    error
	errSisaPekerja    error
}

func (m *mockRepo) AmbilSlip(_ context.Context, _, _ int64) (domain.SlipGaji, error) {
	return m.slip, m.errSlip
}

func (m *mockRepo) DaftarSlip(_ context.Context, _ int64) ([]domain.SlipGaji, error) {
	return m.slips, m.errDaftarSlip
}

func (m *mockRepo) Rekap(_ context.Context, _ int64) (domain.RekapGaji, error) {
	return m.rekap, m.errRekap
}

func (m *mockRepo) AmbilStatusPeriode(_ context.Context, _ int64) (domain.StatusPeriode, error) {
	return m.status, m.errStatus
}

func (m *mockRepo) DaftarRincianSlip(_ context.Context, _ int64) ([]domain.SlipGajiDetail, error) {
	return m.rincian, m.errRincian
}

func (m *mockRepo) DaftarPeriodeTerbaru(_ context.Context, _ int) ([]domain.PeriodeGaji, error) {
	return m.periodes, m.errPeriode
}

func (m *mockRepo) DaftarPeriodeSiapLaporan(_ context.Context, _ int) ([]domain.PeriodeGaji, error) {
	if m.errPeriode != nil {
		return nil, m.errPeriode
	}
	if m.periodesSiap != nil {
		return m.periodesSiap, nil
	}
	var hasil []domain.PeriodeGaji
	for _, p := range m.periodes {
		if p.Status == domain.PeriodeDihitung || p.Status == domain.PeriodeDitutup {
			hasil = append(hasil, p)
		}
	}
	return hasil, nil
}

func (m *mockRepo) HitungAbsensiBelumLengkap(_ context.Context, _ int64) (int, error) {
	return m.absensiBlmLengkap, m.errAbsensi
}

func (m *mockRepo) TotalGajiBersihPeriode(_ context.Context, _ int64) (domain.Rupiah, error) {
	return m.totalGajiBersih, m.errTotalGaji
}

func (m *mockRepo) TotalSisaKasbon(_ context.Context) (domain.Rupiah, error) {
	return m.totalSisaKasbon, m.errTotalKasbon
}

func (m *mockRepo) SisaKasbonPekerja(_ context.Context, _ int64) (domain.Rupiah, error) {
	return m.sisaKasbonPekerja, m.errSisaPekerja
}

func newTestService(repo *mockRepo) *Service {
	return &Service{repo: repo}
}

// --- AmbilSlipGaji ---

func TestAmbilSlipGaji_PeriodeTerbukaDitolak(t *testing.T) {
	periode := domain.PeriodeGaji{ID: 1, Status: domain.PeriodeTerbuka}
	repo := &mockRepo{
		slip: domain.SlipGaji{ID: 1, PeriodeID: 1, PekerjaID: 1, Periode: &periode},
	}
	svc := newTestService(repo)

	_, err := svc.AmbilSlipGaji(context.Background(), 1, 1)
	if !errors.Is(err, ErrPeriodeBelumDihitung) {
		t.Fatalf("AmbilSlipGaji(TERBUKA): ingin ErrPeriodeBelumDihitung, dapat %v", err)
	}
}

func TestAmbilSlipGaji_PeriodeNilDitolakFailClosed(t *testing.T) {
	repo := &mockRepo{
		slip: domain.SlipGaji{ID: 1, PeriodeID: 1, PekerjaID: 1, Periode: nil},
	}
	svc := newTestService(repo)

	_, err := svc.AmbilSlipGaji(context.Background(), 1, 1)
	if !errors.Is(err, ErrPeriodeBelumDihitung) {
		t.Fatalf("AmbilSlipGaji(Periode nil): ingin ErrPeriodeBelumDihitung (fail-closed), dapat %v", err)
	}
}

func TestAmbilSlipGaji_PeriodeDihitungDiizinkan(t *testing.T) {
	periode := domain.PeriodeGaji{ID: 1, Status: domain.PeriodeDihitung}
	repo := &mockRepo{
		slip: domain.SlipGaji{ID: 1, PeriodeID: 1, PekerjaID: 1, Periode: &periode, GajiBersih: 500_000},
	}
	svc := newTestService(repo)

	slip, err := svc.AmbilSlipGaji(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("AmbilSlipGaji(DIHITUNG): error tidak diharapkan: %v", err)
	}
	if slip.GajiBersih != 500_000 {
		t.Fatalf("AmbilSlipGaji(DIHITUNG): gaji bersih = %d, ingin 500000", slip.GajiBersih)
	}
}

func TestAmbilSlipGaji_PeriodeDitutupDiizinkan(t *testing.T) {
	periode := domain.PeriodeGaji{ID: 1, Status: domain.PeriodeDitutup}
	repo := &mockRepo{
		slip: domain.SlipGaji{ID: 1, PeriodeID: 1, PekerjaID: 1, Periode: &periode, GajiBersih: 750_000},
	}
	svc := newTestService(repo)

	slip, err := svc.AmbilSlipGaji(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("AmbilSlipGaji(DITUTUP): error tidak diharapkan: %v", err)
	}
	if slip.GajiBersih != 750_000 {
		t.Fatalf("AmbilSlipGaji(DITUTUP): gaji bersih = %d, ingin 750000", slip.GajiBersih)
	}
}

func TestAmbilSlipGaji_RepoError(t *testing.T) {
	repoErr := errors.New("koneksi terputus")
	repo := &mockRepo{errSlip: repoErr}
	svc := newTestService(repo)

	_, err := svc.AmbilSlipGaji(context.Background(), 1, 1)
	if !errors.Is(err, repoErr) {
		t.Fatalf("AmbilSlipGaji(repo error): ingin %v, dapat %v", repoErr, err)
	}
}

// --- DaftarSlipPeriode & RekapDanDaftarSlip ---

func TestDaftarSlipPeriode_PeriodeTerbukaDitolak(t *testing.T) {
	repo := &mockRepo{status: domain.PeriodeTerbuka}
	svc := newTestService(repo)

	_, err := svc.DaftarSlipPeriode(context.Background(), 1)
	if !errors.Is(err, ErrPeriodeBelumDihitung) {
		t.Fatalf("DaftarSlipPeriode(TERBUKA): ingin ErrPeriodeBelumDihitung, dapat %v", err)
	}
}

func TestRekapDanDaftarSlip_KonsistenDanTidakPotongTelatDuaKali(t *testing.T) {
	repo := &mockRepo{
		status: domain.PeriodeDihitung,
		slips: []domain.SlipGaji{
			{ID: 1, UpahPokok: 600_000, TotalLembur: 31_250, PotonganTelat: 12_500, GajiBruto: 618_750, PotonganKasbon: 185_625, GajiBersih: 433_125},
			{ID: 2, UpahPokok: 500_000, TotalLembur: 0, PotonganTelat: 0, GajiBruto: 500_000, PotonganKasbon: 0, GajiBersih: 500_000},
		},
	}
	svc := newTestService(repo)

	rekap, slips, err := svc.RekapDanDaftarSlip(context.Background(), 1)
	if err != nil {
		t.Fatalf("RekapDanDaftarSlip: %v", err)
	}
	if len(slips) != 2 || rekap.JumlahSlip != 2 {
		t.Fatalf("jumlah slip = %d / %d, ingin 2", len(slips), rekap.JumlahSlip)
	}
	if rekap.TotalBruto != 1_118_750 || rekap.TotalBersih != 933_125 || rekap.TotalPotonganKasbon != 185_625 {
		t.Fatalf("total rekap tidak konsisten: %+v", rekap)
	}
	// Pastikan keterlambatan tidak dipotong dua kali: TotalBersih == TotalBruto - TotalPotonganKasbon.
	if rekap.TotalBersih != rekap.TotalBruto-rekap.TotalPotonganKasbon {
		t.Fatalf("bersih (%d) != bruto (%d) - kasbon (%d)", rekap.TotalBersih, rekap.TotalBruto, rekap.TotalPotonganKasbon)
	}
}

func TestRekapGajiPerPeriode_PeriodeTerbukaDitolak(t *testing.T) {
	repo := &mockRepo{status: domain.PeriodeTerbuka}
	svc := newTestService(repo)

	_, err := svc.RekapGajiPerPeriode(context.Background(), 1)
	if !errors.Is(err, ErrPeriodeBelumDihitung) {
		t.Fatalf("RekapGaji(TERBUKA): ingin ErrPeriodeBelumDihitung, dapat %v", err)
	}
}

// --- DaftarPeriodeTersedia & Dashboard ---

func TestDaftarPeriodeTersedia_FilterBenar(t *testing.T) {
	repo := &mockRepo{
		periodes: []domain.PeriodeGaji{
			{ID: 1, Status: domain.PeriodeTerbuka},
			{ID: 2, Status: domain.PeriodeDihitung},
			{ID: 3, Status: domain.PeriodeDitutup},
		},
	}
	svc := newTestService(repo)

	hasil, err := svc.DaftarPeriodeTersedia(context.Background())
	if err != nil {
		t.Fatalf("DaftarPeriodeTersedia: %v", err)
	}
	if len(hasil) != 2 {
		t.Fatalf("DaftarPeriodeTersedia: jumlah = %d, ingin 2", len(hasil))
	}
}

func TestAmbilRingkasanDashboard_DataLengkap(t *testing.T) {
	repo := &mockRepo{
		periodes: []domain.PeriodeGaji{
			{ID: 1, Status: domain.PeriodeTerbuka},
			{ID: 2, Status: domain.PeriodeDitutup},
		},
		absensiBlmLengkap: 3,
		totalGajiBersih:   2_500_000,
		totalSisaKasbon:   500_000,
	}
	svc := newTestService(repo)

	r := svc.AmbilRingkasanDashboard(context.Background())
	if r.PeriodeBerjalan == nil || r.PeriodeBerjalan.ID != 1 {
		t.Fatal("Dashboard: periode berjalan harus ID=1 (TERBUKA)")
	}
	if r.PeriodeTerakhirSelesai == nil || r.PeriodeTerakhirSelesai.ID != 2 {
		t.Fatal("Dashboard: periode terakhir selesai harus ID=2 (DITUTUP)")
	}
	if r.AbsensiBelumLengkap != 3 || r.GajiBersihTerakhir != 2_500_000 || r.TotalSisaKasbon != 500_000 {
		t.Fatalf("Dashboard ringkasan salah: %+v", r)
	}
}

func TestAmbilRingkasanDashboard_RepoErrorGraceful(t *testing.T) {
	repo := &mockRepo{
		errPeriode:     errors.New("db error"),
		errAbsensi:     errors.New("db error"),
		errTotalGaji:   errors.New("db error"),
		errTotalKasbon: errors.New("db error"),
	}
	svc := newTestService(repo)

	r := svc.AmbilRingkasanDashboard(context.Background())
	if r.PeriodeBerjalan != nil || len(r.DaftarPeriode) != 0 {
		t.Fatal("Dashboard graceful: harus kosong tanpa panic saat error")
	}
}

// --- Pengujian Render Template & Handler (Regresi slip.html 500 & MiddlewarePeran) ---

func middlewarePeranUji(peran ...domain.Peran) func(http.Handler) http.Handler {
	boleh := make(map[domain.Peran]bool, len(peran))
	for _, p := range peran {
		boleh[p] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := web.PenggunaDariContext(r.Context())
			if u == nil || !boleh[u.Role] {
				http.Error(w, "Anda tidak memiliki akses ke halaman ini.", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func TestHandlerRenderSemuaHalamanModul6(t *testing.T) {
	mulai := time.Date(2026, 11, 2, 0, 0, 0, 0, time.Local)
	selesai := time.Date(2026, 11, 7, 0, 0, 0, 0, time.Local)
	periode := domain.PeriodeGaji{ID: 1, Tipe: domain.PeriodeMingguan, TanggalMulai: mulai, TanggalSelesai: selesai, Status: domain.PeriodeDihitung}
	budi := domain.Pekerja{ID: 2, Kode: "PK-01", Nama: "Budi", Jenis: domain.PekerjaHarian, Aktif: true}
	slip := domain.SlipGaji{
		ID: 10, PeriodeID: 1, PekerjaID: 2,
		UpahPokok: 600_000, TotalLembur: 31_250, PotonganTelat: 0, TotalKoreksi: 0,
		GajiBruto: 631_250, PotonganKasbon: 189_375, GajiBersih: 441_875,
		HariKerjaDibayar: 6, Pekerja: &budi, Periode: &periode,
	}
	repo := &mockRepo{
		status:            domain.PeriodeDihitung,
		slip:              slip,
		slips:             []domain.SlipGaji{slip},
		rincian:           []domain.SlipGajiDetail{{Tanggal: mulai, UpahPokokHari: 100_000, MenitLembur: 90, UpahLembur: 31_250}},
		periodes:          []domain.PeriodeGaji{periode},
		totalSisaKasbon:   300_000,
		sisaKasbonPekerja: 110_625,
	}

	renderer := &web.Render{Dir: filepath.Join("..", "..", "web", "templates")}
	mux := http.NewServeMux()
	DaftarkanRute(mux, renderer, newTestService(repo), middlewarePeranUji)

	admin := &domain.Pengguna{ID: 1, Username: "admin", Role: domain.PeranAdmin}

	kasus := []struct {
		path   string
		memuat []string
	}{
		{"/dashboard", []string{"Periode berjalan", "Rp 300.000"}},
		{"/rekap", []string{"Pilih Periode"}},
		{"/rekap/1", []string{"Budi", "Rp 441.875", "/slip/1/2"}},
		{"/slip/1/2", []string{"SLIP GAJI", "02 Nov 2026", "07 Nov 2026", "Rp 441.875", "Rp 110.625"}},
	}

	for _, k := range kasus {
		req := httptest.NewRequest("GET", k.path, nil)
		req = req.WithContext(web.DenganPengguna(req.Context(), admin))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s -> status %d, ingin 200; body: %s", k.path, rec.Code, rec.Body.String())
		}
		for _, teks := range k.memuat {
			if !strings.Contains(rec.Body.String(), teks) {
				t.Fatalf("GET %s tidak memuat %q", k.path, teks)
			}
		}
	}

	// Pastikan MiddlewarePeran benar-benar terpasang pada rute laporan.
	reqAnon := httptest.NewRequest("GET", "/rekap/1", nil)
	recAnon := httptest.NewRecorder()
	mux.ServeHTTP(recAnon, reqAnon)
	if recAnon.Code != http.StatusForbidden {
		t.Fatalf("rute laporan tanpa pengguna seharusnya 403, dapat %d", recAnon.Code)
	}
}
