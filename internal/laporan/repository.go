// Pemilik: Fadli | Modul 6 | Aturan: 13, 17
package laporan

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"penggajian-konveksi/internal/domain"
)

var errDatabaseTidakSiap = errors.New("database laporan belum tersedia")

// Repository mengakses slip dan rekap.
type Repository struct {
	db *sql.DB
}

// NewRepository membuat repository laporan.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// AmbilSlip mengambil slip satu pekerja pada periode beserta data pekerja dan periode.
func (r *Repository) AmbilSlip(ctx context.Context, periodeID, pekerjaID int64) (domain.SlipGaji, error) {
	if r == nil || r.db == nil {
		return domain.SlipGaji{}, errDatabaseTidakSiap
	}
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
func (r *Repository) DaftarSlip(ctx context.Context, periodeID int64) ([]domain.SlipGaji, error) {
	if r == nil || r.db == nil {
		return nil, errDatabaseTidakSiap
	}
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
func (r *Repository) Rekap(ctx context.Context, periodeID int64) (domain.RekapGaji, error) {
	if r == nil || r.db == nil {
		return domain.RekapGaji{}, errDatabaseTidakSiap
	}
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
func (r *Repository) AmbilStatusPeriode(ctx context.Context, periodeID int64) (domain.StatusPeriode, error) {
	if r == nil || r.db == nil {
		return "", errDatabaseTidakSiap
	}
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
func (r *Repository) DaftarPeriodeTerbaru(ctx context.Context, limit int) ([]domain.PeriodeGaji, error) {
	if r == nil || r.db == nil {
		return nil, errDatabaseTidakSiap
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tipe, tanggal_mulai, tanggal_selesai, status, ditutup_pada
		FROM periode_gaji
		ORDER BY tanggal_mulai DESC, id DESC
		LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("daftar periode terbaru: %w", err)
	}
	defer rows.Close()

	return scanDaftarPeriode(rows)
}

// DaftarPeriodeSiapLaporan mengambil periode berstatus DIHITUNG atau DITUTUP langsung dari SQL.
func (r *Repository) DaftarPeriodeSiapLaporan(ctx context.Context, limit int) ([]domain.PeriodeGaji, error) {
	if r == nil || r.db == nil {
		return nil, errDatabaseTidakSiap
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tipe, tanggal_mulai, tanggal_selesai, status, ditutup_pada
		FROM periode_gaji
		WHERE status IN ('DIHITUNG', 'DITUTUP')
		ORDER BY tanggal_mulai DESC, id DESC
		LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("daftar periode siap laporan: %w", err)
	}
	defer rows.Close()

	return scanDaftarPeriode(rows)
}

func scanDaftarPeriode(rows *sql.Rows) ([]domain.PeriodeGaji, error) {
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

// HitungAbsensiBelumLengkap menghitung jumlah absensi tanpa jam_keluar pada periode tertentu
// yang sesuai dengan jenis pekerja tipe periode tersebut (MINGGUAN->HARIAN, BULANAN->BULANAN).
func (r *Repository) HitungAbsensiBelumLengkap(ctx context.Context, periodeID int64) (int, error) {
	if r == nil || r.db == nil {
		return 0, errDatabaseTidakSiap
	}
	var jumlah int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM absensi a
		JOIN jadwal_shift js ON js.id = a.jadwal_shift_id
		JOIN pekerja p ON p.id = js.pekerja_id
		JOIN periode_gaji pg ON pg.id = ?
		WHERE js.tanggal BETWEEN pg.tanggal_mulai AND pg.tanggal_selesai
		  AND p.jenis = CASE WHEN pg.tipe = 'MINGGUAN' THEN 'HARIAN' ELSE 'BULANAN' END
		  AND a.jam_keluar IS NULL`,
		periodeID,
	).Scan(&jumlah)
	if err != nil {
		return 0, fmt.Errorf("hitung absensi belum lengkap: %w", err)
	}
	return jumlah, nil
}

// TotalGajiBersihPeriode menghitung total gaji bersih pada satu periode.
func (r *Repository) TotalGajiBersihPeriode(ctx context.Context, periodeID int64) (domain.Rupiah, error) {
	if r == nil || r.db == nil {
		return 0, errDatabaseTidakSiap
	}
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
func (r *Repository) TotalSisaKasbon(ctx context.Context) (domain.Rupiah, error) {
	if r == nil || r.db == nil {
		return 0, errDatabaseTidakSiap
	}
	var total domain.Rupiah
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(sisa), 0) FROM kasbon WHERE sisa > 0`,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("total sisa kasbon: %w", err)
	}
	return total, nil
}

// SisaKasbonPekerja menghitung total sisa kasbon berjalan milik satu pekerja untuk ditampilkan di slip.
func (r *Repository) SisaKasbonPekerja(ctx context.Context, pekerjaID int64) (domain.Rupiah, error) {
	if r == nil || r.db == nil {
		return 0, errDatabaseTidakSiap
	}
	var total domain.Rupiah
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(sisa), 0) FROM kasbon WHERE pekerja_id = ? AND sisa > 0`,
		pekerjaID,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("sisa kasbon pekerja: %w", err)
	}
	return total, nil
}

// DaftarRincianSlip mengambil rincian upah harian berdasarkan ID slip.
func (r *Repository) DaftarRincianSlip(ctx context.Context, slipID int64) ([]domain.SlipGajiDetail, error) {
	if r == nil || r.db == nil {
		return nil, errDatabaseTidakSiap
	}
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
			&d.ID, &d.SlipID, &d.AbsensiID, &d.Tanggal,
			&d.UpahPokokHari, &d.UpahPerJam,
			&d.MenitTelat, &d.PotonganTelat,
			&d.MenitLembur, &d.UpahLembur,
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
