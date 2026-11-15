// Pemilik: Farah | Modul 4 | Aturan: 5, 6, 7, 8, 9, 12, 17, 18
package penggajian

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"penggajian-konveksi/internal/domain"
)

const layoutTanggalSQL = "2006-01-02"

var errDatabaseTidakSiap = errors.New("Database belum tersedia.")

var _ Repositori = (*Repository)(nil)

// Repository mengakses tabel slip_gaji, slip_gaji_detail, dan potongan_kasbon.
// Query pendukung dipakai sebagai cadangan selama provider modul lain belum
// diimplementasikan.
type Repository struct {
	db *sql.DB
}

// NewRepository membuat repository penggajian.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// HasilHitung adalah satu slip beserta rincian dan potongan kasbonnya.
type HasilHitung struct {
	Slip     domain.SlipGaji
	Rincian  []domain.SlipGajiDetail
	Potongan []domain.PotonganKasbon
}

// RincianHarian adalah rincian slip dengan data absensi untuk halaman rincian.
type RincianHarian struct {
	domain.SlipGajiDetail
	NamaShift string
	JamMasuk  time.Time
	JamKeluar *time.Time
}

// SimpanSlipDanRincian menyimpan slip, rincian, dan potongan dalam satu transaksi.
//
// Masukan: context, slip, rincian harian, potongan kasbon.
// Hasil: error.
// Aturan: 12, 13, 18.
// Tabel: slip_gaji, slip_gaji_detail, potongan_kasbon.
// Memanggil: database/sql BeginTx, ExecContext.
func (r *Repository) SimpanSlipDanRincian(ctx context.Context, slip domain.SlipGaji, rincian []domain.SlipGajiDetail, potongan []domain.PotonganKasbon) error {
	if r.db == nil {
		return errDatabaseTidakSiap
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := simpanSatuSlip(ctx, tx, &slip, rincian, potongan); err != nil {
		return err
	}
	return tx.Commit()
}

// SimpanHasilPeriode menimpa seluruh slip satu periode dalam satu transaksi.
// Dipakai saat hitung dan hitung ulang agar hasil periode selalu konsisten.
//
// Masukan: context, ID periode, daftar hasil hitung.
// Hasil: error.
// Aturan: 13, 18.
// Tabel: slip_gaji, slip_gaji_detail, potongan_kasbon.
// Memanggil: database/sql BeginTx, ExecContext.
func (r *Repository) SimpanHasilPeriode(ctx context.Context, periodeID int64, hasil []HasilHitung) error {
	if r.db == nil {
		return errDatabaseTidakSiap
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Hapus anak lebih dulu karena foreign key memakai RESTRICT.
	if _, err := tx.ExecContext(ctx, `
		DELETE pk FROM potongan_kasbon pk
		JOIN slip_gaji sg ON sg.id = pk.slip_id
		WHERE sg.periode_id = ?`, periodeID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		DELETE d FROM slip_gaji_detail d
		JOIN slip_gaji sg ON sg.id = d.slip_id
		WHERE sg.periode_id = ?`, periodeID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM slip_gaji WHERE periode_id = ?`, periodeID); err != nil {
		return err
	}

	for i := range hasil {
		slip := hasil[i].Slip
		slip.PeriodeID = periodeID
		if err := simpanSatuSlip(ctx, tx, &slip, hasil[i].Rincian, hasil[i].Potongan); err != nil {
			return err
		}
		hasil[i].Slip.ID = slip.ID
	}
	return tx.Commit()
}

func simpanSatuSlip(ctx context.Context, tx *sql.Tx, slip *domain.SlipGaji, rincian []domain.SlipGajiDetail, potongan []domain.PotonganKasbon) error {
	if slip.DibuatPada.IsZero() {
		slip.DibuatPada = time.Now()
	}

	var id int64
	err := tx.QueryRowContext(ctx, `
		SELECT id FROM slip_gaji
		WHERE periode_id = ? AND pekerja_id = ?
		FOR UPDATE`, slip.PeriodeID, slip.PekerjaID).Scan(&id)
	switch {
	case err == nil:
		_, err = tx.ExecContext(ctx, `
			UPDATE slip_gaji SET
				upah_pokok = ?, total_lembur = ?, potongan_telat = ?,
				total_koreksi = ?, gaji_bruto = ?, potongan_kasbon = ?,
				gaji_bersih = ?, hari_kerja_dibayar = ?, hari_kerja_periode = ?,
				dibuat_pada = ?
			WHERE id = ?`,
			int64(slip.UpahPokok), int64(slip.TotalLembur), int64(slip.PotonganTelat),
			int64(slip.TotalKoreksi), int64(slip.GajiBruto), int64(slip.PotonganKasbon),
			int64(slip.GajiBersih), slip.HariKerjaDibayar, nullableInt(slip.HariKerjaPeriode),
			slip.DibuatPada, id)
		if err != nil {
			return err
		}
		slip.ID = id
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM potongan_kasbon WHERE slip_id = ?`, id); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM slip_gaji_detail WHERE slip_id = ?`, id); err != nil {
			return err
		}
	case errors.Is(err, sql.ErrNoRows):
		hasil, err := tx.ExecContext(ctx, `
			INSERT INTO slip_gaji (
				periode_id, pekerja_id, upah_pokok, total_lembur, potongan_telat,
				total_koreksi, gaji_bruto, potongan_kasbon, gaji_bersih,
				hari_kerja_dibayar, hari_kerja_periode, dibuat_pada)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			slip.PeriodeID, slip.PekerjaID, int64(slip.UpahPokok), int64(slip.TotalLembur),
			int64(slip.PotonganTelat), int64(slip.TotalKoreksi), int64(slip.GajiBruto),
			int64(slip.PotonganKasbon), int64(slip.GajiBersih), slip.HariKerjaDibayar,
			nullableInt(slip.HariKerjaPeriode), slip.DibuatPada)
		if err != nil {
			return err
		}
		id, err = hasil.LastInsertId()
		if err != nil {
			return err
		}
		slip.ID = id
	default:
		return err
	}

	for _, d := range rincian {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO slip_gaji_detail (
				slip_id, absensi_id, tanggal, upah_pokok_hari, upah_per_jam,
				menit_telat, potongan_telat, menit_lembur, upah_lembur)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			slip.ID, d.AbsensiID, formatTanggal(d.Tanggal), int64(d.UpahPokokHari),
			int64(d.UpahPerJam), d.MenitTelat, int64(d.PotonganTelat),
			d.MenitLembur, int64(d.UpahLembur)); err != nil {
			return err
		}
	}
	for _, p := range potongan {
		if p.Jumlah <= 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO potongan_kasbon (slip_id, kasbon_id, jumlah)
			VALUES (?, ?, ?)`,
			slip.ID, p.KasbonID, int64(p.Jumlah)); err != nil {
			return err
		}
	}
	return nil
}

// AmbilSlip mengambil slip dan rincian satu pekerja.
//
// Masukan: context, ID periode, ID pekerja.
// Hasil: slip dan rincian harian.
// Aturan: 17.
// Tabel: slip_gaji, slip_gaji_detail, pekerja, periode_gaji.
// Memanggil: database/sql QueryRowContext, QueryContext.
func (r *Repository) AmbilSlip(ctx context.Context, periodeID, pekerjaID int64) (domain.SlipGaji, []domain.SlipGajiDetail, error) {
	slip, rincian, err := r.AmbilRincian(ctx, periodeID, pekerjaID)
	if err != nil {
		return domain.SlipGaji{}, nil, err
	}
	detail := make([]domain.SlipGajiDetail, 0, len(rincian))
	for _, d := range rincian {
		detail = append(detail, d.SlipGajiDetail)
	}
	return slip, detail, nil
}

// AmbilRincian mengambil slip beserta jam absensi dan nama shift.
//
// Masukan: context, ID periode, ID pekerja.
// Hasil: slip dan rincian harian untuk tampilan.
// Aturan: 17.
// Tabel: slip_gaji, slip_gaji_detail, absensi, jadwal_shift, shift, pekerja, periode_gaji.
// Memanggil: database/sql QueryRowContext, QueryContext.
func (r *Repository) AmbilRincian(ctx context.Context, periodeID, pekerjaID int64) (domain.SlipGaji, []RincianHarian, error) {
	if r.db == nil {
		return domain.SlipGaji{}, nil, errDatabaseTidakSiap
	}

	var (
		slip         domain.SlipGaji
		pekerja      domain.Pekerja
		periode      domain.PeriodeGaji
		jenis        string
		tipe         string
		status       string
		tanggalStop  sql.NullTime
		hariPeriode  sql.NullInt64
		periodeTutup sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, `
		SELECT
			sg.id, sg.periode_id, sg.pekerja_id, sg.upah_pokok, sg.total_lembur,
			sg.potongan_telat, sg.total_koreksi, sg.gaji_bruto, sg.potongan_kasbon,
			sg.gaji_bersih, sg.hari_kerja_dibayar, sg.hari_kerja_periode, sg.dibuat_pada,
			p.id, p.kode, p.nama, p.jenis, p.tanggal_mulai, p.tanggal_berhenti, p.aktif,
			pg.id, pg.tipe, pg.tanggal_mulai, pg.tanggal_selesai, pg.status, pg.ditutup_pada
		FROM slip_gaji sg
		JOIN pekerja p ON p.id = sg.pekerja_id
		JOIN periode_gaji pg ON pg.id = sg.periode_id
		WHERE sg.periode_id = ? AND sg.pekerja_id = ?`,
		periodeID, pekerjaID).Scan(
		&slip.ID, &slip.PeriodeID, &slip.PekerjaID, &slip.UpahPokok, &slip.TotalLembur,
		&slip.PotonganTelat, &slip.TotalKoreksi, &slip.GajiBruto, &slip.PotonganKasbon,
		&slip.GajiBersih, &slip.HariKerjaDibayar, &hariPeriode, &slip.DibuatPada,
		&pekerja.ID, &pekerja.Kode, &pekerja.Nama, &jenis, &pekerja.TanggalMulai,
		&tanggalStop, &pekerja.Aktif,
		&periode.ID, &tipe, &periode.TanggalMulai, &periode.TanggalSelesai,
		&status, &periodeTutup,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.SlipGaji{}, nil, domain.ErrTidakDitemukan
	}
	if err != nil {
		return domain.SlipGaji{}, nil, err
	}
	if hariPeriode.Valid {
		slip.HariKerjaPeriode = int(hariPeriode.Int64)
	}
	if tanggalStop.Valid {
		tanggal := tanggalStop.Time
		pekerja.TanggalBerhenti = &tanggal
	}
	pekerja.Jenis = domain.JenisPekerja(jenis)
	periode.Tipe = domain.TipePeriode(tipe)
	periode.Status = domain.StatusPeriode(status)
	if periodeTutup.Valid {
		tutup := periodeTutup.Time
		periode.DitutupPada = &tutup
	}
	slip.Pekerja = &pekerja
	slip.Periode = &periode

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			d.id, d.slip_id, d.absensi_id, d.tanggal, d.upah_pokok_hari,
			d.upah_per_jam, d.menit_telat, d.potongan_telat, d.menit_lembur,
			d.upah_lembur, s.nama, a.jam_masuk, a.jam_keluar
		FROM slip_gaji_detail d
		JOIN absensi a ON a.id = d.absensi_id
		JOIN jadwal_shift j ON j.id = a.jadwal_shift_id
		JOIN shift s ON s.id = j.shift_id
		WHERE d.slip_id = ?
		ORDER BY d.tanggal, d.id`, slip.ID)
	if err != nil {
		return domain.SlipGaji{}, nil, err
	}
	defer rows.Close()

	var rincian []RincianHarian
	for rows.Next() {
		var (
			d         RincianHarian
			jamKeluar sql.NullTime
		)
		if err := rows.Scan(
			&d.ID, &d.SlipID, &d.AbsensiID, &d.Tanggal, &d.UpahPokokHari,
			&d.UpahPerJam, &d.MenitTelat, &d.PotonganTelat, &d.MenitLembur,
			&d.UpahLembur, &d.NamaShift, &d.JamMasuk, &jamKeluar,
		); err != nil {
			return domain.SlipGaji{}, nil, err
		}
		if jamKeluar.Valid {
			keluar := jamKeluar.Time
			d.JamKeluar = &keluar
		}
		rincian = append(rincian, d)
	}
	if err := rows.Err(); err != nil {
		return domain.SlipGaji{}, nil, err
	}
	return slip, rincian, nil
}

// DaftarSlip mengambil seluruh slip pada periode.
//
// Masukan: context dan ID periode.
// Hasil: daftar slip.
// Aturan: 17.
// Tabel: slip_gaji, pekerja.
// Memanggil: database/sql QueryContext.
func (r *Repository) DaftarSlip(ctx context.Context, periodeID int64) ([]domain.SlipGaji, error) {
	if r.db == nil {
		return nil, errDatabaseTidakSiap
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			sg.id, sg.periode_id, sg.pekerja_id, sg.upah_pokok, sg.total_lembur,
			sg.potongan_telat, sg.total_koreksi, sg.gaji_bruto, sg.potongan_kasbon,
			sg.gaji_bersih, sg.hari_kerja_dibayar, sg.hari_kerja_periode, sg.dibuat_pada,
			p.id, p.kode, p.nama, p.jenis, p.tanggal_mulai, p.tanggal_berhenti, p.aktif
		FROM slip_gaji sg
		JOIN pekerja p ON p.id = sg.pekerja_id
		WHERE sg.periode_id = ?
		ORDER BY p.nama, p.kode`, periodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []domain.SlipGaji
	for rows.Next() {
		var (
			slip        domain.SlipGaji
			pekerja     domain.Pekerja
			jenis       string
			tanggalStop sql.NullTime
			hariPeriode sql.NullInt64
		)
		if err := rows.Scan(
			&slip.ID, &slip.PeriodeID, &slip.PekerjaID, &slip.UpahPokok,
			&slip.TotalLembur, &slip.PotonganTelat, &slip.TotalKoreksi,
			&slip.GajiBruto, &slip.PotonganKasbon, &slip.GajiBersih,
			&slip.HariKerjaDibayar, &hariPeriode, &slip.DibuatPada,
			&pekerja.ID, &pekerja.Kode, &pekerja.Nama, &jenis,
			&pekerja.TanggalMulai, &tanggalStop, &pekerja.Aktif,
		); err != nil {
			return nil, err
		}
		if hariPeriode.Valid {
			slip.HariKerjaPeriode = int(hariPeriode.Int64)
		}
		if tanggalStop.Valid {
			tanggal := tanggalStop.Time
			pekerja.TanggalBerhenti = &tanggal
		}
		pekerja.Jenis = domain.JenisPekerja(jenis)
		slip.Pekerja = &pekerja
		hasil = append(hasil, slip)
	}
	return hasil, rows.Err()
}

// DaftarPeriode mengambil semua periode untuk pilihan halaman penggajian.
func (r *Repository) DaftarPeriode(ctx context.Context) ([]domain.PeriodeGaji, error) {
	if r.db == nil {
		return nil, errDatabaseTidakSiap
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tipe, tanggal_mulai, tanggal_selesai, status, ditutup_pada
		FROM periode_gaji
		ORDER BY tanggal_mulai DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []domain.PeriodeGaji
	for rows.Next() {
		periode, err := scanPeriode(rows)
		if err != nil {
			return nil, err
		}
		hasil = append(hasil, periode)
	}
	return hasil, rows.Err()
}

// AmbilPeriode mengambil satu periode.
func (r *Repository) AmbilPeriode(ctx context.Context, id int64) (domain.PeriodeGaji, error) {
	if r.db == nil {
		return domain.PeriodeGaji{}, errDatabaseTidakSiap
	}
	row := r.db.QueryRowContext(ctx, `
		SELECT id, tipe, tanggal_mulai, tanggal_selesai, status, ditutup_pada
		FROM periode_gaji WHERE id = ?`, id)
	periode, err := scanPeriode(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PeriodeGaji{}, domain.ErrTidakDitemukan
	}
	return periode, err
}

// PeriodeSebelumnya mengambil periode bertipe sama yang lebih awal.
func (r *Repository) PeriodeSebelumnya(ctx context.Context, periode domain.PeriodeGaji) (domain.PeriodeGaji, bool, error) {
	if r.db == nil {
		return domain.PeriodeGaji{}, false, errDatabaseTidakSiap
	}
	row := r.db.QueryRowContext(ctx, `
		SELECT id, tipe, tanggal_mulai, tanggal_selesai, status, ditutup_pada
		FROM periode_gaji
		WHERE tipe = ? AND tanggal_mulai < ?
		ORDER BY tanggal_mulai DESC, id DESC
		LIMIT 1`, string(periode.Tipe), formatTanggal(periode.TanggalMulai))
	hasil, err := scanPeriode(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PeriodeGaji{}, false, nil
	}
	if err != nil {
		return domain.PeriodeGaji{}, false, err
	}
	return hasil, true, nil
}

// PekerjaAktif mengambil pekerja aktif pada rentang periode.
func (r *Repository) PekerjaAktif(ctx context.Context, jenis domain.JenisPekerja, mulai, selesai time.Time) ([]domain.Pekerja, error) {
	if r.db == nil {
		return nil, errDatabaseTidakSiap
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, kode, nama, jenis, tanggal_mulai, tanggal_berhenti, aktif
		FROM pekerja
		WHERE tanggal_mulai <= ?
		  AND (tanggal_berhenti IS NULL OR tanggal_berhenti >= ?)
		  AND jenis = ?
		ORDER BY nama, kode`,
		formatTanggal(selesai), formatTanggal(mulai), string(jenis))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []domain.Pekerja
	for rows.Next() {
		pekerja, err := scanPekerja(rows)
		if err != nil {
			return nil, err
		}
		hasil = append(hasil, pekerja)
	}
	return hasil, rows.Err()
}

// AbsensiPeriode mengambil seluruh absensi lengkap pada periode.
func (r *Repository) AbsensiPeriode(ctx context.Context, periodeID int64) ([]domain.AbsensiLengkap, error) {
	if r.db == nil {
		return nil, errDatabaseTidakSiap
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			a.id, a.jadwal_shift_id, a.jam_masuk, a.jam_keluar,
			j.id, j.pekerja_id, j.shift_id, j.tanggal,
			s.id, s.nama, CAST(s.jam_mulai AS CHAR), CAST(s.jam_selesai AS CHAR),
			p.id, p.kode, p.nama, p.jenis, p.tanggal_mulai, p.tanggal_berhenti, p.aktif
		FROM absensi a
		JOIN jadwal_shift j ON j.id = a.jadwal_shift_id
		JOIN shift s ON s.id = j.shift_id
		JOIN pekerja p ON p.id = j.pekerja_id
		JOIN periode_gaji pg ON pg.id = ?
		WHERE j.tanggal BETWEEN pg.tanggal_mulai AND pg.tanggal_selesai
		ORDER BY j.tanggal, p.nama, p.kode`, periodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []domain.AbsensiLengkap
	for rows.Next() {
		var (
			a           domain.AbsensiLengkap
			jamKeluar   sql.NullTime
			jenis       string
			tanggalStop sql.NullTime
		)
		if err := rows.Scan(
			&a.Absensi.ID, &a.Absensi.JadwalShiftID, &a.Absensi.JamMasuk, &jamKeluar,
			&a.Jadwal.ID, &a.Jadwal.PekerjaID, &a.Jadwal.ShiftID, &a.Jadwal.Tanggal,
			&a.Jadwal.Shift.ID, &a.Jadwal.Shift.Nama, &a.Jadwal.Shift.JamMulai,
			&a.Jadwal.Shift.JamSelesai,
			&a.Pekerja.ID, &a.Pekerja.Kode, &a.Pekerja.Nama, &jenis,
			&a.Pekerja.TanggalMulai, &tanggalStop, &a.Pekerja.Aktif,
		); err != nil {
			return nil, err
		}
		if jamKeluar.Valid {
			keluar := jamKeluar.Time
			a.Absensi.JamKeluar = &keluar
		}
		if tanggalStop.Valid {
			tanggal := tanggalStop.Time
			a.Pekerja.TanggalBerhenti = &tanggal
		}
		a.Pekerja.Jenis = domain.JenisPekerja(jenis)
		hasil = append(hasil, a)
	}
	return hasil, rows.Err()
}

// AbsensiBelumLengkap menghitung absensi tanpa jam keluar pada periode.
func (r *Repository) AbsensiBelumLengkap(ctx context.Context, periodeID int64) (int, error) {
	if r.db == nil {
		return 0, errDatabaseTidakSiap
	}
	var jumlah int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM absensi a
		JOIN jadwal_shift j ON j.id = a.jadwal_shift_id
		JOIN periode_gaji pg ON pg.id = ?
		WHERE j.tanggal BETWEEN pg.tanggal_mulai AND pg.tanggal_selesai
		  AND a.jam_keluar IS NULL`, periodeID).Scan(&jumlah)
	return jumlah, err
}

// UpahSampai mengambil riwayat upah sampai tanggal tertentu, terurut naik.
func (r *Repository) UpahSampai(ctx context.Context, pekerjaID int64, sampai time.Time) ([]domain.Upah, error) {
	if r.db == nil {
		return nil, errDatabaseTidakSiap
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, pekerja_id, nominal, berlaku_mulai, alasan
		FROM riwayat_upah
		WHERE pekerja_id = ? AND berlaku_mulai <= ?
		ORDER BY berlaku_mulai ASC, id ASC`,
		pekerjaID, formatTanggal(sampai))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []domain.Upah
	for rows.Next() {
		var (
			upah   domain.Upah
			alasan sql.NullString
		)
		if err := rows.Scan(&upah.ID, &upah.PekerjaID, &upah.Nominal,
			&upah.BerlakuMulai, &alasan); err != nil {
			return nil, err
		}
		upah.Alasan = alasan.String
		hasil = append(hasil, upah)
	}
	return hasil, rows.Err()
}

// KasbonSampai mengambil kasbon bersisa sampai akhir periode, terurut tertua.
func (r *Repository) KasbonSampai(ctx context.Context, pekerjaID int64, sampai time.Time) ([]domain.Kasbon, error) {
	if r.db == nil {
		return nil, errDatabaseTidakSiap
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, pekerja_id, tanggal, jumlah, sisa, keterangan
		FROM kasbon
		WHERE pekerja_id = ? AND sisa > 0 AND tanggal <= ?
		ORDER BY tanggal ASC, id ASC`,
		pekerjaID, formatTanggal(sampai))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []domain.Kasbon
	for rows.Next() {
		var (
			k      domain.Kasbon
			ket    sql.NullString
			jumlah int64
			sisa   int64
		)
		if err := rows.Scan(&k.ID, &k.PekerjaID, &k.Tanggal, &jumlah, &sisa, &ket); err != nil {
			return nil, err
		}
		k.Jumlah = domain.Rupiah(jumlah)
		k.Sisa = domain.Rupiah(sisa)
		k.Keterangan = ket.String
		hasil = append(hasil, k)
	}
	return hasil, rows.Err()
}

// SisaKasbon menjumlahkan sisa kasbon pekerja.
func (r *Repository) SisaKasbon(ctx context.Context, pekerjaID int64) (domain.Rupiah, error) {
	if r.db == nil {
		return 0, errDatabaseTidakSiap
	}
	var total int64
	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(sisa), 0) FROM kasbon WHERE pekerja_id = ?`,
		pekerjaID).Scan(&total)
	return domain.Rupiah(total), err
}

// KoreksiUntukPeriode mengambil koreksi yang belum diterapkan atau sudah
// diterapkan tepat pada periode ini. Dengan begitu hitung ulang tidak
// menghitung koreksi dua kali.
func (r *Repository) KoreksiUntukPeriode(ctx context.Context, periodeID, pekerjaID int64) ([]domain.KoreksiGaji, error) {
	if r.db == nil {
		return nil, errDatabaseTidakSiap
	}
	return r.daftarKoreksi(ctx, `
		SELECT id, pekerja_id, periode_asal_id, periode_terapan_id, jumlah, keterangan, dibuat_pada
		FROM koreksi_gaji
		WHERE pekerja_id = ? AND (periode_terapan_id IS NULL OR periode_terapan_id = ?)
		ORDER BY dibuat_pada, id`, pekerjaID, periodeID)
}

// KoreksiDiterapkan mengambil koreksi yang sudah diterapkan pada periode ini.
func (r *Repository) KoreksiDiterapkan(ctx context.Context, periodeID, pekerjaID int64) ([]domain.KoreksiGaji, error) {
	if r.db == nil {
		return nil, errDatabaseTidakSiap
	}
	return r.daftarKoreksi(ctx, `
		SELECT id, pekerja_id, periode_asal_id, periode_terapan_id, jumlah, keterangan, dibuat_pada
		FROM koreksi_gaji
		WHERE pekerja_id = ? AND periode_terapan_id = ?
		ORDER BY dibuat_pada, id`, pekerjaID, periodeID)
}

func (r *Repository) daftarKoreksi(ctx context.Context, query string, args ...any) ([]domain.KoreksiGaji, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []domain.KoreksiGaji
	for rows.Next() {
		var (
			k     domain.KoreksiGaji
			terap sql.NullInt64
		)
		if err := rows.Scan(&k.ID, &k.PekerjaID, &k.PeriodeAsalID, &terap,
			&k.Jumlah, &k.Keterangan, &k.DibuatPada); err != nil {
			return nil, err
		}
		if terap.Valid {
			id := terap.Int64
			k.PeriodeTerapanID = &id
		}
		hasil = append(hasil, k)
	}
	return hasil, rows.Err()
}

// TandaiKoreksiPeriode menandai koreksi milik pekerja yang punya slip pada
// periode ini sebagai sudah diterapkan.
func (r *Repository) TandaiKoreksiPeriode(ctx context.Context, periodeID int64) error {
	if r.db == nil {
		return errDatabaseTidakSiap
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE koreksi_gaji k
		JOIN slip_gaji s ON s.pekerja_id = k.pekerja_id AND s.periode_id = ?
		SET k.periode_terapan_id = ?
		WHERE k.periode_terapan_id IS NULL`, periodeID, periodeID)
	return err
}

// TanggalTerkunci memeriksa langsung apakah tanggal ada di periode DITUTUP.
func (r *Repository) TanggalTerkunci(ctx context.Context, tanggal time.Time) (bool, error) {
	if r.db == nil {
		return false, errDatabaseTidakSiap
	}
	var terkunci bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM periode_gaji
			WHERE status = 'DITUTUP' AND ? BETWEEN tanggal_mulai AND tanggal_selesai
		)`, formatTanggal(tanggal)).Scan(&terkunci)
	return terkunci, err
}

// UbahStatusPeriode memperbarui status periode.
func (r *Repository) UbahStatusPeriode(ctx context.Context, periodeID int64, status domain.StatusPeriode) error {
	if r.db == nil {
		return errDatabaseTidakSiap
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE periode_gaji SET status = ? WHERE id = ?`, string(status), periodeID)
	return err
}

type pemindai interface {
	Scan(dest ...any) error
}

func scanPeriode(s pemindai) (domain.PeriodeGaji, error) {
	var (
		periode domain.PeriodeGaji
		tipe    string
		status  string
		tutup   sql.NullTime
	)
	if err := s.Scan(&periode.ID, &tipe, &periode.TanggalMulai,
		&periode.TanggalSelesai, &status, &tutup); err != nil {
		return domain.PeriodeGaji{}, err
	}
	periode.Tipe = domain.TipePeriode(tipe)
	periode.Status = domain.StatusPeriode(status)
	if tutup.Valid {
		tanggal := tutup.Time
		periode.DitutupPada = &tanggal
	}
	return periode, nil
}

func scanPekerja(s pemindai) (domain.Pekerja, error) {
	var (
		pekerja domain.Pekerja
		jenis   string
		stop    sql.NullTime
	)
	if err := s.Scan(&pekerja.ID, &pekerja.Kode, &pekerja.Nama, &jenis,
		&pekerja.TanggalMulai, &stop, &pekerja.Aktif); err != nil {
		return domain.Pekerja{}, err
	}
	pekerja.Jenis = domain.JenisPekerja(jenis)
	if stop.Valid {
		tanggal := stop.Time
		pekerja.TanggalBerhenti = &tanggal
	}
	return pekerja, nil
}

func formatTanggal(t time.Time) string {
	return t.Format(layoutTanggalSQL)
}

func nullableInt(nilai int) any {
	if nilai <= 0 {
		return nil
	}
	return nilai
}
