// Pemilik: Anya | Modul 5 | Aturan: 1, 12, 13, 14, 15, 18
package periode

import (
	"context"
	"time"

	"penggajian-konveksi/internal/domain"
)

// Service mengatur periode, penutupan, dan koreksi.
type Service struct {
	repo *Repository
}

// NewService membuat service periode.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// BuatPeriode membuat periode mingguan atau bulanan.
//
// Masukan: context, tipe periode, tanggal mulai.
// Hasil: ID periode baru.
// Aturan: 1, 13.
// Tabel: periode_gaji.
// Memanggil: Repository.SimpanPeriode.
//
// TODO(Anya):
//  1. Mingguan wajib mulai Senin, selesai Sabtu.
//  2. Bulanan wajib mulai tanggal 1, selesai akhir bulan.
//  3. Tolak tumpang tindih tipe yang sama.
//  4. Set status Terbuka.
func (s *Service) BuatPeriode(ctx context.Context, tipe domain.TipePeriode, tanggalMulai time.Time) (int64, error) {
	return 0, domain.ErrBelumDiimplementasi
}

// DaftarPeriode mengambil daftar periode.
//
// Masukan: context.
// Hasil: daftar periode.
// Aturan: 13.
// Tabel: periode_gaji.
// Memanggil: Repository.DaftarPeriode.
func (s *Service) DaftarPeriode(ctx context.Context) ([]domain.PeriodeGaji, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// AmbilPeriode mengambil satu periode.
//
// Masukan: context dan ID periode.
// Hasil: PeriodeGaji.
// Aturan: 13.
// Tabel: periode_gaji.
// Memanggil: Repository.AmbilPeriode.
func (s *Service) AmbilPeriode(ctx context.Context, id int64) (domain.PeriodeGaji, error) {
	return domain.PeriodeGaji{}, domain.ErrBelumDiimplementasi
}

// PeriksaPeriodeTerkunci mengecek apakah tanggal terkunci.
//
// Masukan: context dan tanggal.
// Hasil: true bila periode sudah ditutup.
// Aturan: 14.
// Tabel: periode_gaji.
// Memanggil: Repository.PeriodeTerkunci.
func (s *Service) PeriksaPeriodeTerkunci(ctx context.Context, tanggal time.Time) (bool, error) {
	return false, domain.ErrBelumDiimplementasi
}

// PeriodeSebelumnyaDitutup mengecek prasyarat periode sebelumnya.
//
// Masukan: context dan periode acuan.
// Hasil: true bila tidak ada periode sebelumnya atau sebelumnya sudah ditutup.
// Aturan: 13.
// Tabel: periode_gaji.
// Memanggil: Repository.PeriodeSebelumnya.
//
// TODO(Anya):
//  1. Cari periode bertipe sama yang lebih awal.
//  2. Bila tidak ada, anggap terpenuhi.
//  3. Bila ada, statusnya harus Ditutup.
func (s *Service) PeriodeSebelumnyaDitutup(ctx context.Context, periode domain.PeriodeGaji) (bool, error) {
	return false, domain.ErrBelumDiimplementasi
}

// UbahStatusPeriode memperbarui status periode.
//
// Masukan: context, ID periode, status baru.
// Hasil: error.
// Aturan: 13.
// Tabel: periode_gaji.
// Memanggil: Repository.UbahStatus.
func (s *Service) UbahStatusPeriode(ctx context.Context, periodeID int64, status domain.StatusPeriode) error {
	return domain.ErrBelumDiimplementasi
}

// TutupPeriode mengunci periode dan memperbarui sisa kasbon.
//
// Masukan: context dan ID periode.
// Hasil: error.
// Aturan: 12, 13, 14, 15.
// Tabel: periode_gaji, slip_gaji, potongan_kasbon, kasbon, koreksi_gaji.
// Memanggil: Repository.AmbilPeriode, Repository.Tutup.
//
// TODO(Anya):
//  1. Periode hanya boleh ditutup dari status Dihitung.
//  2. Dalam satu transaksi: ubah status, isi ditutup_pada.
//  3. Kurangi sisa kasbon sesuai potongan pada slip.
//  4. Tandai koreksi yang diterapkan pada periode ini sebagai selesai.
func (s *Service) TutupPeriode(ctx context.Context, periodeID int64) error {
	return domain.ErrBelumDiimplementasi
}

// RingkasanTutupPeriode mengambil angka konfirmasi penutupan.
//
// Masukan: context dan ID periode.
// Hasil: RingkasanTutup.
// Aturan: 13.
// Tabel: slip_gaji, potongan_kasbon, koreksi_gaji.
// Memanggil: Repository.Ringkasan.
func (s *Service) RingkasanTutupPeriode(ctx context.Context, periodeID int64) (RingkasanTutup, error) {
	return RingkasanTutup{}, domain.ErrBelumDiimplementasi
}

// CatatKoreksi mencatat koreksi untuk periode yang sudah ditutup.
//
// Masukan: context dan data koreksi.
// Hasil: ID koreksi baru.
// Aturan: 15.
// Tabel: koreksi_gaji, periode_gaji.
// Memanggil: Repository.AmbilPeriode, Repository.SimpanKoreksi.
//
// TODO(Anya):
//  1. Periode asal wajib dan harus Ditutup.
//  2. Nominal tidak boleh nol dan alasan wajib.
//  3. Status Menunggu diturunkan dari periode_terapan_id yang kosong.
func (s *Service) CatatKoreksi(ctx context.Context, koreksi domain.KoreksiGaji) (int64, error) {
	return 0, domain.ErrBelumDiimplementasi
}

// DaftarKoreksi mengambil daftar koreksi.
//
// Masukan: context dan ID pekerja (boleh nil).
// Hasil: daftar koreksi.
// Aturan: 15, 18.
// Tabel: koreksi_gaji, periode_gaji.
// Memanggil: Repository.DaftarKoreksi.
func (s *Service) DaftarKoreksi(ctx context.Context, pekerjaID *int64) ([]domain.KoreksiGaji, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// DaftarKoreksiMenunggu mengambil koreksi yang belum diterapkan.
//
// Masukan: context dan ID pekerja.
// Hasil: daftar koreksi.
// Aturan: 15, 18.
// Tabel: koreksi_gaji.
// Memanggil: Repository.DaftarKoreksiMenunggu.
func (s *Service) DaftarKoreksiMenunggu(ctx context.Context, pekerjaID int64) ([]domain.KoreksiGaji, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// TandaiKoreksiDiterapkan menandai koreksi masuk ke periode.
//
// Masukan: context dan ID periode penerapan.
// Hasil: error.
// Aturan: 15, 18.
// Tabel: koreksi_gaji.
// Memanggil: Repository.TandaiKoreksiDiterapkan.
func (s *Service) TandaiKoreksiDiterapkan(ctx context.Context, periodeID int64) error {
	return domain.ErrBelumDiimplementasi
}

// Pastikan Service memenuhi kontrak lintas modul.
var _ interface {
	AmbilPeriode(context.Context, int64) (domain.PeriodeGaji, error)
	PeriodeSebelumnyaDitutup(context.Context, domain.PeriodeGaji) (bool, error)
	UbahStatusPeriode(context.Context, int64, domain.StatusPeriode) error
	PeriksaPeriodeTerkunci(context.Context, time.Time) (bool, error)
	DaftarKoreksiMenunggu(context.Context, int64) ([]domain.KoreksiGaji, error)
	TandaiKoreksiDiterapkan(context.Context, int64) error
} = (*Service)(nil)
