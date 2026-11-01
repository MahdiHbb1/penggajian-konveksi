// Pemilik: Edgar | Modul 2 | Aturan: 12, 14, 16
package kasbon

import (
	"context"
	"time"

	"penggajian-konveksi/internal/domain"
)

// Service mengatur kasbon dan sisa potongannya.
type Service struct {
	repo *Repository
}

// NewService membuat service kasbon.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CatatKasbon mencatat kasbon baru.
//
// Masukan: context dan data kasbon.
// Hasil: ID kasbon baru.
// Aturan: 12, 14.
// Tabel: kasbon, pekerja, periode_gaji.
// Memanggil: Repository.SimpanKasbon, PeriodeProvider.PeriksaPeriodeTerkunci.
//
// TODO(Edgar):
//  1. Pastikan pekerja aktif.
//  2. Tolak tanggal di periode tertutup.
//  3. Validasi jumlah lebih dari 0 dan keterangan maksimal 150 karakter.
//  4. Set sisa awal sama dengan jumlah.
func (s *Service) CatatKasbon(ctx context.Context, kasbon domain.Kasbon) (int64, error) {
	return 0, domain.ErrBelumDiimplementasi
}

// DaftarKasbon mengambil daftar kasbon.
//
// Masukan: context dan Filter.
// Hasil: daftar kasbon.
// Aturan: 12, 16.
// Tabel: kasbon, pekerja.
// Memanggil: Repository.DaftarKasbon.
func (s *Service) DaftarKasbon(ctx context.Context, filter Filter) ([]domain.Kasbon, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// AmbilKasbon mengambil satu kasbon beserta riwayat potongannya.
//
// Masukan: context dan ID kasbon.
// Hasil: Kasbon.
// Aturan: 12.
// Tabel: kasbon, potongan_kasbon.
// Memanggil: Repository.AmbilKasbon.
func (s *Service) AmbilKasbon(ctx context.Context, id int64) (domain.Kasbon, error) {
	return domain.Kasbon{}, domain.ErrBelumDiimplementasi
}

// SisaKasbon menghasilkan total sisa kasbon pekerja.
//
// Masukan: context dan ID pekerja.
// Hasil: total sisa.
// Aturan: 12, 16.
// Tabel: kasbon.
// Memanggil: Repository.SisaKasbon.
func (s *Service) SisaKasbon(ctx context.Context, pekerjaID int64) (domain.Rupiah, error) {
	return 0, domain.ErrBelumDiimplementasi
}

// KasbonDiprosesSampai menghasilkan kasbon yang ikut dipotong pada periode.
//
// Masukan: context, ID pekerja, batas tanggal.
// Hasil: kasbon terurut dari yang tertua.
// Aturan: 12.
// Tabel: kasbon.
// Memanggil: Repository.KasbonDiprosesSampai.
func (s *Service) KasbonDiprosesSampai(ctx context.Context, pekerjaID int64, tanggal time.Time) ([]domain.Kasbon, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// RiwayatPotongan mengambil potongan kasbon per periode.
//
// Masukan: context dan ID kasbon.
// Hasil: daftar potongan.
// Aturan: 12, 13.
// Tabel: potongan_kasbon, slip_gaji, periode_gaji.
// Memanggil: Repository.RiwayatPotongan.
func (s *Service) RiwayatPotongan(ctx context.Context, kasbonID int64) ([]domain.PotonganKasbon, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// Pastikan Service memenuhi kontrak lintas modul.
var _ interface {
	KasbonDiprosesSampai(context.Context, int64, time.Time) ([]domain.Kasbon, error)
	SisaKasbon(context.Context, int64) (domain.Rupiah, error)
} = (*Service)(nil)
