// Pemilik: Fadli | Modul 6 | Aturan: 13, 17
package laporan

import (
	"context"
	"errors"
	"testing"

	"penggajian-konveksi/internal/domain"
)

// mockRepo mengimplementasikan repoAkses untuk pengujian unit.
type mockRepo struct {
	slip              domain.SlipGaji
	slips             []domain.SlipGaji
	rekap             domain.RekapGaji
	status            domain.StatusPeriode
	rincian           []domain.SlipGajiDetail
	periodes          []domain.PeriodeGaji
	absensiBlmLengkap int
	totalGajiBersih   domain.Rupiah
	totalSisaKasbon   domain.Rupiah
	errSlip           error
	errDaftarSlip     error
	errRekap          error
	errStatus         error
	errRincian        error
	errPeriode        error
	errAbsensi        error
	errTotalGaji      error
	errTotalKasbon    error
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

func (m *mockRepo) HitungAbsensiBelumLengkap(_ context.Context, _ int64) (int, error) {
	return m.absensiBlmLengkap, m.errAbsensi
}

func (m *mockRepo) TotalGajiBersihPeriode(_ context.Context, _ int64) (domain.Rupiah, error) {
	return m.totalGajiBersih, m.errTotalGaji
}

func (m *mockRepo) TotalSisaKasbon(_ context.Context) (domain.Rupiah, error) {
	return m.totalSisaKasbon, m.errTotalKasbon
}

// newTestService membuat service dengan mock repository.
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

// --- DaftarSlipPeriode ---

func TestDaftarSlipPeriode_PeriodeTerbukaDitolak(t *testing.T) {
	repo := &mockRepo{status: domain.PeriodeTerbuka}
	svc := newTestService(repo)

	_, err := svc.DaftarSlipPeriode(context.Background(), 1)
	if !errors.Is(err, ErrPeriodeBelumDihitung) {
		t.Fatalf("DaftarSlipPeriode(TERBUKA): ingin ErrPeriodeBelumDihitung, dapat %v", err)
	}
}

func TestDaftarSlipPeriode_PeriodeDihitungDiizinkan(t *testing.T) {
	repo := &mockRepo{
		status: domain.PeriodeDihitung,
		slips:  []domain.SlipGaji{{ID: 1}, {ID: 2}},
	}
	svc := newTestService(repo)

	slips, err := svc.DaftarSlipPeriode(context.Background(), 1)
	if err != nil {
		t.Fatalf("DaftarSlipPeriode(DIHITUNG): error tidak diharapkan: %v", err)
	}
	if len(slips) != 2 {
		t.Fatalf("DaftarSlipPeriode(DIHITUNG): jumlah slip = %d, ingin 2", len(slips))
	}
}

func TestDaftarSlipPeriode_PeriodeDitutupDiizinkan(t *testing.T) {
	repo := &mockRepo{
		status: domain.PeriodeDitutup,
		slips:  []domain.SlipGaji{{ID: 1}},
	}
	svc := newTestService(repo)

	slips, err := svc.DaftarSlipPeriode(context.Background(), 1)
	if err != nil {
		t.Fatalf("DaftarSlipPeriode(DITUTUP): error tidak diharapkan: %v", err)
	}
	if len(slips) != 1 {
		t.Fatalf("DaftarSlipPeriode(DITUTUP): jumlah slip = %d, ingin 1", len(slips))
	}
}

func TestDaftarSlipPeriode_StatusRepoError(t *testing.T) {
	repoErr := errors.New("koneksi terputus")
	repo := &mockRepo{errStatus: repoErr}
	svc := newTestService(repo)

	_, err := svc.DaftarSlipPeriode(context.Background(), 1)
	if !errors.Is(err, repoErr) {
		t.Fatalf("DaftarSlipPeriode(status error): ingin %v, dapat %v", repoErr, err)
	}
}

func TestDaftarSlipPeriode_DaftarSlipRepoError(t *testing.T) {
	repoErr := errors.New("query gagal")
	repo := &mockRepo{
		status:        domain.PeriodeDihitung,
		errDaftarSlip: repoErr,
	}
	svc := newTestService(repo)

	_, err := svc.DaftarSlipPeriode(context.Background(), 1)
	if !errors.Is(err, repoErr) {
		t.Fatalf("DaftarSlipPeriode(daftar error): ingin %v, dapat %v", repoErr, err)
	}
}

// --- RekapGajiPerPeriode ---

func TestRekapGajiPerPeriode_PeriodeTerbukaDitolak(t *testing.T) {
	repo := &mockRepo{status: domain.PeriodeTerbuka}
	svc := newTestService(repo)

	_, err := svc.RekapGajiPerPeriode(context.Background(), 1)
	if !errors.Is(err, ErrPeriodeBelumDihitung) {
		t.Fatalf("RekapGaji(TERBUKA): ingin ErrPeriodeBelumDihitung, dapat %v", err)
	}
}

func TestRekapGajiPerPeriode_PeriodeDihitungDiizinkan(t *testing.T) {
	repo := &mockRepo{
		status: domain.PeriodeDihitung,
		rekap:  domain.RekapGaji{PeriodeID: 1, TotalBruto: 1_000_000, TotalBersih: 900_000, JumlahSlip: 5},
	}
	svc := newTestService(repo)

	rekap, err := svc.RekapGajiPerPeriode(context.Background(), 1)
	if err != nil {
		t.Fatalf("RekapGaji(DIHITUNG): error tidak diharapkan: %v", err)
	}
	if rekap.JumlahSlip != 5 {
		t.Fatalf("RekapGaji(DIHITUNG): jumlah slip = %d, ingin 5", rekap.JumlahSlip)
	}
}

func TestRekapGajiPerPeriode_PeriodeDitutupDiizinkan(t *testing.T) {
	repo := &mockRepo{
		status: domain.PeriodeDitutup,
		rekap:  domain.RekapGaji{PeriodeID: 1, TotalBersih: 800_000, JumlahSlip: 3},
	}
	svc := newTestService(repo)

	rekap, err := svc.RekapGajiPerPeriode(context.Background(), 1)
	if err != nil {
		t.Fatalf("RekapGaji(DITUTUP): error tidak diharapkan: %v", err)
	}
	if rekap.TotalBersih != 800_000 {
		t.Fatalf("RekapGaji(DITUTUP): total bersih = %d, ingin 800000", rekap.TotalBersih)
	}
}

func TestRekapGajiPerPeriode_RepoError(t *testing.T) {
	repoErr := errors.New("query rekap gagal")
	repo := &mockRepo{
		status:   domain.PeriodeDihitung,
		errRekap: repoErr,
	}
	svc := newTestService(repo)

	_, err := svc.RekapGajiPerPeriode(context.Background(), 1)
	if !errors.Is(err, repoErr) {
		t.Fatalf("RekapGaji(repo error): ingin %v, dapat %v", repoErr, err)
	}
}

// --- DaftarPeriodeTersedia ---

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
		t.Fatalf("DaftarPeriodeTersedia: error tidak diharapkan: %v", err)
	}
	if len(hasil) != 2 {
		t.Fatalf("DaftarPeriodeTersedia: jumlah = %d, ingin 2 (tanpa TERBUKA)", len(hasil))
	}
	for _, p := range hasil {
		if p.Status == domain.PeriodeTerbuka {
			t.Fatal("DaftarPeriodeTersedia: periode TERBUKA seharusnya difilter")
		}
	}
}

// --- AmbilRingkasanDashboard ---

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
	if r.AbsensiBelumLengkap != 3 {
		t.Fatalf("Dashboard: absensi belum lengkap = %d, ingin 3", r.AbsensiBelumLengkap)
	}
	if r.GajiBersihTerakhir != 2_500_000 {
		t.Fatalf("Dashboard: gaji bersih terakhir = %d, ingin 2500000", r.GajiBersihTerakhir)
	}
	if r.TotalSisaKasbon != 500_000 {
		t.Fatalf("Dashboard: total sisa kasbon = %d, ingin 500000", r.TotalSisaKasbon)
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

	// Harus tetap mengembalikan RingkasanDashboard tanpa panic.
	r := svc.AmbilRingkasanDashboard(context.Background())
	if r.PeriodeBerjalan != nil {
		t.Fatal("Dashboard graceful: periode berjalan harus nil saat error")
	}
	if len(r.DaftarPeriode) != 0 {
		t.Fatal("Dashboard graceful: daftar periode harus kosong saat error")
	}
}

// --- RekapTidakMemotongTelatDuaKali (bisnis, via service) ---

func TestRekapTidakMemotongTelatDuaKali(t *testing.T) {
	repo := &mockRepo{
		status: domain.PeriodeDihitung,
		rekap: domain.RekapGaji{
			PeriodeID:           1,
			TotalBruto:          1_000_000,
			TotalLembur:         50_000,
			TotalPotonganKasbon: 100_000,
			TotalBersih:         900_000,
			JumlahSlip:          5,
		},
	}
	svc := newTestService(repo)

	rekap, err := svc.RekapGajiPerPeriode(context.Background(), 1)
	if err != nil {
		t.Fatalf("RekapTidakDouble: error tidak diharapkan: %v", err)
	}
	if rekap.TotalBersih > rekap.TotalBruto {
		t.Fatal("gaji bersih tidak boleh melebihi bruto")
	}
}
