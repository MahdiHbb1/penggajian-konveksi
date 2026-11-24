// Pemilik: Edgar | Modul 2 | Aturan: 3, 14, 19
package jadwal

import (
	"context"
	"database/sql"
	"errors"
	"time"

	mysql "github.com/go-sql-driver/mysql"

	"penggajian-konveksi/internal/domain"
)

var errDatabaseTidakSiap = errors.New("database jadwal belum tersedia")

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

func tipePeriodeUntukJenis(jenis domain.JenisPekerja) domain.TipePeriode {
	if jenis == domain.PekerjaBulanan {
		return domain.PeriodeBulanan
	}
	return domain.PeriodeMingguan
}

func adalahDuplikatMySQL(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

// DaftarShift mengambil definisi shift.
func (r *Repository) DaftarShift(ctx context.Context) ([]domain.Shift, error) {
	if r == nil || r.db == nil {
		return nil, errDatabaseTidakSiap
	}
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
	if r == nil || r.db == nil {
		return false, errDatabaseTidakSiap
	}
	var ada bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM shift WHERE id = ?)`, shiftID).Scan(&ada)
	return ada, err
}

// PekerjaAktif memeriksa apakah pekerja ada dan berstatus aktif.
func (r *Repository) PekerjaAktif(ctx context.Context, pekerjaID int64) (bool, error) {
	aktif, _, err := r.PekerjaAktifDanJenis(ctx, pekerjaID)
	return aktif, err
}

// PekerjaAktifDanJenis mengambil status aktif dan jenis pekerja dalam satu query.
func (r *Repository) PekerjaAktifDanJenis(ctx context.Context, pekerjaID int64) (bool, domain.JenisPekerja, error) {
	if r == nil || r.db == nil {
		return false, "", errDatabaseTidakSiap
	}
	var (
		aktif bool
		jenis string
	)
	err := r.db.QueryRowContext(ctx,
		`SELECT aktif, jenis FROM pekerja WHERE id = ?`, pekerjaID).Scan(&aktif, &jenis)
	if errors.Is(err, sql.ErrNoRows) {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	return aktif, domain.JenisPekerja(jenis), nil
}

// PekerjaAktifSampai mengambil pekerja aktif yang sudah mulai bekerja
// paling lambat pada tanggal akhir pekan, untuk baris grid.
func (r *Repository) PekerjaAktifSampai(ctx context.Context, akhirPekan time.Time) ([]domain.Pekerja, error) {
	if r == nil || r.db == nil {
		return nil, errDatabaseTidakSiap
	}
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
func (r *Repository) JadwalRentang(ctx context.Context, mulai, selesai time.Time) ([]SelJadwal, error) {
	if r == nil || r.db == nil {
		return nil, errDatabaseTidakSiap
	}
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
func (r *Repository) JadwalByID(ctx context.Context, id int64) (SelJadwal, error) {
	if r == nil || r.db == nil {
		return SelJadwal{}, errDatabaseTidakSiap
	}
	var s SelJadwal
	err := r.db.QueryRowContext(ctx, selectSel+` WHERE j.id = ?`, id).
		Scan(&s.ID, &s.PekerjaID, &s.ShiftID, &s.Tanggal, &s.PunyaAbsensi)
	if errors.Is(err, sql.ErrNoRows) {
		return SelJadwal{}, ErrJadwalTidakAda
	}
	return s, err
}

// CariJadwalID mencari ID jadwal pekerja pada tanggal tertentu.
func (r *Repository) CariJadwalID(ctx context.Context, pekerjaID int64, tanggal time.Time) (int64, bool, error) {
	if r == nil || r.db == nil {
		return 0, false, errDatabaseTidakSiap
	}
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

// SimpanJadwal menyimpan satu jadwal pekerja; duplikat MySQL 1062 diterjemahkan ke ErrJadwalGanda.
func (r *Repository) SimpanJadwal(ctx context.Context, jadwal *domain.JadwalShift) error {
	if r == nil || r.db == nil {
		return errDatabaseTidakSiap
	}
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO jadwal_shift (pekerja_id, shift_id, tanggal) VALUES (?, ?, ?)`,
		jadwal.PekerjaID, jadwal.ShiftID, jadwal.Tanggal.Format(formatTanggalSQL))
	if err != nil {
		if adalahDuplikatMySQL(err) {
			return ErrJadwalGanda
		}
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	jadwal.ID = id
	return nil
}

// UbahJadwalAtomik mengunci baris jadwal_shift (FOR UPDATE), memverifikasi belum ada absensi
// dan periode tidak terkunci sesuai jenis pekerja, lalu memperbarui shift dalam satu transaksi.
func (r *Repository) UbahJadwalAtomik(ctx context.Context, id, shiftID int64) error {
	if r == nil || r.db == nil {
		return errDatabaseTidakSiap
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var (
		tanggal time.Time
		jenis   domain.JenisPekerja
	)
	err = tx.QueryRowContext(ctx, `
		SELECT j.tanggal, p.jenis
		FROM jadwal_shift j
		JOIN pekerja p ON p.id = j.pekerja_id
		WHERE j.id = ?
		FOR UPDATE`, id,
	).Scan(&tanggal, &jenis)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrJadwalTidakAda
	}
	if err != nil {
		return err
	}

	var punyaAbsensi bool
	if err := tx.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM absensi WHERE jadwal_shift_id = ?)`, id,
	).Scan(&punyaAbsensi); err != nil {
		return err
	}
	if punyaAbsensi {
		return ErrSudahAdaAbsensi
	}

	var terkunci bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM periode_gaji
			WHERE status = 'DITUTUP'
			  AND tipe = ?
			  AND ? BETWEEN tanggal_mulai AND tanggal_selesai
		)`, tipePeriodeUntukJenis(jenis), tanggal.Format(formatTanggalSQL),
	).Scan(&terkunci); err != nil {
		return err
	}
	if terkunci {
		return ErrPeriodeTertutup
	}

	var shiftAda bool
	if err := tx.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM shift WHERE id = ?)`, shiftID,
	).Scan(&shiftAda); err != nil {
		return err
	}
	if !shiftAda {
		return ErrShiftTidakValid
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE jadwal_shift SET shift_id = ? WHERE id = ?`, shiftID, id); err != nil {
		return err
	}
	return tx.Commit()
}

// BatalkanJadwalAtomik mengunci baris jadwal_shift (FOR UPDATE), memverifikasi belum ada absensi
// dan periode tidak terkunci sesuai jenis pekerja, lalu menghapus jadwal dalam satu transaksi.
func (r *Repository) BatalkanJadwalAtomik(ctx context.Context, id int64) error {
	if r == nil || r.db == nil {
		return errDatabaseTidakSiap
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var (
		tanggal time.Time
		jenis   domain.JenisPekerja
	)
	err = tx.QueryRowContext(ctx, `
		SELECT j.tanggal, p.jenis
		FROM jadwal_shift j
		JOIN pekerja p ON p.id = j.pekerja_id
		WHERE j.id = ?
		FOR UPDATE`, id,
	).Scan(&tanggal, &jenis)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrJadwalTidakAda
	}
	if err != nil {
		return err
	}

	var punyaAbsensi bool
	if err := tx.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM absensi WHERE jadwal_shift_id = ?)`, id,
	).Scan(&punyaAbsensi); err != nil {
		return err
	}
	if punyaAbsensi {
		return ErrSudahAdaAbsensi
	}

	var terkunci bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM periode_gaji
			WHERE status = 'DITUTUP'
			  AND tipe = ?
			  AND ? BETWEEN tanggal_mulai AND tanggal_selesai
		)`, tipePeriodeUntukJenis(jenis), tanggal.Format(formatTanggalSQL),
	).Scan(&terkunci); err != nil {
		return err
	}
	if terkunci {
		return ErrPeriodeTertutup
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM jadwal_shift WHERE id = ?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

// SalinJadwalBatch menyisipkan daftar jadwal baru dalam satu transaksi secara aman terhadap konflik
// (INSERT IGNORE agar tidak menimpa dan tidak gagal di tengah jalan saat salin bersamaan).
func (r *Repository) SalinJadwalBatch(ctx context.Context, daftar []domain.JadwalShift) (int, int, error) {
	if r == nil || r.db == nil {
		return 0, 0, errDatabaseTidakSiap
	}
	if len(daftar) == 0 {
		return 0, 0, nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, err
	}
	defer func() { _ = tx.Rollback() }()

	disalin, dilewati := 0, 0
	for _, j := range daftar {
		res, err := tx.ExecContext(ctx,
			`INSERT IGNORE INTO jadwal_shift (pekerja_id, shift_id, tanggal) VALUES (?, ?, ?)`,
			j.PekerjaID, j.ShiftID, j.Tanggal.Format(formatTanggalSQL))
		if err != nil {
			return 0, 0, err
		}
		n, _ := res.RowsAffected()
		if n > 0 {
			disalin++
		} else {
			dilewati++
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, 0, err
	}
	return disalin, dilewati, nil
}

// UbahJadwal mengganti shift pada jadwal yang sudah ada.
func (r *Repository) UbahJadwal(ctx context.Context, jadwal domain.JadwalShift) error {
	if r == nil || r.db == nil {
		return errDatabaseTidakSiap
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE jadwal_shift SET shift_id = ? WHERE id = ?`, jadwal.ShiftID, jadwal.ID)
	return err
}

// BatalkanJadwal menghapus jadwal yang belum punya absensi.
func (r *Repository) BatalkanJadwal(ctx context.Context, id int64) error {
	if r == nil || r.db == nil {
		return errDatabaseTidakSiap
	}
	_, err := r.db.ExecContext(ctx, `DELETE FROM jadwal_shift WHERE id = ?`, id)
	return err
}

// AmbilJadwal mengambil jadwal satu pekerja pada satu tanggal beserta shift.
func (r *Repository) AmbilJadwal(ctx context.Context, pekerjaID int64, tanggal time.Time) (domain.JadwalShift, error) {
	if r == nil || r.db == nil {
		return domain.JadwalShift{}, errDatabaseTidakSiap
	}
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
func (r *Repository) TanggalTerkunci(ctx context.Context, tanggal time.Time) (bool, error) {
	if r == nil || r.db == nil {
		return false, errDatabaseTidakSiap
	}
	var terkunci bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(
		   SELECT 1 FROM periode_gaji
		   WHERE status = 'DITUTUP' AND ? BETWEEN tanggal_mulai AND tanggal_selesai
		 )`,
		tanggal.Format(formatTanggalSQL)).Scan(&terkunci)
	return terkunci, err
}

// TanggalTerkunciJenis memeriksa apakah tanggal ada di periode DITUTUP sesuai jenis pekerja.
func (r *Repository) TanggalTerkunciJenis(ctx context.Context, jenis domain.JenisPekerja, tanggal time.Time) (bool, error) {
	if r == nil || r.db == nil {
		return false, errDatabaseTidakSiap
	}
	var terkunci bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(
		   SELECT 1 FROM periode_gaji
		   WHERE status = 'DITUTUP'
		     AND tipe = ?
		     AND ? BETWEEN tanggal_mulai AND tanggal_selesai
		 )`,
		tipePeriodeUntukJenis(jenis),
		tanggal.Format(formatTanggalSQL)).Scan(&terkunci)
	return terkunci, err
}
