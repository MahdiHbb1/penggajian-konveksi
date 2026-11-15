// Pemilik: Edgar | Modul 2 | Aturan: 3, 14, 19
package jadwal

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"penggajian-konveksi/internal/domain"
)

// SelJadwal adalah satu sel grid: jadwal satu pekerja pada satu tanggal.
type SelJadwal struct {
	ID           int64
	PekerjaID    int64
	ShiftID      int64
	Tanggal      time.Time
	PunyaAbsensi bool
}

// Repository mengakses tabel shift dan jadwal_shift.
type Repository struct {
	db *sql.DB
}

// NewRepository membuat repository jadwal.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const formatTanggalSQL = "2006-01-02"

// DaftarShift mengambil definisi shift.
//
// Masukan: context.
// Hasil: daftar shift; jam berbentuk "07:00".
// Aturan: 3.
// Tabel: shift.
// Memanggil: database/sql QueryContext.
func (r *Repository) DaftarShift(ctx context.Context) ([]domain.Shift, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, nama, TIME_FORMAT(jam_mulai, '%H:%i'), TIME_FORMAT(jam_selesai, '%H:%i')
		 FROM shift ORDER BY jam_mulai ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []domain.Shift
	for rows.Next() {
		var s domain.Shift
		if err := rows.Scan(&s.ID, &s.Nama, &s.JamMulai, &s.JamSelesai); err != nil {
			return nil, err
		}
		hasil = append(hasil, s)
	}
	return hasil, rows.Err()
}

// ShiftAda memeriksa apakah shift dengan ID itu terdaftar.
func (r *Repository) ShiftAda(ctx context.Context, shiftID int64) (bool, error) {
	var ada bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM shift WHERE id = ?)`, shiftID).Scan(&ada)
	return ada, err
}

// PekerjaAktif memeriksa apakah pekerja ada dan berstatus aktif.
//
// Aturan: 16 (hanya pekerja aktif yang dapat dijadwalkan).
// Tabel: pekerja.
func (r *Repository) PekerjaAktif(ctx context.Context, pekerjaID int64) (bool, error) {
	var ada bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM pekerja WHERE id = ? AND aktif = TRUE)`, pekerjaID).Scan(&ada)
	return ada, err
}

// PekerjaAktifSampai mengambil pekerja aktif yang sudah mulai bekerja
// paling lambat pada tanggal akhir pekan, untuk baris grid.
//
// Tabel: pekerja.
func (r *Repository) PekerjaAktifSampai(ctx context.Context, akhirPekan time.Time) ([]domain.Pekerja, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, kode, nama, jenis, aktif FROM pekerja
		 WHERE aktif = TRUE AND tanggal_mulai <= ?
		 ORDER BY nama ASC`,
		akhirPekan.Format(formatTanggalSQL))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []domain.Pekerja
	for rows.Next() {
		var p domain.Pekerja
		var jenis string
		if err := rows.Scan(&p.ID, &p.Kode, &p.Nama, &jenis, &p.Aktif); err != nil {
			return nil, err
		}
		p.Jenis = domain.JenisPekerja(jenis)
		hasil = append(hasil, p)
	}
	return hasil, rows.Err()
}

const selectSel = `
	SELECT j.id, j.pekerja_id, j.shift_id, j.tanggal,
	       EXISTS(SELECT 1 FROM absensi a WHERE a.jadwal_shift_id = j.id)
	FROM jadwal_shift j`

// JadwalRentang mengambil semua jadwal pada rentang tanggal (ikut dihitung).
//
// Aturan: 3, 19.
// Tabel: jadwal_shift, absensi.
func (r *Repository) JadwalRentang(ctx context.Context, mulai, selesai time.Time) ([]SelJadwal, error) {
	rows, err := r.db.QueryContext(ctx,
		selectSel+` WHERE j.tanggal BETWEEN ? AND ? ORDER BY j.tanggal ASC, j.id ASC`,
		mulai.Format(formatTanggalSQL), selesai.Format(formatTanggalSQL))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []SelJadwal
	for rows.Next() {
		var s SelJadwal
		if err := rows.Scan(&s.ID, &s.PekerjaID, &s.ShiftID, &s.Tanggal, &s.PunyaAbsensi); err != nil {
			return nil, err
		}
		hasil = append(hasil, s)
	}
	return hasil, rows.Err()
}

// JadwalByID mengambil satu jadwal beserta penanda sudah punya absensi.
//
// Hasil: ErrJadwalTidakAda bila tidak ditemukan.
// Aturan: 14, 19.
// Tabel: jadwal_shift, absensi.
func (r *Repository) JadwalByID(ctx context.Context, id int64) (SelJadwal, error) {
	var s SelJadwal
	err := r.db.QueryRowContext(ctx, selectSel+` WHERE j.id = ?`, id).
		Scan(&s.ID, &s.PekerjaID, &s.ShiftID, &s.Tanggal, &s.PunyaAbsensi)
	if errors.Is(err, sql.ErrNoRows) {
		return SelJadwal{}, ErrJadwalTidakAda
	}
	return s, err
}

// CariJadwalID mencari ID jadwal pekerja pada tanggal tertentu.
//
// Hasil: ID dan penanda ada tidaknya jadwal.
func (r *Repository) CariJadwalID(ctx context.Context, pekerjaID int64, tanggal time.Time) (int64, bool, error) {
	var id int64
	err := r.db.QueryRowContext(ctx,
		`SELECT id FROM jadwal_shift WHERE pekerja_id = ? AND tanggal = ?`,
		pekerjaID, tanggal.Format(formatTanggalSQL)).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return id, true, nil
}

// SimpanJadwal menyimpan satu jadwal pekerja.
//
// Masukan: context dan pointer JadwalShift.
// Hasil: error; ID hasil insert diisi ke pointer.
// Aturan: 3, 19.
// Tabel: jadwal_shift.
// Memanggil: database/sql ExecContext.
func (r *Repository) SimpanJadwal(ctx context.Context, jadwal *domain.JadwalShift) error {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO jadwal_shift (pekerja_id, shift_id, tanggal) VALUES (?, ?, ?)`,
		jadwal.PekerjaID, jadwal.ShiftID, jadwal.Tanggal.Format(formatTanggalSQL))
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	jadwal.ID = id
	return nil
}

// UbahJadwal mengganti shift pada jadwal yang sudah ada.
// Pekerja dan tanggal tidak diubah.
//
// Aturan: 14, 19.
// Tabel: jadwal_shift.
func (r *Repository) UbahJadwal(ctx context.Context, jadwal domain.JadwalShift) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE jadwal_shift SET shift_id = ? WHERE id = ?`, jadwal.ShiftID, jadwal.ID)
	return err
}

// BatalkanJadwal menghapus jadwal yang belum punya absensi.
// Pemeriksaan absensi dan periode dilakukan Service sebelum memanggil ini.
//
// Aturan: 14, 19.
// Tabel: jadwal_shift.
func (r *Repository) BatalkanJadwal(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM jadwal_shift WHERE id = ?`, id)
	return err
}

// AmbilJadwal mengambil jadwal satu pekerja pada satu tanggal beserta shift.
//
// Hasil: JadwalShift; ErrJadwalTidakAda bila pekerja tidak dijadwalkan.
// Aturan: 3.
// Tabel: jadwal_shift, shift, pekerja.
func (r *Repository) AmbilJadwal(ctx context.Context, pekerjaID int64, tanggal time.Time) (domain.JadwalShift, error) {
	var j domain.JadwalShift
	var jenis string
	err := r.db.QueryRowContext(ctx,
		`SELECT j.id, j.pekerja_id, j.shift_id, j.tanggal,
		        s.nama, TIME_FORMAT(s.jam_mulai, '%H:%i'), TIME_FORMAT(s.jam_selesai, '%H:%i'),
		        p.kode, p.nama, p.jenis
		 FROM jadwal_shift j
		 JOIN shift s ON s.id = j.shift_id
		 JOIN pekerja p ON p.id = j.pekerja_id
		 WHERE j.pekerja_id = ? AND j.tanggal = ?`,
		pekerjaID, tanggal.Format(formatTanggalSQL),
	).Scan(&j.ID, &j.PekerjaID, &j.ShiftID, &j.Tanggal,
		&j.Shift.Nama, &j.Shift.JamMulai, &j.Shift.JamSelesai,
		&j.Pekerja.Kode, &j.Pekerja.Nama, &jenis)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.JadwalShift{}, ErrJadwalTidakAda
	}
	if err != nil {
		return domain.JadwalShift{}, err
	}
	j.Shift.ID = j.ShiftID
	j.Pekerja.ID = j.PekerjaID
	j.Pekerja.Jenis = domain.JenisPekerja(jenis)
	return j, nil
}

// TanggalTerkunci memeriksa langsung apakah tanggal ada di periode DITUTUP.
// Dipakai sebagai cadangan selama PeriodeProvider Modul 5 belum jadi.
//
// Aturan: 14.
// Tabel: periode_gaji.
func (r *Repository) TanggalTerkunci(ctx context.Context, tanggal time.Time) (bool, error) {
	var terkunci bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(
		   SELECT 1 FROM periode_gaji
		   WHERE status = 'DITUTUP' AND ? BETWEEN tanggal_mulai AND tanggal_selesai
		 )`,
		tanggal.Format(formatTanggalSQL)).Scan(&terkunci)
	return terkunci, err
}