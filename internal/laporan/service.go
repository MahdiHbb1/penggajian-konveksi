// Pemilik: Fadli | Modul 6 | Aturan: 13, 17
package laporan

import (
	"context"

	"penggajian-konveksi/internal/domain"
)

// Service menyediakan slip dan rekap.
type Service struct {
	repo *Repository
}

// NewService membuat service laporan.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// AmbilSlipGaji mengambil satu slip.
//
// Masukan: context, ID periode, ID pekerja.
// Hasil: SlipGaji.
// Aturan: 13, 17.
// Tabel: slip_gaji, pekerja, periode_gaji.
// Memanggil: Repository.AmbilSlip.
//
// TODO(Fadli):
//  1. Hanya periode Dihitung atau Ditutup yang boleh menampilkan slip.
//  2. Slip dibaca dari data tersimpan, tidak dihitung ulang.
func (s *Service) AmbilSlipGaji(ctx context.Context, periodeID, pekerjaID int64) (domain.SlipGaji, error) {
	return domain.SlipGaji{}, domain.ErrBelumDiimplementasi
}

// DaftarSlipPeriode mengambil seluruh slip periode.
//
// Masukan: context dan ID periode.
// Hasil: daftar slip.
// Aturan: 13, 17.
// Tabel: slip_gaji.
// Memanggil: Repository.DaftarSlip.
func (s *Service) DaftarSlipPeriode(ctx context.Context, periodeID int64) ([]domain.SlipGaji, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// RekapGajiPerPeriode mengambil rekap satu periode.
//
// Masukan: context dan ID periode.
// Hasil: RekapGaji.
// Aturan: 12, 13.
// Tabel: slip_gaji, potongan_kasbon.
// Memanggil: Repository.Rekap.
//
// TODO(Fadli):
//  1. Total potongan pada rekap berarti potongan kasbon.
//  2. Potongan keterlambatan sudah mengurangi bruto, jangan dipotong lagi.
func (s *Service) RekapGajiPerPeriode(ctx context.Context, periodeID int64) (domain.RekapGaji, error) {
	return domain.RekapGaji{}, domain.ErrBelumDiimplementasi
}

// Pastikan Service memenuhi kontrak lintas modul.
var _ interface {
	AmbilSlipGaji(context.Context, int64, int64) (domain.SlipGaji, error)
	DaftarSlipPeriode(context.Context, int64) ([]domain.SlipGaji, error)
	RekapGajiPerPeriode(context.Context, int64) (domain.RekapGaji, error)
} = (*Service)(nil)
