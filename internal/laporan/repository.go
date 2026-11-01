// Pemilik: Fadli | Modul 6 | Aturan: 13, 17
package laporan

import (
	"context"
	"database/sql"

	"penggajian-konveksi/internal/domain"
)

// Repository mengakses slip dan rekap.
type Repository struct {
	db *sql.DB
}

// NewRepository membuat repository laporan.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// AmbilSlip mengambil slip satu pekerja pada periode.
//
// Masukan: context, ID periode, ID pekerja.
// Hasil: SlipGaji beserta pekerja dan periode.
// Aturan: 13, 17.
// Tabel: slip_gaji, pekerja, periode_gaji.
// Memanggil: database/sql QueryRowContext.
func (r *Repository) AmbilSlip(ctx context.Context, periodeID, pekerjaID int64) (domain.SlipGaji, error) {
	return domain.SlipGaji{}, domain.ErrBelumDiimplementasi
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

// Rekap mengambil total rekap satu periode.
//
// Masukan: context dan ID periode.
// Hasil: RekapGaji.
// Aturan: 12, 13.
// Tabel: slip_gaji, potongan_kasbon.
// Memanggil: database/sql QueryRowContext.
func (r *Repository) Rekap(ctx context.Context, periodeID int64) (domain.RekapGaji, error) {
	return domain.RekapGaji{}, domain.ErrBelumDiimplementasi
}
