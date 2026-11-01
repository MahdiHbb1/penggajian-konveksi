// Pemilik: Edgar | Modul 2 | Aturan: 12, 14, 16
package kasbon

import (
	"context"
	"database/sql"
	"time"

	"penggajian-konveksi/internal/domain"
)

// Filter adalah parameter daftar kasbon.
type Filter struct {
	PekerjaID *int64
	Status    string
}

// Repository mengakses tabel kasbon dan potongan_kasbon.
type Repository struct {
	db *sql.DB
}

// NewRepository membuat repository kasbon.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// SimpanKasbon menyimpan kasbon baru.
//
// Masukan: context dan pointer Kasbon.
// Hasil: error; ID hasil insert diisi ke pointer.
// Aturan: 12.
// Tabel: kasbon.
// Memanggil: database/sql ExecContext.
func (r *Repository) SimpanKasbon(ctx context.Context, kasbon *domain.Kasbon) error {
	return domain.ErrBelumDiimplementasi
}

// DaftarKasbon mengambil kasbon sesuai filter.
//
// Masukan: context dan Filter.
// Hasil: daftar kasbon.
// Aturan: 12, 16.
// Tabel: kasbon, pekerja.
// Memanggil: database/sql QueryContext.
func (r *Repository) DaftarKasbon(ctx context.Context, filter Filter) ([]domain.Kasbon, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// AmbilKasbon mengambil satu kasbon.
//
// Masukan: context dan ID kasbon.
// Hasil: Kasbon.
// Aturan: 12.
// Tabel: kasbon, pekerja.
// Memanggil: database/sql QueryRowContext.
func (r *Repository) AmbilKasbon(ctx context.Context, id int64) (domain.Kasbon, error) {
	return domain.Kasbon{}, domain.ErrBelumDiimplementasi
}

// SisaKasbon menjumlahkan sisa kasbon pekerja.
//
// Masukan: context dan ID pekerja.
// Hasil: total sisa.
// Aturan: 12, 16.
// Tabel: kasbon.
// Memanggil: database/sql QueryRowContext.
func (r *Repository) SisaKasbon(ctx context.Context, pekerjaID int64) (domain.Rupiah, error) {
	return 0, domain.ErrBelumDiimplementasi
}

// KasbonDiprosesSampai mengambil kasbon yang belum lunas sampai tanggal.
//
// Masukan: context, ID pekerja, batas tanggal.
// Hasil: kasbon terurut dari yang tertua.
// Aturan: 12.
// Tabel: kasbon.
// Memanggil: database/sql QueryContext.
func (r *Repository) KasbonDiprosesSampai(ctx context.Context, pekerjaID int64, tanggal time.Time) ([]domain.Kasbon, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// RiwayatPotongan mengambil potongan kasbon per periode.
//
// Masukan: context dan ID kasbon.
// Hasil: daftar potongan.
// Aturan: 12, 13.
// Tabel: potongan_kasbon, slip_gaji, periode_gaji.
// Memanggil: database/sql QueryContext.
func (r *Repository) RiwayatPotongan(ctx context.Context, kasbonID int64) ([]domain.PotonganKasbon, error) {
	return nil, domain.ErrBelumDiimplementasi
}
