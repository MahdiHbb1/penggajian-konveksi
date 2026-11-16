package laporan

import (
	"context"
	"fmt"

	"penggajian-konveksi/internal/domain"
)

// Service menyediakan slip dan rekap penggajian.
type Service struct {
	repo *Repository
}

// NewService membuat service laporan.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// AmbilSlipGaji mengambil satu slip dari data tersimpan.
func (s *Service) AmbilSlipGaji(ctx context.Context, periodeID, pekerjaID int64) (domain.SlipGaji, error) {
	slip, err := s.repo.AmbilSlip(ctx, periodeID, pekerjaID)
	if err != nil {
		return domain.SlipGaji{}, err
	}

	if slip.Periode != nil {
		status := slip.Periode.Status
		if status != domain.PeriodeDihitung && status != domain.PeriodeDitutup {
			return domain.SlipGaji{}, fmt.Errorf("slip hanya tersedia untuk periode Dihitung atau Ditutup")
		}
	}

	return slip, nil
}

// DaftarSlipPeriode mengambil semua slip dalam satu periode.
func (s *Service) DaftarSlipPeriode(ctx context.Context, periodeID int64) ([]domain.SlipGaji, error) {
	return s.repo.DaftarSlip(ctx, periodeID)
}

// DaftarRincianSlip mengambil rincian upah harian dari repository.
func (s *Service) DaftarRincianSlip(ctx context.Context, slipID int64) ([]domain.SlipGajiDetail, error) {
	return s.repo.DaftarRincianSlip(ctx, slipID)
}

// RekapGajiPerPeriode mengambil rekap penggajian satu periode.
func (s *Service) RekapGajiPerPeriode(ctx context.Context, periodeID int64) (domain.RekapGaji, error) {
	return s.repo.Rekap(ctx, periodeID)
}

// Memastikan Service memenuhi kontrak layanan laporan.
var _ interface {
	AmbilSlipGaji(context.Context, int64, int64) (domain.SlipGaji, error)
	DaftarSlipPeriode(context.Context, int64) ([]domain.SlipGaji, error)
	DaftarRincianSlip(context.Context, int64) ([]domain.SlipGajiDetail, error)
	RekapGajiPerPeriode(context.Context, int64) (domain.RekapGaji, error)
} = (*Service)(nil)
