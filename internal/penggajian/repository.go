// Pemilik: Farah | Modul 4 | Aturan: 5, 6, 7, 8, 9, 12, 17, 18
package penggajian

import (
	"context"
	"database/sql"

	"penggajian-konveksi/internal/domain"
)

// Repository mengakses tabel slip_gaji, slip_gaji_detail, dan potongan_kasbon.
type Repository struct {
	db *sql.DB
}

// NewRepository membuat repository penggajian.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// SimpanSlipDanRincian menyimpan slip, rincian, dan potongan dalam satu transaksi.
//
// Masukan: context, slip, rincian harian, potongan kasbon.
// Hasil: error.
// Aturan: 12, 13, 18.
// Tabel: slip_gaji, slip_gaji_detail, potongan_kasbon, periode_gaji.
// Memanggil: database/sql BeginTx, ExecContext.
func (r *Repository) SimpanSlipDanRincian(ctx context.Context, slip domain.SlipGaji, rincian []domain.SlipGajiDetail, potongan []domain.PotonganKasbon) error {
	return domain.ErrBelumDiimplementasi
}

// AmbilSlip mengambil slip dan rincian satu pekerja.
//
// Masukan: context, ID periode, ID pekerja.
// Hasil: slip dan rincian harian.
// Aturan: 17.
// Tabel: slip_gaji, slip_gaji_detail, pekerja, periode_gaji.
// Memanggil: database/sql QueryRowContext, QueryContext.
func (r *Repository) AmbilSlip(ctx context.Context, periodeID, pekerjaID int64) (domain.SlipGaji, []domain.SlipGajiDetail, error) {
	return domain.SlipGaji{}, nil, domain.ErrBelumDiimplementasi
}

// DaftarSlip mengambil seluruh slip pada periode.
//
// Masukan: context dan ID periode.
// Hasil: daftar slip.
// Aturan: 17.
// Tabel: slip_gaji, pekerja.
// Memanggil: database/sql QueryContext.
func (r *Repository) DaftarSlip(ctx context.Context, periodeID int64) ([]domain.SlipGaji, error) {
	return nil, domain.ErrBelumDiimplementasi
}
