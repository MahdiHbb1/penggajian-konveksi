// Pemilik: Fadli | Modul 6 | Aturan: 13, 17
package laporan

import (
	"context"
	"fmt"

	"penggajian-konveksi/internal/domain"
)

// ErrPeriodeBelumDihitung dikembalikan bila periode masih berstatus Terbuka.
var ErrPeriodeBelumDihitung = fmt.Errorf("laporan hanya tersedia untuk periode Dihitung atau Ditutup")

// repoAkses mendefinisikan operasi repository yang dibutuhkan Service.
// Interface ini memungkinkan pengujian unit tanpa database.
type repoAkses interface {
	AmbilSlip(ctx context.Context, periodeID, pekerjaID int64) (domain.SlipGaji, error)
	DaftarSlip(ctx context.Context, periodeID int64) ([]domain.SlipGaji, error)
	Rekap(ctx context.Context, periodeID int64) (domain.RekapGaji, error)
	AmbilStatusPeriode(ctx context.Context, periodeID int64) (domain.StatusPeriode, error)
	DaftarRincianSlip(ctx context.Context, slipID int64) ([]domain.SlipGajiDetail, error)
	DaftarPeriodeTerbaru(ctx context.Context, limit int) ([]domain.PeriodeGaji, error)
	HitungAbsensiBelumLengkap(ctx context.Context, periodeID int64) (int, error)
	TotalGajiBersihPeriode(ctx context.Context, periodeID int64) (domain.Rupiah, error)
	TotalSisaKasbon(ctx context.Context) (domain.Rupiah, error)
}

// Pastikan Repository memenuhi repoAkses.
var _ repoAkses = (*Repository)(nil)

// Service menyediakan slip dan rekap penggajian.
type Service struct {
	repo repoAkses
}

// NewService membuat service laporan.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// periksaStatusPeriode memastikan periode berstatus Dihitung atau Ditutup.
// Mengembalikan ErrPeriodeBelumDihitung bila status Terbuka.
func (s *Service) periksaStatusPeriode(ctx context.Context, periodeID int64) error {
	status, err := s.repo.AmbilStatusPeriode(ctx, periodeID)
	if err != nil {
		return err
	}
	if status != domain.PeriodeDihitung && status != domain.PeriodeDitutup {
		return ErrPeriodeBelumDihitung
	}
	return nil
}

// AmbilSlipGaji mengambil satu slip dari data tersimpan.
// Hanya periode Dihitung atau Ditutup yang boleh ditampilkan.
func (s *Service) AmbilSlipGaji(ctx context.Context, periodeID, pekerjaID int64) (domain.SlipGaji, error) {
	slip, err := s.repo.AmbilSlip(ctx, periodeID, pekerjaID)
	if err != nil {
		return domain.SlipGaji{}, err
	}
	// Gunakan field Periode yang sudah di-JOIN oleh repository.
	if slip.Periode != nil {
		status := slip.Periode.Status
		if status != domain.PeriodeDihitung && status != domain.PeriodeDitutup {
			return domain.SlipGaji{}, ErrPeriodeBelumDihitung
		}
	}
	return slip, nil
}

// DaftarSlipPeriode mengambil semua slip dalam satu periode.
// Periode Terbuka ditolak; hanya Dihitung atau Ditutup yang diizinkan.
func (s *Service) DaftarSlipPeriode(ctx context.Context, periodeID int64) ([]domain.SlipGaji, error) {
	if err := s.periksaStatusPeriode(ctx, periodeID); err != nil {
		return nil, err
	}
	return s.repo.DaftarSlip(ctx, periodeID)
}

// DaftarRincianSlip mengambil rincian upah harian dari repository.
func (s *Service) DaftarRincianSlip(ctx context.Context, slipID int64) ([]domain.SlipGajiDetail, error) {
	return s.repo.DaftarRincianSlip(ctx, slipID)
}

// RekapGajiPerPeriode mengambil rekap penggajian satu periode.
// Periode Terbuka ditolak; hanya Dihitung atau Ditutup yang diizinkan.
// Potongan keterlambatan tidak dipotong ulang di rekap karena sudah terkandung dalam gaji_bruto.
func (s *Service) RekapGajiPerPeriode(ctx context.Context, periodeID int64) (domain.RekapGaji, error) {
	if err := s.periksaStatusPeriode(ctx, periodeID); err != nil {
		return domain.RekapGaji{}, err
	}
	return s.repo.Rekap(ctx, periodeID)
}

// RingkasanDashboard berisi data ringkasan untuk halaman dashboard.
type RingkasanDashboard struct {
	PeriodeBerjalan        *domain.PeriodeGaji
	AbsensiBelumLengkap    int
	GajiBersihTerakhir     domain.Rupiah
	TotalSisaKasbon        domain.Rupiah
	DaftarPeriode          []domain.PeriodeGaji
	PeriodeTerakhirSelesai *domain.PeriodeGaji
}

// AmbilRingkasanDashboard mengumpulkan data ringkasan dari database.
// Kesalahan pada setiap bagian ditangani secara graceful agar dashboard tetap tampil.
func (s *Service) AmbilRingkasanDashboard(ctx context.Context) RingkasanDashboard {
	var ringkasan RingkasanDashboard

	periodes, err := s.repo.DaftarPeriodeTerbaru(ctx, 10)
	if err == nil {
		ringkasan.DaftarPeriode = periodes
		// Cari periode berjalan (TERBUKA) dan terakhir selesai (DITUTUP).
		for i := range periodes {
			if periodes[i].Status == domain.PeriodeTerbuka && ringkasan.PeriodeBerjalan == nil {
				ringkasan.PeriodeBerjalan = &periodes[i]
			}
			if periodes[i].Status == domain.PeriodeDitutup && ringkasan.PeriodeTerakhirSelesai == nil {
				ringkasan.PeriodeTerakhirSelesai = &periodes[i]
			}
		}
	}

	// Hitung absensi belum lengkap pada periode berjalan.
	if ringkasan.PeriodeBerjalan != nil {
		n, err := s.repo.HitungAbsensiBelumLengkap(ctx, ringkasan.PeriodeBerjalan.ID)
		if err == nil {
			ringkasan.AbsensiBelumLengkap = n
		}
	}

	// Ambil total gaji bersih dari periode terakhir yang ditutup.
	if ringkasan.PeriodeTerakhirSelesai != nil {
		total, err := s.repo.TotalGajiBersihPeriode(ctx, ringkasan.PeriodeTerakhirSelesai.ID)
		if err == nil {
			ringkasan.GajiBersihTerakhir = total
		}
	}

	// Ambil total sisa kasbon.
	totalKasbon, err := s.repo.TotalSisaKasbon(ctx)
	if err == nil {
		ringkasan.TotalSisaKasbon = totalKasbon
	}

	return ringkasan
}

// DaftarPeriodeTersedia mengambil daftar periode dengan status Dihitung atau Ditutup
// untuk pemilih periode di halaman rekap.
func (s *Service) DaftarPeriodeTersedia(ctx context.Context) ([]domain.PeriodeGaji, error) {
	periodes, err := s.repo.DaftarPeriodeTerbaru(ctx, 50)
	if err != nil {
		return nil, err
	}
	var hasil []domain.PeriodeGaji
	for _, p := range periodes {
		if p.Status == domain.PeriodeDihitung || p.Status == domain.PeriodeDitutup {
			hasil = append(hasil, p)
		}
	}
	return hasil, nil
}

// Memastikan Service memenuhi kontrak layanan laporan.
var _ interface {
	AmbilSlipGaji(context.Context, int64, int64) (domain.SlipGaji, error)
	DaftarSlipPeriode(context.Context, int64) ([]domain.SlipGaji, error)
	DaftarRincianSlip(context.Context, int64) ([]domain.SlipGajiDetail, error)
	RekapGajiPerPeriode(context.Context, int64) (domain.RekapGaji, error)
} = (*Service)(nil)
