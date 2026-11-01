// Pemilik: Zahra | Modul 3 | Aturan: 2, 3, 5, 6, 14
package absensi

import (
	"context"
	"database/sql"
	"time"

	"penggajian-konveksi/internal/domain"
)

// Repository mengakses tabel absensi.
type Repository struct {
	db *sql.DB
}

// NewRepository membuat repository absensi.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// SimpanAbsensi menyimpan absensi baru.
//
// Masukan: context dan pointer Absensi.
// Hasil: error; ID hasil insert diisi ke pointer.
// Aturan: 2, 3.
// Tabel: absensi, jadwal_shift.
// Memanggil: database/sql ExecContext.
func (r *Repository) SimpanAbsensi(ctx context.Context, absensi *domain.Absensi) error {
	return domain.ErrBelumDiimplementasi
}

// UbahAbsensi memperbarui jam masuk dan jam keluar.
//
// Masukan: context dan Absensi.
// Hasil: error.
// Aturan: 3, 14.
// Tabel: absensi.
// Memanggil: database/sql ExecContext.
func (r *Repository) UbahAbsensi(ctx context.Context, absensi domain.Absensi) error {
	return domain.ErrBelumDiimplementasi
}

// DaftarAbsensi mengambil absensi pada satu tanggal.
//
// Masukan: context dan tanggal.
// Hasil: daftar absensi lengkap.
// Aturan: 2, 3.
// Tabel: absensi, jadwal_shift, pekerja, shift.
// Memanggil: database/sql QueryContext.
func (r *Repository) DaftarAbsensi(ctx context.Context, tanggal time.Time) ([]domain.AbsensiLengkap, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// DaftarBelumLengkap mengambil absensi tanpa jam keluar pada periode.
//
// Masukan: context dan ID periode.
// Hasil: daftar absensi belum lengkap.
// Aturan: 2.
// Tabel: absensi, jadwal_shift, pekerja, shift, periode_gaji.
// Memanggil: database/sql QueryContext.
func (r *Repository) DaftarBelumLengkap(ctx context.Context, periodeID int64) ([]domain.AbsensiLengkap, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// AbsensiDalamPeriode mengambil seluruh absensi lengkap pada periode.
//
// Masukan: context dan ID periode.
// Hasil: daftar absensi lengkap.
// Aturan: 2, 3.
// Tabel: absensi, jadwal_shift, pekerja, shift, periode_gaji.
// Memanggil: database/sql QueryContext.
func (r *Repository) AbsensiDalamPeriode(ctx context.Context, periodeID int64) ([]domain.AbsensiLengkap, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// AdaBelumLengkap menghitung absensi tanpa jam keluar.
//
// Masukan: context dan ID periode.
// Hasil: ada atau tidak, jumlah, error.
// Aturan: 2.
// Tabel: absensi, jadwal_shift, periode_gaji.
// Memanggil: database/sql QueryRowContext.
func (r *Repository) AdaBelumLengkap(ctx context.Context, periodeID int64) (bool, int, error) {
	return false, 0, domain.ErrBelumDiimplementasi
}
