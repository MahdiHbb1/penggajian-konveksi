// Pemilik: Edgar | Modul 2 | Aturan: 3, 14, 19
package jadwal

import (
	"context"
	"time"

	"penggajian-konveksi/internal/domain"
)

// Service mengatur jadwal shift mingguan.
type Service struct {
	repo *Repository
}

// NewService membuat service jadwal.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// DaftarShift mengambil daftar shift.
//
// Masukan: context.
// Hasil: daftar shift.
// Aturan: 3.
// Tabel: shift.
// Memanggil: Repository.DaftarShift.
func (s *Service) DaftarShift(ctx context.Context) ([]domain.Shift, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// BuatJadwal menambahkan jadwal pekerja.
//
// Masukan: context dan data jadwal.
// Hasil: ID jadwal baru.
// Aturan: 3, 14, 19.
// Tabel: jadwal_shift, pekerja, periode_gaji.
// Memanggil: Repository.SimpanJadwal, PeriodeProvider.PeriksaPeriodeTerkunci.
//
// TODO(Edgar):
//  1. Tolak hari Minggu.
//  2. Pastikan pekerja aktif.
//  3. Pastikan tanggal tidak berada di periode tertutup.
//  4. Tolak jadwal ganda pada tanggal yang sama.
func (s *Service) BuatJadwal(ctx context.Context, jadwal domain.JadwalShift) (int64, error) {
	return 0, domain.ErrBelumDiimplementasi
}

// UbahJadwal mengganti shift pada jadwal.
//
// Masukan: context dan data jadwal.
// Hasil: error.
// Aturan: 14, 19.
// Tabel: jadwal_shift, absensi.
// Memanggil: Repository.UbahJadwal.
//
// TODO(Edgar):
//  1. Tolak perubahan bila jadwal sudah punya absensi.
//  2. Tolak perubahan bila tanggal di periode tertutup.
func (s *Service) UbahJadwal(ctx context.Context, jadwal domain.JadwalShift) error {
	return domain.ErrBelumDiimplementasi
}

// BatalkanJadwal membatalkan jadwal yang belum punya absensi.
//
// Masukan: context dan ID jadwal.
// Hasil: error.
// Aturan: 14, 19.
// Tabel: jadwal_shift, absensi.
// Memanggil: Repository.BatalkanJadwal.
//
// TODO(Edgar):
//  1. Tolak pembatalan bila jadwal sudah punya absensi.
//  2. Tolak pembatalan bila tanggal di periode tertutup.
func (s *Service) BatalkanJadwal(ctx context.Context, id int64) error {
	return domain.ErrBelumDiimplementasi
}

// SalinJadwalMingguLalu menyalin jadwal pekan sebelumnya.
//
// Masukan: context dan tanggal Senin pekan tujuan.
// Hasil: jumlah disalin, jumlah dilewati, error.
// Aturan: 3, 19.
// Tabel: jadwal_shift, pekerja, periode_gaji.
// Memanggil: Repository.SalinJadwalMingguLalu.
//
// TODO(Edgar):
//  1. Pastikan pekan tujuan belum terkunci.
//  2. Jangan menimpa sel yang sudah terisi.
//  3. Kembalikan jumlah sel disalin dan dilewati.
func (s *Service) SalinJadwalMingguLalu(ctx context.Context, mingguMulai time.Time) (int, int, error) {
	return 0, 0, domain.ErrBelumDiimplementasi
}

// AmbilJadwal mengambil jadwal pekerja pada tanggal.
//
// Masukan: context, ID pekerja, tanggal.
// Hasil: JadwalShift.
// Aturan: 3.
// Tabel: jadwal_shift, shift.
// Memanggil: Repository.AmbilJadwal.
func (s *Service) AmbilJadwal(ctx context.Context, pekerjaID int64, tanggal time.Time) (domain.JadwalShift, error) {
	return domain.JadwalShift{}, domain.ErrBelumDiimplementasi
}

// Pastikan Service memenuhi kontrak lintas modul.
var _ interface {
	AmbilJadwal(context.Context, int64, time.Time) (domain.JadwalShift, error)
} = (*Service)(nil)
