// Pemilik: Edgar | Modul 2 | Aturan: 3, 14, 19
package jadwal

import (
	"context"
	"database/sql"
	"time"

	"penggajian-konveksi/internal/domain"
)

// Repository mengakses tabel shift dan jadwal_shift.
type Repository struct {
	db *sql.DB
}

// NewRepository membuat repository jadwal.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// DaftarShift mengambil definisi shift.
//
// Masukan: context.
// Hasil: daftar shift.
// Aturan: 3.
// Tabel: shift.
// Memanggil: database/sql QueryContext.
func (r *Repository) DaftarShift(ctx context.Context) ([]domain.Shift, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// SimpanJadwal menyimpan satu jadwal pekerja.
//
// Masukan: context dan pointer JadwalShift.
// Hasil: error; ID hasil insert diisi ke pointer.
// Aturan: 3, 19.
// Tabel: jadwal_shift.
// Memanggil: database/sql ExecContext.
func (r *Repository) SimpanJadwal(ctx context.Context, jadwal *domain.JadwalShift) error {
	return domain.ErrBelumDiimplementasi
}

// UbahJadwal mengganti shift pada jadwal yang sudah ada.
//
// Masukan: context dan JadwalShift.
// Hasil: error.
// Aturan: 14, 19.
// Tabel: jadwal_shift, absensi.
// Memanggil: database/sql ExecContext.
func (r *Repository) UbahJadwal(ctx context.Context, jadwal domain.JadwalShift) error {
	return domain.ErrBelumDiimplementasi
}

// BatalkanJadwal menghapus jadwal yang belum punya absensi.
//
// Masukan: context dan ID jadwal.
// Hasil: error.
// Aturan: 14, 19.
// Tabel: jadwal_shift, absensi.
// Memanggil: database/sql ExecContext.
func (r *Repository) BatalkanJadwal(ctx context.Context, id int64) error {
	return domain.ErrBelumDiimplementasi
}

// SalinJadwalMingguLalu menyalin jadwal pekan sebelumnya.
//
// Masukan: context dan tanggal Senin pekan tujuan.
// Hasil: jumlah disalin, jumlah dilewati, error.
// Aturan: 3, 19.
// Tabel: jadwal_shift.
// Memanggil: database/sql QueryContext, database/sql ExecContext.
func (r *Repository) SalinJadwalMingguLalu(ctx context.Context, mingguMulai time.Time) (int, int, error) {
	return 0, 0, domain.ErrBelumDiimplementasi
}

// AmbilJadwal mengambil jadwal satu pekerja pada satu tanggal.
//
// Masukan: context, ID pekerja, tanggal.
// Hasil: JadwalShift beserta shift.
// Aturan: 3.
// Tabel: jadwal_shift, shift, pekerja.
// Memanggil: database/sql QueryRowContext.
func (r *Repository) AmbilJadwal(ctx context.Context, pekerjaID int64, tanggal time.Time) (domain.JadwalShift, error) {
	return domain.JadwalShift{}, domain.ErrBelumDiimplementasi
}
