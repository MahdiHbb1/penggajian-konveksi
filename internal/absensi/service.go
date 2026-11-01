// Pemilik: Zahra | Modul 3 | Aturan: 2, 3, 4, 5, 6, 14
package absensi

import (
	"context"
	"time"

	"penggajian-konveksi/internal/domain"
)

// Service mengatur absensi dan hitungan menit.
type Service struct {
	repo *Repository
}

// NewService membuat service absensi.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CatatAbsensi mencatat absensi untuk jadwal yang ada.
//
// Masukan: context dan data absensi.
// Hasil: ID absensi baru.
// Aturan: 2, 3, 14.
// Tabel: absensi, jadwal_shift, periode_gaji.
// Memanggil: Repository.SimpanAbsensi, PeriodeProvider.PeriksaPeriodeTerkunci.
//
// TODO(Zahra):
//  1. Pastikan jadwal ada pada tanggal itu.
//  2. Tolak absensi ganda untuk satu jadwal.
//  3. Validasi jam keluar setelah jam masuk, termasuk lewat tengah malam.
//  4. Tolak tanggal di periode tertutup.
func (s *Service) CatatAbsensi(ctx context.Context, absensi domain.Absensi) (int64, error) {
	return 0, domain.ErrBelumDiimplementasi
}

// UbahAbsensi memperbarui absensi.
//
// Masukan: context dan data absensi.
// Hasil: error.
// Aturan: 3, 14.
// Tabel: absensi, periode_gaji.
// Memanggil: Repository.UbahAbsensi, PeriodeProvider.PeriksaPeriodeTerkunci.
//
// TODO(Zahra):
//  1. Tolak perubahan bila periode tertutup.
//  2. Validasi jam keluar setelah jam masuk.
func (s *Service) UbahAbsensi(ctx context.Context, absensi domain.Absensi) error {
	return domain.ErrBelumDiimplementasi
}

// DaftarAbsensi mengambil absensi pada tanggal.
//
// Masukan: context dan tanggal.
// Hasil: daftar absensi lengkap.
// Aturan: 2, 3.
// Tabel: absensi, jadwal_shift.
// Memanggil: Repository.DaftarAbsensi.
func (s *Service) DaftarAbsensi(ctx context.Context, tanggal time.Time) ([]domain.AbsensiLengkap, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// DaftarAbsensiBelumLengkap mengambil absensi tanpa jam keluar.
//
// Masukan: context dan ID periode.
// Hasil: daftar absensi belum lengkap.
// Aturan: 2.
// Tabel: absensi, jadwal_shift.
// Memanggil: Repository.DaftarBelumLengkap.
func (s *Service) DaftarAbsensiBelumLengkap(ctx context.Context, periodeID int64) ([]domain.AbsensiLengkap, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// AbsensiDalamPeriode mengambil absensi lengkap pada periode.
//
// Masukan: context dan ID periode.
// Hasil: daftar absensi lengkap.
// Aturan: 2, 3.
// Tabel: absensi, jadwal_shift.
// Memanggil: Repository.AbsensiDalamPeriode.
func (s *Service) AbsensiDalamPeriode(ctx context.Context, periodeID int64) ([]domain.AbsensiLengkap, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// AdaAbsensiBelumLengkap mengecek absensi tanpa jam keluar.
//
// Masukan: context dan ID periode.
// Hasil: ada atau tidak, jumlah, error.
// Aturan: 2.
// Tabel: absensi, jadwal_shift.
// Memanggil: Repository.AdaBelumLengkap.
func (s *Service) AdaAbsensiBelumLengkap(ctx context.Context, periodeID int64) (bool, int, error) {
	return false, 0, domain.ErrBelumDiimplementasi
}

// HitungMenitTerlambat menghitung selisih jam masuk terhadap jam mulai shift.
//
// Masukan: jam masuk dan jam mulai shift.
// Hasil: menit terlambat, minimal 0.
// Aturan: 4, 5.
// Tabel: tidak ada.
// Memanggil: tidak ada.
//
// TODO(Zahra):
//  1. Hitung selisih jam masuk dikurangi jam mulai shift.
//  2. Kembalikan 0 bila negatif, karena datang lebih awal tidak dihitung.
func (s *Service) HitungMenitTerlambat(jamMasuk time.Time, jamMulaiShift time.Time) int {
	return 0
}

// HitungMenitLembur menghitung selisih jam keluar terhadap jam selesai shift.
//
// Masukan: jam keluar dan jam selesai shift.
// Hasil: menit lembur, minimal 0.
// Aturan: 4, 6.
// Tabel: tidak ada.
// Memanggil: tidak ada.
//
// TODO(Zahra):
//  1. Hitung selisih jam keluar dikurangi jam selesai shift.
//  2. Kembalikan 0 bila negatif, karena pulang lebih awal tidak dipotong.
func (s *Service) HitungMenitLembur(jamKeluar time.Time, jamSelesaiShift time.Time) int {
	return 0
}

// Pastikan Service memenuhi kontrak lintas modul.
var _ interface {
	AbsensiDalamPeriode(context.Context, int64) ([]domain.AbsensiLengkap, error)
	AdaAbsensiBelumLengkap(context.Context, int64) (bool, int, error)
	HitungMenitTerlambat(time.Time, time.Time) int
	HitungMenitLembur(time.Time, time.Time) int
} = (*Service)(nil)
