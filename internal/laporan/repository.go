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

// AmbilStatusPeriode mengambil status satu periode berdasarkan ID-nya.
// Digunakan oleh service untuk memvalidasi status sebelum menampilkan slip atau rekap.
//
// Masukan: context dan ID periode.
// Hasil: StatusPeriode.
// Tabel: periode_gaji.
func (r *Repository) AmbilStatusPeriode(ctx context.Context, periodeID int64) (domain.StatusPeriode, error) {
	var status domain.StatusPeriode
	err := r.db.QueryRowContext(ctx,
		`SELECT status FROM periode_gaji WHERE id = ? LIMIT 1`,
		periodeID,
	).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return "", domain.ErrTidakDitemukan
	}
	if err != nil {
		return "", fmt.Errorf("ambil status periode: %w", err)
	}
	return status, nil
}

// DaftarPeriodeTerbaru mengambil daftar periode terbaru.
//
// Masukan: context dan limit jumlah periode.
// Hasil: daftar PeriodeGaji urut dari terbaru.
// Tabel: periode_gaji.
func (r *Repository) DaftarPeriodeTerbaru(ctx context.Context, limit int) ([]domain.PeriodeGaji, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tipe, tanggal_mulai, tanggal_selesai, status, ditutup_pada
		FROM periode_gaji
		ORDER BY tanggal_mulai DESC
		LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("daftar periode terbaru: %w", err)
	}
	defer rows.Close()

	var hasil []domain.PeriodeGaji
	for rows.Next() {
		var p domain.PeriodeGaji
		var ditutupPada sql.NullTime
		if err := rows.Scan(&p.ID, &p.Tipe, &p.TanggalMulai, &p.TanggalSelesai, &p.Status, &ditutupPada); err != nil {
			return nil, fmt.Errorf("scan periode: %w", err)
		}
		if ditutupPada.Valid {
			p.DitutupPada = &ditutupPada.Time
		}
		hasil = append(hasil, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterasi periode: %w", err)
	}
	return hasil, nil
}

// HitungAbsensiBelumLengkap menghitung jumlah absensi yang belum memiliki jam_keluar
// dalam rentang tanggal periode tertentu.
//
// Masukan: context dan ID periode.
// Hasil: jumlah absensi belum lengkap.
// Tabel: absensi, jadwal_shift, periode_gaji.
func (r *Repository) HitungAbsensiBelumLengkap(ctx context.Context, periodeID int64) (int, error) {
	var jumlah int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM absensi a
		JOIN jadwal_shift js ON js.id = a.jadwal_shift_id
		JOIN periode_gaji pg ON pg.id = ?
		WHERE js.tanggal BETWEEN pg.tanggal_mulai AND pg.tanggal_selesai
		  AND a.jam_keluar IS NULL`,
		periodeID,
	).Scan(&jumlah)
	if err != nil {
		return 0, fmt.Errorf("hitung absensi belum lengkap: %w", err)
	}
	return jumlah, nil
}

// TotalGajiBersihPeriode menghitung total gaji bersih pada satu periode.
//
// Masukan: context dan ID periode.
// Hasil: total Rupiah gaji bersih.
// Tabel: slip_gaji.
func (r *Repository) TotalGajiBersihPeriode(ctx context.Context, periodeID int64) (domain.Rupiah, error) {
	var total domain.Rupiah
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(gaji_bersih), 0) FROM slip_gaji WHERE periode_id = ?`,
		periodeID,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("total gaji bersih: %w", err)
	}
	return total, nil
}

// TotalSisaKasbon menghitung total sisa kasbon seluruh pekerja.
//
// Masukan: context.
// Hasil: total Rupiah sisa kasbon.
// Tabel: kasbon.
func (r *Repository) TotalSisaKasbon(ctx context.Context) (domain.Rupiah, error) {
	var total domain.Rupiah
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(sisa), 0) FROM kasbon WHERE sisa > 0`,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("total sisa kasbon: %w", err)
	}
	return total, nil
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
