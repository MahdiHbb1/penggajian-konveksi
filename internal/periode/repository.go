// Pemilik: Anya | Modul 5 | Aturan: 1, 12, 13, 14, 15, 18
package periode

import (
	"context"
	"database/sql"
	"time"

	"penggajian-konveksi/internal/domain"
)

// RingkasanTutup adalah data konfirmasi sebelum periode ditutup.
type RingkasanTutup struct {
	JumlahSlip          int
	TotalBersih         domain.Rupiah
	TotalPotonganKasbon domain.Rupiah
	JumlahKoreksi       int
}

// Repository mengakses tabel periode_gaji, koreksi_gaji, dan kasbon.
type Repository struct {
	db *sql.DB
}

// NewRepository membuat repository periode.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// SimpanPeriode menyimpan periode baru.
//
// Masukan: context dan pointer PeriodeGaji.
// Hasil: error; ID hasil insert diisi ke pointer.
// Aturan: 1, 13.
// Tabel: periode_gaji.
// Memanggil: database/sql ExecContext.
func (r *Repository) SimpanPeriode(ctx context.Context, periode *domain.PeriodeGaji) error {
	return domain.ErrBelumDiimplementasi
}

// DaftarPeriode mengambil seluruh periode terbaru.
//
// Masukan: context.
// Hasil: daftar periode.
// Aturan: 13.
// Tabel: periode_gaji.
// Memanggil: database/sql QueryContext.
func (r *Repository) DaftarPeriode(ctx context.Context) ([]domain.PeriodeGaji, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// AmbilPeriode mengambil satu periode.
//
// Masukan: context dan ID periode.
// Hasil: PeriodeGaji.
// Aturan: 13.
// Tabel: periode_gaji.
// Memanggil: database/sql QueryRowContext.
func (r *Repository) AmbilPeriode(ctx context.Context, id int64) (domain.PeriodeGaji, error) {
	return domain.PeriodeGaji{}, domain.ErrBelumDiimplementasi
}

// PeriodeTerkunci mengecek apakah tanggal berada di periode Ditutup.
//
// Masukan: context dan tanggal.
// Hasil: true bila terkunci.
// Aturan: 14.
// Tabel: periode_gaji.
// Memanggil: database/sql QueryRowContext.
func (r *Repository) PeriodeTerkunci(ctx context.Context, tanggal time.Time) (bool, error) {
	return false, domain.ErrBelumDiimplementasi
}

// PeriodeSebelumnya mengambil periode bertipe sama sebelum periode ini.
//
// Masukan: context dan periode acuan.
// Hasil: periode sebelumnya dan status ditemukan.
// Aturan: 13.
// Tabel: periode_gaji.
// Memanggil: database/sql QueryRowContext.
func (r *Repository) PeriodeSebelumnya(ctx context.Context, periode domain.PeriodeGaji) (domain.PeriodeGaji, bool, error) {
	return domain.PeriodeGaji{}, false, domain.ErrBelumDiimplementasi
}

// UbahStatus memperbarui status periode.
//
// Masukan: context, ID periode, status baru.
// Hasil: error.
// Aturan: 13.
// Tabel: periode_gaji.
// Memanggil: database/sql ExecContext.
func (r *Repository) UbahStatus(ctx context.Context, id int64, status domain.StatusPeriode) error {
	return domain.ErrBelumDiimplementasi
}

// Tutup menyimpan penutupan periode dalam satu transaksi.
//
// Masukan: context dan ID periode.
// Hasil: error.
// Aturan: 12, 13, 14, 15.
// Tabel: periode_gaji, slip_gaji, potongan_kasbon, kasbon, koreksi_gaji.
// Memanggil: database/sql BeginTx, ExecContext.
func (r *Repository) Tutup(ctx context.Context, id int64) error {
	return domain.ErrBelumDiimplementasi
}

// Ringkasan mengambil angka konfirmasi tutup periode.
//
// Masukan: context dan ID periode.
// Hasil: RingkasanTutup.
// Aturan: 13.
// Tabel: slip_gaji, potongan_kasbon, koreksi_gaji.
// Memanggil: database/sql QueryRowContext.
func (r *Repository) Ringkasan(ctx context.Context, id int64) (RingkasanTutup, error) {
	return RingkasanTutup{}, domain.ErrBelumDiimplementasi
}

// SimpanKoreksi menyimpan koreksi gaji baru.
//
// Masukan: context dan pointer KoreksiGaji.
// Hasil: error; ID hasil insert diisi ke pointer.
// Aturan: 15.
// Tabel: koreksi_gaji.
// Memanggil: database/sql ExecContext.
func (r *Repository) SimpanKoreksi(ctx context.Context, koreksi *domain.KoreksiGaji) error {
	return domain.ErrBelumDiimplementasi
}

// DaftarKoreksi mengambil koreksi, opsional per pekerja.
//
// Masukan: context dan ID pekerja (boleh nil).
// Hasil: daftar koreksi.
// Aturan: 15, 18.
// Tabel: koreksi_gaji, pekerja, periode_gaji.
// Memanggil: database/sql QueryContext.
func (r *Repository) DaftarKoreksi(ctx context.Context, pekerjaID *int64) ([]domain.KoreksiGaji, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// DaftarKoreksiMenunggu mengambil koreksi yang belum diterapkan.
//
// Masukan: context dan ID pekerja.
// Hasil: daftar koreksi Menunggu.
// Aturan: 15, 18.
// Tabel: koreksi_gaji.
// Memanggil: database/sql QueryContext.
func (r *Repository) DaftarKoreksiMenunggu(ctx context.Context, pekerjaID int64) ([]domain.KoreksiGaji, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// TandaiKoreksiDiterapkan mengisi periode_terapan_id.
//
// Masukan: context dan ID periode penerapan.
// Hasil: error.
// Aturan: 15, 18.
// Tabel: koreksi_gaji.
// Memanggil: database/sql ExecContext.
func (r *Repository) TandaiKoreksiDiterapkan(ctx context.Context, periodeID int64) error {
	return domain.ErrBelumDiimplementasi
}
