// Pemilik: Zahra | Modul 3 | Aturan: 2, 3, 5, 6, 14
package absensi

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"penggajian-konveksi/internal/domain"
)

// Repository mengakses tabel absensi dan membaca jadwal/periode terkait.
type Repository struct {
	db *sql.DB
}

// NewRepository membuat repository absensi.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// SimpanAbsensi menyimpan absensi baru. Kunci unik jadwal_shift_id di database
// menjadi perlindungan terakhir terhadap dua request yang masuk bersamaan.
func (r *Repository) SimpanAbsensi(ctx context.Context, data *domain.Absensi) error {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO absensi (jadwal_shift_id, jam_masuk, jam_keluar) VALUES (?, ?, ?)`,
		data.JadwalShiftID, data.JamMasuk, nilaiJamKeluar(data.JamKeluar),
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate entry") ||
			strings.Contains(err.Error(), "Error 1062") {
			return ErrAbsensiGanda
		}
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	data.ID = id
	return nil
}

// UbahAbsensi memperbarui jam masuk dan jam keluar.
func (r *Repository) UbahAbsensi(ctx context.Context, data domain.Absensi) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE absensi SET jam_masuk = ?, jam_keluar = ? WHERE id = ?`,
		data.JamMasuk, nilaiJamKeluar(data.JamKeluar), data.ID,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		var ada int
		if err := r.db.QueryRowContext(ctx, `SELECT 1 FROM absensi WHERE id = ?`, data.ID).Scan(&ada); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrAbsensiTidakDitemukan
			}
			return err
		}
	}
	return nil
}

// DaftarAbsensi mengambil absensi yang sudah tercatat pada satu tanggal.
func (r *Repository) DaftarAbsensi(ctx context.Context, tanggal time.Time) ([]domain.AbsensiLengkap, error) {
	query := selectAbsensiLengkap + `
		WHERE js.tanggal = ? AND a.id IS NOT NULL
		ORDER BY p.nama, s.nama`
	return r.queryAbsensiLengkap(ctx, query, tanggal.Format("2006-01-02"))
}

// DaftarJadwalAbsensi mengambil seluruh jadwal tanggal tertentu, termasuk baris
// yang belum mempunyai absensi (LEFT JOIN).
func (r *Repository) DaftarJadwalAbsensi(ctx context.Context, tanggal time.Time) ([]domain.AbsensiLengkap, error) {
	query := selectAbsensiLengkap + `
		WHERE js.tanggal = ?
		ORDER BY p.nama, s.nama`
	return r.queryAbsensiLengkap(ctx, query, tanggal.Format("2006-01-02"))
}

// DaftarBelumLengkap mengambil absensi tanpa jam keluar pada periode.
func (r *Repository) DaftarBelumLengkap(ctx context.Context, periodeID int64) ([]domain.AbsensiLengkap, error) {
	query := selectAbsensiLengkap + `
		JOIN periode_gaji pg ON js.tanggal BETWEEN pg.tanggal_mulai AND pg.tanggal_selesai
		WHERE pg.id = ? AND a.id IS NOT NULL AND a.jam_keluar IS NULL
		ORDER BY js.tanggal, p.nama`
	return r.queryAbsensiLengkap(ctx, query, periodeID)
}

// AbsensiDalamPeriode mengambil hanya absensi yang sudah mempunyai jam keluar.
func (r *Repository) AbsensiDalamPeriode(ctx context.Context, periodeID int64) ([]domain.AbsensiLengkap, error) {
	query := selectAbsensiLengkap + `
		JOIN periode_gaji pg ON js.tanggal BETWEEN pg.tanggal_mulai AND pg.tanggal_selesai
		WHERE pg.id = ? AND a.id IS NOT NULL AND a.jam_keluar IS NOT NULL
		ORDER BY js.tanggal, p.nama`
	return r.queryAbsensiLengkap(ctx, query, periodeID)
}

// AdaBelumLengkap mengembalikan flag dan jumlah absensi tanpa jam keluar.
func (r *Repository) AdaBelumLengkap(ctx context.Context, periodeID int64) (bool, int, error) {
	var jumlah int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM absensi a
		JOIN jadwal_shift js ON js.id = a.jadwal_shift_id
		JOIN periode_gaji pg ON js.tanggal BETWEEN pg.tanggal_mulai AND pg.tanggal_selesai
		WHERE pg.id = ? AND a.jam_keluar IS NULL`, periodeID,
	).Scan(&jumlah)
	if err != nil {
		return false, 0, err
	}
	return jumlah > 0, jumlah, nil
}

// AmbilJadwalByID mengambil jadwal dan definisi shift berdasarkan ID jadwal.
func (r *Repository) AmbilJadwalByID(ctx context.Context, jadwalID int64) (domain.JadwalShift, error) {
	var jadwal domain.JadwalShift
	var jamMulai, jamSelesai string
	err := r.db.QueryRowContext(ctx, `
		SELECT js.id, js.pekerja_id, js.shift_id, js.tanggal,
		       s.id, s.nama, TIME_FORMAT(s.jam_mulai, '%H:%i:%s'),
		       TIME_FORMAT(s.jam_selesai, '%H:%i:%s')
		FROM jadwal_shift js
		JOIN shift s ON s.id = js.shift_id
		WHERE js.id = ?`, jadwalID,
	).Scan(
		&jadwal.ID, &jadwal.PekerjaID, &jadwal.ShiftID, &jadwal.Tanggal,
		&jadwal.Shift.ID, &jadwal.Shift.Nama, &jamMulai, &jamSelesai,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.JadwalShift{}, ErrJadwalTidakDitemukan
		}
		return domain.JadwalShift{}, err
	}
	jadwal.Shift.JamMulai = jamMulai
	jadwal.Shift.JamSelesai = jamSelesai
	return jadwal, nil
}

// AmbilAbsensiByID mengambil absensi sekaligus jadwal yang menentukan tanggal kuncinya.
func (r *Repository) AmbilAbsensiByID(ctx context.Context, absensiID int64) (domain.Absensi, domain.JadwalShift, error) {
	var data domain.Absensi
	var jadwal domain.JadwalShift
	var jamKeluar sql.NullTime
	var jamMulai, jamSelesai string
	err := r.db.QueryRowContext(ctx, `
		SELECT a.id, a.jadwal_shift_id, a.jam_masuk, a.jam_keluar,
		       js.id, js.pekerja_id, js.shift_id, js.tanggal,
		       s.id, s.nama, TIME_FORMAT(s.jam_mulai, '%H:%i:%s'),
		       TIME_FORMAT(s.jam_selesai, '%H:%i:%s')
		FROM absensi a
		JOIN jadwal_shift js ON js.id = a.jadwal_shift_id
		JOIN shift s ON s.id = js.shift_id
		WHERE a.id = ?`, absensiID,
	).Scan(
		&data.ID, &data.JadwalShiftID, &data.JamMasuk, &jamKeluar,
		&jadwal.ID, &jadwal.PekerjaID, &jadwal.ShiftID, &jadwal.Tanggal,
		&jadwal.Shift.ID, &jadwal.Shift.Nama, &jamMulai, &jamSelesai,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Absensi{}, domain.JadwalShift{}, ErrAbsensiTidakDitemukan
		}
		return domain.Absensi{}, domain.JadwalShift{}, err
	}
	if jamKeluar.Valid {
		data.JamKeluar = &jamKeluar.Time
	}
	jadwal.Shift.JamMulai = jamMulai
	jadwal.Shift.JamSelesai = jamSelesai
	return data, jadwal, nil
}

// AdaAbsensiUntukJadwal memeriksa apakah satu jadwal sudah mempunyai absensi.
func (r *Repository) AdaAbsensiUntukJadwal(ctx context.Context, jadwalID int64) (bool, error) {
	var ada bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM absensi WHERE jadwal_shift_id = ?)`, jadwalID,
	).Scan(&ada)
	return ada, err
}

// PeriodeTerkunci memeriksa apakah tanggal masuk ke periode berstatus DITUTUP.
func (r *Repository) PeriodeTerkunci(ctx context.Context, tanggal time.Time) (bool, error) {
	var terkunci bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM periode_gaji
			WHERE status = 'DITUTUP'
			  AND ? BETWEEN tanggal_mulai AND tanggal_selesai
		)`, tanggal.Format("2006-01-02"),
	).Scan(&terkunci)
	return terkunci, err
}

// DaftarPeriode mengambil periode terbaru untuk pilihan filter halaman.
func (r *Repository) DaftarPeriode(ctx context.Context) ([]domain.PeriodeGaji, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tipe, tanggal_mulai, tanggal_selesai, status, ditutup_pada
		FROM periode_gaji
		ORDER BY tanggal_mulai DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hasil := make([]domain.PeriodeGaji, 0)
	for rows.Next() {
		var p domain.PeriodeGaji
		var ditutup sql.NullTime
		if err := rows.Scan(&p.ID, &p.Tipe, &p.TanggalMulai, &p.TanggalSelesai, &p.Status, &ditutup); err != nil {
			return nil, err
		}
		if ditutup.Valid {
			p.DitutupPada = &ditutup.Time
		}
		hasil = append(hasil, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return hasil, nil
}

const selectAbsensiLengkap = `
	SELECT
		a.id, a.jadwal_shift_id, a.jam_masuk, a.jam_keluar,
		js.id, js.pekerja_id, js.shift_id, js.tanggal,
		p.id, p.kode, p.nama, p.jenis, p.tanggal_mulai, p.tanggal_berhenti, p.aktif,
		s.id, s.nama, TIME_FORMAT(s.jam_mulai, '%H:%i:%s'),
		TIME_FORMAT(s.jam_selesai, '%H:%i:%s')
	FROM jadwal_shift js
	JOIN pekerja p ON p.id = js.pekerja_id
	JOIN shift s ON s.id = js.shift_id
	LEFT JOIN absensi a ON a.jadwal_shift_id = js.id
`

type rowScanner interface {
	Scan(dest ...any) error
}

func (r *Repository) queryAbsensiLengkap(ctx context.Context, query string, args ...any) ([]domain.AbsensiLengkap, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hasil := make([]domain.AbsensiLengkap, 0)
	for rows.Next() {
		item, err := scanAbsensiLengkap(rows)
		if err != nil {
			return nil, err
		}
		hasil = append(hasil, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return hasil, nil
}

func scanAbsensiLengkap(row rowScanner) (domain.AbsensiLengkap, error) {
	var item domain.AbsensiLengkap
	var absensiID, absensiJadwalID sql.NullInt64
	var jamMasuk, jamKeluar sql.NullTime
	var tanggalBerhenti sql.NullTime
	var jamMulai, jamSelesai string

	err := row.Scan(
		&absensiID, &absensiJadwalID, &jamMasuk, &jamKeluar,
		&item.Jadwal.ID, &item.Jadwal.PekerjaID, &item.Jadwal.ShiftID, &item.Jadwal.Tanggal,
		&item.Pekerja.ID, &item.Pekerja.Kode, &item.Pekerja.Nama, &item.Pekerja.Jenis,
		&item.Pekerja.TanggalMulai, &tanggalBerhenti, &item.Pekerja.Aktif,
		&item.Jadwal.Shift.ID, &item.Jadwal.Shift.Nama, &jamMulai, &jamSelesai,
	)
	if err != nil {
		return domain.AbsensiLengkap{}, err
	}
	item.Jadwal.Pekerja = item.Pekerja
	item.Jadwal.Shift.JamMulai = jamMulai
	item.Jadwal.Shift.JamSelesai = jamSelesai
	if tanggalBerhenti.Valid {
		item.Pekerja.TanggalBerhenti = &tanggalBerhenti.Time
		item.Jadwal.Pekerja.TanggalBerhenti = &tanggalBerhenti.Time
	}
	if absensiID.Valid {
		item.Absensi.ID = absensiID.Int64
		if absensiJadwalID.Valid {
			item.Absensi.JadwalShiftID = absensiJadwalID.Int64
		}
		if jamMasuk.Valid {
			item.Absensi.JamMasuk = jamMasuk.Time
		}
		if jamKeluar.Valid {
			item.Absensi.JamKeluar = &jamKeluar.Time
		}
	}
	return item, nil
}

func nilaiJamKeluar(jamKeluar *time.Time) any {
	if jamKeluar == nil {
		return nil
	}
	return *jamKeluar
}
