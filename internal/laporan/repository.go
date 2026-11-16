// Pemilik: Fadli | Modul 6 | Aturan: 13, 17
package laporan

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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

// AmbilSlip mengambil slip satu pekerja pada periode beserta data pekerja dan periode.
//
// Masukan: context, ID periode, ID pekerja.
// Hasil: SlipGaji beserta pekerja dan periode.
// Aturan: 13, 17.
// Tabel: slip_gaji, pekerja, periode_gaji.
// Memanggil: database/sql QueryRowContext.
func (r *Repository) AmbilSlip(ctx context.Context, periodeID, pekerjaID int64) (domain.SlipGaji, error) {
	var s domain.SlipGaji
	var pekerja domain.Pekerja
	var periode domain.PeriodeGaji
	var tanggalBerhenti sql.NullTime
	var ditutupPada sql.NullTime
	var hariKerjaPeriode sql.NullInt32

	err := r.db.QueryRowContext(ctx, `
		SELECT
			sg.id, sg.periode_id, sg.pekerja_id,
			sg.upah_pokok, sg.total_lembur, sg.potongan_telat, sg.total_koreksi,
			sg.gaji_bruto, sg.potongan_kasbon, sg.gaji_bersih,
			sg.hari_kerja_dibayar, sg.hari_kerja_periode, sg.dibuat_pada,
			p.id, p.kode, p.nama, p.jenis, p.tanggal_mulai, p.tanggal_berhenti, p.aktif,
			pg.id, pg.tipe, pg.tanggal_mulai, pg.tanggal_selesai, pg.status, pg.ditutup_pada
		FROM slip_gaji sg
		JOIN pekerja p ON p.id = sg.pekerja_id
		JOIN periode_gaji pg ON pg.id = sg.periode_id
		WHERE sg.periode_id = ? AND sg.pekerja_id = ?
		LIMIT 1`,
		periodeID, pekerjaID,
	).Scan(
		&s.ID, &s.PeriodeID, &s.PekerjaID,
		&s.UpahPokok, &s.TotalLembur, &s.PotonganTelat, &s.TotalKoreksi,
		&s.GajiBruto, &s.PotonganKasbon, &s.GajiBersih,
		&s.HariKerjaDibayar, &hariKerjaPeriode, &s.DibuatPada,
		&pekerja.ID, &pekerja.Kode, &pekerja.Nama, &pekerja.Jenis,
		&pekerja.TanggalMulai, &tanggalBerhenti, &pekerja.Aktif,
		&periode.ID, &periode.Tipe, &periode.TanggalMulai, &periode.TanggalSelesai,
		&periode.Status, &ditutupPada,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.SlipGaji{}, domain.ErrTidakDitemukan
	}
	if err != nil {
		return domain.SlipGaji{}, fmt.Errorf("ambil slip: %w", err)
	}
	if tanggalBerhenti.Valid {
		pekerja.TanggalBerhenti = &tanggalBerhenti.Time
	}
	if ditutupPada.Valid {
		periode.DitutupPada = &ditutupPada.Time
	}
	if hariKerjaPeriode.Valid {
		s.HariKerjaPeriode = int(hariKerjaPeriode.Int32)
	}
	s.Pekerja = &pekerja
	s.Periode = &periode
	return s, nil
}

// DaftarSlip mengambil seluruh slip pada periode beserta data pekerja.
//
// Masukan: context dan ID periode.
// Hasil: daftar slip.
// Aturan: 17.
// Tabel: slip_gaji, pekerja.
// Memanggil: database/sql QueryContext.
func (r *Repository) DaftarSlip(ctx context.Context, periodeID int64) ([]domain.SlipGaji, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			sg.id, sg.periode_id, sg.pekerja_id,
			sg.upah_pokok, sg.total_lembur, sg.potongan_telat, sg.total_koreksi,
			sg.gaji_bruto, sg.potongan_kasbon, sg.gaji_bersih,
			sg.hari_kerja_dibayar, sg.hari_kerja_periode, sg.dibuat_pada,
			p.id, p.kode, p.nama, p.jenis, p.tanggal_mulai, p.tanggal_berhenti, p.aktif
		FROM slip_gaji sg
		JOIN pekerja p ON p.id = sg.pekerja_id
		WHERE sg.periode_id = ?
		ORDER BY p.jenis, p.nama`,
		periodeID,
	)
	if err != nil {
		return nil, fmt.Errorf("daftar slip: %w", err)
	}
	defer rows.Close()

	var hasil []domain.SlipGaji
	for rows.Next() {
		var s domain.SlipGaji
		var pekerja domain.Pekerja
		var tanggalBerhenti sql.NullTime
		var hariKerjaPeriode sql.NullInt32
		if err := rows.Scan(
			&s.ID, &s.PeriodeID, &s.PekerjaID,
			&s.UpahPokok, &s.TotalLembur, &s.PotonganTelat, &s.TotalKoreksi,
			&s.GajiBruto, &s.PotonganKasbon, &s.GajiBersih,
			&s.HariKerjaDibayar, &hariKerjaPeriode, &s.DibuatPada,
			&pekerja.ID, &pekerja.Kode, &pekerja.Nama, &pekerja.Jenis,
			&pekerja.TanggalMulai, &tanggalBerhenti, &pekerja.Aktif,
		); err != nil {
			return nil, fmt.Errorf("scan slip: %w", err)
		}
		if tanggalBerhenti.Valid {
			pekerja.TanggalBerhenti = &tanggalBerhenti.Time
		}
		if hariKerjaPeriode.Valid {
			s.HariKerjaPeriode = int(hariKerjaPeriode.Int32)
		}
		s.Pekerja = &pekerja
		hasil = append(hasil, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterasi daftar slip: %w", err)
	}
	return hasil, nil
}

// Rekap mengambil total rekap satu periode dari baris slip yang sudah tersimpan.
//
// Masukan: context dan ID periode.
// Hasil: RekapGaji.
// Aturan: 12, 13.
// Tabel: slip_gaji.
// Memanggil: database/sql QueryRowContext.
func (r *Repository) Rekap(ctx context.Context, periodeID int64) (domain.RekapGaji, error) {
	var rekap domain.RekapGaji
	rekap.PeriodeID = periodeID

	err := r.db.QueryRowContext(ctx, `
		SELECT
			COALESCE(SUM(gaji_bruto), 0),
			COALESCE(SUM(total_lembur), 0),
			COALESCE(SUM(potongan_kasbon), 0),
			COALESCE(SUM(gaji_bersih), 0),
			COUNT(*)
		FROM slip_gaji
		WHERE periode_id = ?`,
		periodeID,
	).Scan(
		&rekap.TotalBruto,
		&rekap.TotalLembur,
		&rekap.TotalPotonganKasbon,
		&rekap.TotalBersih,
		&rekap.JumlahSlip,
	)
	if err != nil {
		return domain.RekapGaji{}, fmt.Errorf("rekap periode: %w", err)
	}
	return rekap, nil
}

// DaftarRincianSlip mengambil rincian upah harian berdasarkan ID slip.
func (r *Repository) DaftarRincianSlip(ctx context.Context, slipID int64) ([]domain.SlipGajiDetail, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id, slip_id, absensi_id, tanggal,
			upah_pokok_hari, upah_per_jam,
			menit_telat, potongan_telat,
			menit_lembur, upah_lembur
		FROM slip_gaji_detail
		WHERE slip_id = ?
		ORDER BY tanggal ASC`,
		slipID,
	)
	if err != nil {
		return nil, fmt.Errorf("daftar rincian slip: %w", err)
	}
	defer rows.Close()

	var hasil []domain.SlipGajiDetail

	for rows.Next() {
		var d domain.SlipGajiDetail

		if err := rows.Scan(
			&d.ID,
			&d.SlipID,
			&d.AbsensiID,
			&d.Tanggal,
			&d.UpahPokokHari,
			&d.UpahPerJam,
			&d.MenitTelat,
			&d.PotonganTelat,
			&d.MenitLembur,
			&d.UpahLembur,
		); err != nil {
			return nil, fmt.Errorf("scan rincian slip: %w", err)
		}

		hasil = append(hasil, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterasi rincian slip: %w", err)
	}

	return hasil, nil
}
