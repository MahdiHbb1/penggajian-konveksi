// Pemilik: Mahdi | Modul 1 | Aturan: 8, 9, 16
package pekerja

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	mysql "github.com/go-sql-driver/mysql"

	"penggajian-konveksi/internal/domain"
)

// Filter adalah parameter pencarian daftar pekerja.
type Filter struct {
	Cari  string
	Jenis domain.JenisPekerja
	Aktif *bool
}

// Repository mengakses tabel pekerja dan riwayat_upah.
type Repository struct {
	db *sql.DB
}

// NewRepository membuat repository pekerja.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

var errDatabaseTidakSiap = errors.New("Database belum tersedia.")

type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type barisScanner interface {
	Scan(dest ...any) error
}

const (
	layoutTanggal = "2006-01-02"
	kolomPekerja  = "id, kode, nama, jenis, tanggal_mulai, tanggal_berhenti, aktif"
	kolomUpah     = "id, pekerja_id, nominal, berlaku_mulai, alasan"
)

// SimpanPekerja menyimpan satu baris pekerja baru.
//
// Masukan: context dan pointer Pekerja.
// Hasil: error; ID hasil insert diisi ke pointer.
// Aturan: 1, 11, 16.
// Tabel: pekerja.
// Memanggil: database/sql ExecContext.
func (r *Repository) SimpanPekerja(ctx context.Context, pekerja *domain.Pekerja) error {
	if r.db == nil {
		return errDatabaseTidakSiap
	}
	return r.simpanPekerja(ctx, r.db, pekerja)
}

// UbahPekerja memperbarui data pekerja yang sudah ada.
//
// Masukan: context dan Pekerja.
// Hasil: error.
// Aturan: 11, 16.
// Tabel: pekerja.
// Memanggil: database/sql ExecContext.
func (r *Repository) UbahPekerja(ctx context.Context, pekerja domain.Pekerja) error {
	if r.db == nil {
		return errDatabaseTidakSiap
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE pekerja
		SET kode = ?, nama = ?, tanggal_mulai = ?, tanggal_berhenti = ?, aktif = ?
		WHERE id = ?`,
		pekerja.Kode,
		pekerja.Nama,
		formatTanggal(pekerja.TanggalMulai),
		nullableTanggal(pekerja.TanggalBerhenti),
		pekerja.Aktif,
		pekerja.ID,
	)
	if err != nil {
		return terjemahkanDuplikat(err, "Kode pekerja sudah dipakai.")
	}
	return nil
}

// NonaktifkanPekerja mengisi tanggal berhenti dan aktif=false.
//
// Masukan: context, ID pekerja, tanggal berhenti.
// Hasil: error.
// Aturan: 16.
// Tabel: pekerja.
// Memanggil: database/sql ExecContext.
func (r *Repository) NonaktifkanPekerja(ctx context.Context, id int64, tanggal time.Time) error {
	if r.db == nil {
		return errDatabaseTidakSiap
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE pekerja
		SET tanggal_berhenti = ?, aktif = FALSE
		WHERE id = ?`,
		formatTanggal(tanggal), id,
	)
	return err
}

// DaftarPekerja mengambil pekerja sesuai filter.
//
// Masukan: context dan Filter.
// Hasil: daftar pekerja.
// Aturan: 11, 16.
// Tabel: pekerja.
// Memanggil: database/sql QueryContext.
func (r *Repository) DaftarPekerja(ctx context.Context, filter Filter) ([]domain.Pekerja, error) {
	if r.db == nil {
		return nil, errDatabaseTidakSiap
	}

	var kueri strings.Builder
	kueri.WriteString("SELECT " + kolomPekerja + " FROM pekerja WHERE 1 = 1")
	args := make([]any, 0, 3)

	if cari := strings.TrimSpace(filter.Cari); cari != "" {
		pola := "%" + cari + "%"
		kueri.WriteString(" AND (kode LIKE ? OR nama LIKE ?)")
		args = append(args, pola, pola)
	}
	if filter.Jenis != "" {
		kueri.WriteString(" AND jenis = ?")
		args = append(args, string(filter.Jenis))
	}
	if filter.Aktif != nil {
		kueri.WriteString(" AND aktif = ?")
		args = append(args, *filter.Aktif)
	}
	kueri.WriteString(" ORDER BY kode")

	rows, err := r.db.QueryContext(ctx, kueri.String(), args...)
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return hasil, nil
}

// CariPekerja mengambil satu pekerja berdasarkan ID.
//
// Masukan: context dan ID pekerja.
// Hasil: Pekerja.
// Aturan: 16.
// Tabel: pekerja.
// Memanggil: database/sql QueryRowContext.
func (r *Repository) CariPekerja(ctx context.Context, id int64) (domain.Pekerja, error) {
	if r.db == nil {
		return domain.Pekerja{}, errDatabaseTidakSiap
	}
	row := r.db.QueryRowContext(ctx, "SELECT "+kolomPekerja+" FROM pekerja WHERE id = ?", id)
	pekerja, err := scanPekerja(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Pekerja{}, domain.ErrTidakDitemukan
	}
	return pekerja, err
}

// SimpanUpah menyimpan satu baris riwayat upah.
//
// Masukan: context dan pointer Upah.
// Hasil: error; ID hasil insert diisi ke pointer.
// Aturan: 7, 8.
// Tabel: riwayat_upah.
// Memanggil: database/sql ExecContext.
func (r *Repository) SimpanUpah(ctx context.Context, upah *domain.Upah) error {
	if r.db == nil {
		return errDatabaseTidakSiap
	}
	return r.simpanUpah(ctx, r.db, upah)
}

// DaftarUpah mengambil riwayat upah satu pekerja.
//
// Masukan: context dan ID pekerja.
// Hasil: daftar upah terurut tanggal berlaku terbaru lebih dulu.
// Aturan: 8.
// Tabel: riwayat_upah.
// Memanggil: database/sql QueryContext.
func (r *Repository) DaftarUpah(ctx context.Context, pekerjaID int64) ([]domain.Upah, error) {
	if r.db == nil {
		return nil, errDatabaseTidakSiap
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+kolomUpah+`
		FROM riwayat_upah
		WHERE pekerja_id = ?
		ORDER BY berlaku_mulai DESC, id DESC`, pekerjaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []domain.Upah
	for rows.Next() {
		upah, err := scanUpah(rows)
		if err != nil {
			return nil, err
		}
		hasil = append(hasil, upah)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return hasil, nil
}

// UpahBerlaku mengambil upah terbaru yang tidak melewati tanggal.
//
// Masukan: context, ID pekerja, tanggal kerja.
// Hasil: Upah yang berlaku.
// Aturan: 8.
// Tabel: riwayat_upah.
// Memanggil: database/sql QueryRowContext.
func (r *Repository) UpahBerlaku(ctx context.Context, pekerjaID int64, tanggal time.Time) (domain.Upah, error) {
	if r.db == nil {
		return domain.Upah{}, errDatabaseTidakSiap
	}
	row := r.db.QueryRowContext(ctx, `
		SELECT `+kolomUpah+`
		FROM riwayat_upah
		WHERE pekerja_id = ? AND berlaku_mulai <= ?
		ORDER BY berlaku_mulai DESC, id DESC
		LIMIT 1`,
		pekerjaID, formatTanggal(tanggal),
	)
	upah, err := scanUpah(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Upah{}, domain.ErrTidakDitemukan
	}
	return upah, err
}

// PekerjaAktif mengambil pekerja yang aktif pada rentang periode.
//
// Masukan: context, jenis pekerja, awal dan akhir periode.
// Hasil: daftar pekerja.
// Aturan: 9, 16, 17.
// Tabel: pekerja, riwayat_upah.
// Memanggil: database/sql QueryContext.
func (r *Repository) PekerjaAktif(ctx context.Context, jenis domain.JenisPekerja, mulai, selesai time.Time) ([]domain.Pekerja, error) {
	if r.db == nil {
		return nil, errDatabaseTidakSiap
	}

	var kueri strings.Builder
	kueri.WriteString("SELECT " + kolomPekerja + " FROM pekerja")
	kueri.WriteString(" WHERE tanggal_mulai <= ?")
	kueri.WriteString(" AND (tanggal_berhenti IS NULL OR tanggal_berhenti >= ?)")
	args := []any{formatTanggal(selesai), formatTanggal(mulai)}
	if jenis != "" {
		kueri.WriteString(" AND jenis = ?")
		args = append(args, string(jenis))
	}
	kueri.WriteString(" ORDER BY nama, kode")

	rows, err := r.db.QueryContext(ctx, kueri.String(), args...)
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return hasil, nil
}

// SimpanPekerjaDanUpah menyimpan pekerja dan upah pertamanya dalam satu transaksi.
//
// Masukan: context, pointer Pekerja, pointer Upah.
// Hasil: error; ID pekerja dan upah diisi ke pointer.
// Aturan: 8, 11, 16.
// Tabel: pekerja, riwayat_upah.
// Memanggil: database/sql BeginTx, SimpanPekerja, SimpanUpah, Commit.
func (r *Repository) SimpanPekerjaDanUpah(ctx context.Context, pekerja *domain.Pekerja, upah *domain.Upah) error {
	if r.db == nil {
		return errDatabaseTidakSiap
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := r.simpanPekerja(ctx, tx, pekerja); err != nil {
		return err
	}
	upah.PekerjaID = pekerja.ID
	if err := r.simpanUpah(ctx, tx, upah); err != nil {
		return err
	}
	return tx.Commit()
}

// KodeSudahDipakai memeriksa apakah kode pekerja sudah dipakai pekerja lain.
//
// Masukan: context, kode, dan ID pekerja yang dikecualikan (0 saat tambah).
// Hasil: true bila kode sudah dipakai.
// Aturan: 11.
// Tabel: pekerja.
// Memanggil: database/sql QueryRowContext.
func (r *Repository) KodeSudahDipakai(ctx context.Context, kode string, kecualiID int64) (bool, error) {
	if r.db == nil {
		return false, errDatabaseTidakSiap
	}
	var jumlah int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM pekerja WHERE kode = ? AND id <> ?",
		kode, kecualiID,
	).Scan(&jumlah)
	if err != nil {
		return false, err
	}
	return jumlah > 0, nil
}

// UpahPadaTanggal memeriksa apakah sudah ada upah pada tanggal berlaku tertentu.
//
// Masukan: context, ID pekerja, tanggal berlaku.
// Hasil: true bila sudah ada.
// Aturan: 8.
// Tabel: riwayat_upah.
// Memanggil: database/sql QueryRowContext.
func (r *Repository) UpahPadaTanggal(ctx context.Context, pekerjaID int64, tanggal time.Time) (bool, error) {
	if r.db == nil {
		return false, errDatabaseTidakSiap
	}
	var jumlah int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM riwayat_upah WHERE pekerja_id = ? AND berlaku_mulai = ?",
		pekerjaID, formatTanggal(tanggal),
	).Scan(&jumlah)
	if err != nil {
		return false, err
	}
	return jumlah > 0, nil
}

func (r *Repository) simpanPekerja(ctx context.Context, ex execer, pekerja *domain.Pekerja) error {
	hasil, err := ex.ExecContext(ctx, `
		INSERT INTO pekerja (kode, nama, jenis, tanggal_mulai, tanggal_berhenti, aktif)
		VALUES (?, ?, ?, ?, ?, ?)`,
		pekerja.Kode,
		pekerja.Nama,
		string(pekerja.Jenis),
		formatTanggal(pekerja.TanggalMulai),
		nullableTanggal(pekerja.TanggalBerhenti),
		pekerja.Aktif,
	)
	if err != nil {
		return terjemahkanDuplikat(err, "Kode pekerja sudah dipakai.")
	}
	id, err := hasil.LastInsertId()
	if err != nil {
		return err
	}
	pekerja.ID = id
	return nil
}

func (r *Repository) simpanUpah(ctx context.Context, ex execer, upah *domain.Upah) error {
	hasil, err := ex.ExecContext(ctx, `
		INSERT INTO riwayat_upah (pekerja_id, nominal, berlaku_mulai, alasan)
		VALUES (?, ?, ?, ?)`,
		upah.PekerjaID,
		int64(upah.Nominal),
		formatTanggal(upah.BerlakuMulai),
		nullableTeks(upah.Alasan),
	)
	if err != nil {
		return terjemahkanDuplikat(err, "Sudah ada upah dengan tanggal berlaku yang sama.")
	}
	id, err := hasil.LastInsertId()
	if err != nil {
		return err
	}
	upah.ID = id
	return nil
}

func scanPekerja(s barisScanner) (domain.Pekerja, error) {
	var (
		pekerja  domain.Pekerja
		jenis    string
		berhenti sql.NullTime
	)
	err := s.Scan(
		&pekerja.ID,
		&pekerja.Kode,
		&pekerja.Nama,
		&jenis,
		&pekerja.TanggalMulai,
		&berhenti,
		&pekerja.Aktif,
	)
	if err != nil {
		return domain.Pekerja{}, err
	}
	pekerja.Jenis = domain.JenisPekerja(jenis)
	if berhenti.Valid {
		tanggal := berhenti.Time
		pekerja.TanggalBerhenti = &tanggal
	}
	return pekerja, nil
}

func scanUpah(s barisScanner) (domain.Upah, error) {
	var (
		upah    domain.Upah
		nominal int64
		alasan  sql.NullString
	)
	err := s.Scan(
		&upah.ID,
		&upah.PekerjaID,
		&nominal,
		&upah.BerlakuMulai,
		&alasan,
	)
	if err != nil {
		return domain.Upah{}, err
	}
	upah.Nominal = domain.Rupiah(nominal)
	if alasan.Valid {
		upah.Alasan = alasan.String
	}
	return upah, nil
}

func formatTanggal(t time.Time) string {
	return t.Format(layoutTanggal)
}

func nullableTanggal(t *time.Time) any {
	if t == nil {
		return nil
	}
	return formatTanggal(*t)
}

func nullableTeks(teks string) any {
	if strings.TrimSpace(teks) == "" {
		return nil
	}
	return teks
}

func terjemahkanDuplikat(err error, pesan string) error {
	var galatMySQL *mysql.MySQLError
	if errors.As(err, &galatMySQL) && galatMySQL.Number == 1062 {
		return errors.New(pesan)
	}
	return err
}
