// Pemilik: Mahdi | Modul 1 | Aturan: 8, 9, 16
package pekerja

import (
	"context"
	"database/sql"
	"time"

	"penggajian-konveksi/internal/domain"
)

// Filter adalah parameter pencarian daftar pekerja.
type Filter struct {
	Cari  string
	Jenis domain.JenisPekerja
	Aktif *bool
}

// Repository mengakses tabel pekerja dan riwayat_upah.
type Repository struct {
	db *sql.DB
}

// NewRepository membuat repository pekerja.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// SimpanPekerja menyimpan satu baris pekerja baru.
//
// Masukan: context dan pointer Pekerja.
// Hasil: error; ID hasil insert diisi ke pointer.
// Aturan: 1, 11, 16.
// Tabel: pekerja.
// Memanggil: database/sql ExecContext.
func (r *Repository) SimpanPekerja(ctx context.Context, pekerja *domain.Pekerja) error {
	return domain.ErrBelumDiimplementasi
}

// UbahPekerja memperbarui data pekerja yang sudah ada.
//
// Masukan: context dan Pekerja.
// Hasil: error.
// Aturan: 11, 16.
// Tabel: pekerja.
// Memanggil: database/sql ExecContext.
func (r *Repository) UbahPekerja(ctx context.Context, pekerja domain.Pekerja) error {
	return domain.ErrBelumDiimplementasi
}

// NonaktifkanPekerja mengisi tanggal berhenti dan aktif=false.
//
// Masukan: context, ID pekerja, tanggal berhenti.
// Hasil: error.
// Aturan: 16.
// Tabel: pekerja.
// Memanggil: database/sql ExecContext.
func (r *Repository) NonaktifkanPekerja(ctx context.Context, id int64, tanggal time.Time) error {
	return domain.ErrBelumDiimplementasi
}

// DaftarPekerja mengambil pekerja sesuai filter.
//
// Masukan: context dan Filter.
// Hasil: daftar pekerja.
// Aturan: 11, 16.
// Tabel: pekerja.
// Memanggil: database/sql QueryContext.
func (r *Repository) DaftarPekerja(ctx context.Context, filter Filter) ([]domain.Pekerja, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// CariPekerja mengambil satu pekerja berdasarkan ID.
//
// Masukan: context dan ID pekerja.
// Hasil: Pekerja.
// Aturan: 16.
// Tabel: pekerja.
// Memanggil: database/sql QueryRowContext.
func (r *Repository) CariPekerja(ctx context.Context, id int64) (domain.Pekerja, error) {
	return domain.Pekerja{}, domain.ErrBelumDiimplementasi
}

// SimpanUpah menyimpan satu baris riwayat upah.
//
// Masukan: context dan pointer Upah.
// Hasil: error; ID hasil insert diisi ke pointer.
// Aturan: 7, 8.
// Tabel: riwayat_upah.
// Memanggil: database/sql ExecContext.
func (r *Repository) SimpanUpah(ctx context.Context, upah *domain.Upah) error {
	return domain.ErrBelumDiimplementasi
}

// DaftarUpah mengambil riwayat upah satu pekerja.
//
// Masukan: context dan ID pekerja.
// Hasil: daftar upah terurut tanggal berlaku.
// Aturan: 8.
// Tabel: riwayat_upah.
// Memanggil: database/sql QueryContext.
func (r *Repository) DaftarUpah(ctx context.Context, pekerjaID int64) ([]domain.Upah, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// UpahBerlaku mengambil upah terbaru yang tidak melewati tanggal.
//
// Masukan: context, ID pekerja, tanggal kerja.
// Hasil: Upah yang berlaku.
// Aturan: 8.
// Tabel: riwayat_upah.
// Memanggil: database/sql QueryRowContext.
func (r *Repository) UpahBerlaku(ctx context.Context, pekerjaID int64, tanggal time.Time) (domain.Upah, error) {
	return domain.Upah{}, domain.ErrBelumDiimplementasi
}

// PekerjaAktif mengambil pekerja yang aktif pada rentang periode.
//
// Masukan: context, jenis pekerja, awal dan akhir periode.
// Hasil: daftar pekerja.
// Aturan: 9, 16, 17.
// Tabel: pekerja, riwayat_upah.
// Memanggil: database/sql QueryContext.
func (r *Repository) PekerjaAktif(ctx context.Context, jenis domain.JenisPekerja, mulai, selesai time.Time) ([]domain.Pekerja, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// SimpanPekerjaDanUpah menyimpan pekerja dan upah pertamanya dalam satu transaksi.
//
// Masukan: context, pointer Pekerja, pointer Upah.
// Hasil: error; ID pekerja dan upah diisi ke pointer.
// Aturan: 8, 11, 16.
// Tabel: pekerja, riwayat_upah.
// Memanggil: database/sql BeginTx, SimpanPekerja, SimpanUpah, Commit.
func (r *Repository) SimpanPekerjaDanUpah(ctx context.Context, pekerja *domain.Pekerja, upah *domain.Upah) error {
	return domain.ErrBelumDiimplementasi
}

// KodeSudahDipakai memeriksa apakah kode pekerja sudah dipakai pekerja lain.
//
// Masukan: context, kode, dan ID pekerja yang dikecualikan (0 saat tambah).
// Hasil: true bila kode sudah dipakai.
// Aturan: 11.
// Tabel: pekerja.
// Memanggil: database/sql QueryRowContext.
func (r *Repository) KodeSudahDipakai(ctx context.Context, kode string, kecualiID int64) (bool, error) {
	return false, domain.ErrBelumDiimplementasi
}

// UpahPadaTanggal memeriksa apakah sudah ada upah pada tanggal berlaku tertentu.
//
// Masukan: context, ID pekerja, tanggal berlaku.
// Hasil: true bila sudah ada.
// Aturan: 8.
// Tabel: riwayat_upah.
// Memanggil: database/sql QueryRowContext.
func (r *Repository) UpahPadaTanggal(ctx context.Context, pekerjaID int64, tanggal time.Time) (bool, error) {
	return false, domain.ErrBelumDiimplementasi
}
