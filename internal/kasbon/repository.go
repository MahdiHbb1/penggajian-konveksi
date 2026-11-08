// Pemilik: Edgar | Modul 2 | Aturan: 12, 14, 16
package kasbon

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"penggajian-konveksi/internal/domain"
)

// Filter adalah parameter daftar kasbon.
// Status berisi "BERJALAN", "LUNAS", atau kosong (semua).
type Filter struct {
	PekerjaID *int64
	Status    string
}

// KasbonTampil adalah kasbon beserta nama pekerja untuk tampilan.
type KasbonTampil struct {
	domain.Kasbon
	NamaPekerja string
}

// PotonganTampil adalah satu potongan kasbon beserta periodenya.
type PotonganTampil struct {
	SlipID        int64
	Jumlah        domain.Rupiah
	Tipe          domain.TipePeriode
	Mulai         time.Time
	Selesai       time.Time
	StatusPeriode domain.StatusPeriode
}

// Repository mengakses tabel kasbon dan potongan_kasbon.
type Repository struct {
	db *sql.DB
}

// NewRepository membuat repository kasbon.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const formatTanggalSQL = "2006-01-02"

// SimpanKasbon menyimpan kasbon baru.
//
// Masukan: context dan pointer Kasbon.
// Hasil: error; ID hasil insert diisi ke pointer.
// Aturan: 12.
// Tabel: kasbon.
// Memanggil: database/sql ExecContext.
func (r *Repository) SimpanKasbon(ctx context.Context, kasbon *domain.Kasbon) error {
	var ket sql.NullString
	if kasbon.Keterangan != "" {
		ket = sql.NullString{String: kasbon.Keterangan, Valid: true}
	}
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO kasbon (pekerja_id, tanggal, jumlah, sisa, keterangan)
		 VALUES (?, ?, ?, ?, ?)`,
		kasbon.PekerjaID,
		kasbon.Tanggal.Format(formatTanggalSQL),
		int64(kasbon.Jumlah),
		int64(kasbon.Sisa),
		ket,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	kasbon.ID = id
	return nil
}

type pemindai interface {
	Scan(dest ...any) error
}

func pindaiKasbonTampil(sc pemindai) (KasbonTampil, error) {
	var k KasbonTampil
	var jumlah, sisa int64
	var ket sql.NullString
	if err := sc.Scan(&k.ID, &k.PekerjaID, &k.NamaPekerja, &k.Tanggal, &jumlah, &sisa, &ket); err != nil {
		return KasbonTampil{}, err
	}
	k.Jumlah = domain.Rupiah(jumlah)
	k.Sisa = domain.Rupiah(sisa)
	k.Keterangan = ket.String
	return k, nil
}

const selectKasbonTampil = `
	SELECT k.id, k.pekerja_id, p.nama, k.tanggal, k.jumlah, k.sisa, k.keterangan
	FROM kasbon k
	JOIN pekerja p ON p.id = k.pekerja_id`

// DaftarKasbon mengambil kasbon sesuai filter.
//
// Masukan: context dan Filter.
// Hasil: daftar kasbon, terbaru lebih dulu.
// Aturan: 12, 16.
// Tabel: kasbon, pekerja.
// Memanggil: database/sql QueryContext.
func (r *Repository) DaftarKasbon(ctx context.Context, filter Filter) ([]KasbonTampil, error) {
	var q strings.Builder
	q.WriteString(selectKasbonTampil)
	q.WriteString(" WHERE 1 = 1")
	var args []any
	if filter.PekerjaID != nil {
		q.WriteString(" AND k.pekerja_id = ?")
		args = append(args, *filter.PekerjaID)
	}
	switch filter.Status {
	case "BERJALAN":
		q.WriteString(" AND k.sisa > 0")
	case "LUNAS":
		q.WriteString(" AND k.sisa = 0")
	}
	q.WriteString(" ORDER BY k.tanggal DESC, k.id DESC")

	rows, err := r.db.QueryContext(ctx, q.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []KasbonTampil
	for rows.Next() {
		k, err := pindaiKasbonTampil(rows)
		if err != nil {
			return nil, err
		}
		hasil = append(hasil, k)
	}
	return hasil, rows.Err()
}

// AmbilKasbon mengambil satu kasbon.
//
// Masukan: context dan ID kasbon.
// Hasil: Kasbon; ErrKasbonTidakAda bila tidak ditemukan.
// Aturan: 12.
// Tabel: kasbon, pekerja.
// Memanggil: database/sql QueryRowContext.
func (r *Repository) AmbilKasbon(ctx context.Context, id int64) (KasbonTampil, error) {
	row := r.db.QueryRowContext(ctx, selectKasbonTampil+" WHERE k.id = ?", id)
	k, err := pindaiKasbonTampil(row)
	if errors.Is(err, sql.ErrNoRows) {
		return KasbonTampil{}, ErrKasbonTidakAda
	}
	return k, err
}

// SisaKasbon menjumlahkan sisa kasbon pekerja.
//
// Masukan: context dan ID pekerja.
// Hasil: total sisa.
// Aturan: 12, 16.
// Tabel: kasbon.
// Memanggil: database/sql QueryRowContext.
func (r *Repository) SisaKasbon(ctx context.Context, pekerjaID int64) (domain.Rupiah, error) {
	var total int64
	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(sisa), 0) FROM kasbon WHERE pekerja_id = ?`, pekerjaID,
	).Scan(&total)
	return domain.Rupiah(total), err
}

// KasbonDiprosesSampai mengambil kasbon yang belum lunas sampai tanggal.
//
// Masukan: context, ID pekerja, batas tanggal (ikut dihitung).
// Hasil: kasbon terurut dari yang tertua.
// Aturan: 12.
// Tabel: kasbon.
// Memanggil: database/sql QueryContext.
func (r *Repository) KasbonDiprosesSampai(ctx context.Context, pekerjaID int64, tanggal time.Time) ([]domain.Kasbon, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, pekerja_id, tanggal, jumlah, sisa, keterangan
		 FROM kasbon
		 WHERE pekerja_id = ? AND sisa > 0 AND tanggal <= ?
		 ORDER BY tanggal ASC, id ASC`,
		pekerjaID, tanggal.Format(formatTanggalSQL),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []domain.Kasbon
	for rows.Next() {
		var k domain.Kasbon
		var jumlah, sisa int64
		var ket sql.NullString
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

// RiwayatPotongan mengambil potongan kasbon per periode.
//
// Masukan: context dan ID kasbon.
// Hasil: daftar potongan, periode paling awal lebih dulu.
// Aturan: 12, 13.
// Tabel: potongan_kasbon, slip_gaji, periode_gaji.
// Memanggil: database/sql QueryContext.
func (r *Repository) RiwayatPotongan(ctx context.Context, kasbonID int64) ([]PotonganTampil, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT pk.slip_id, pk.jumlah, pg.tipe, pg.tanggal_mulai, pg.tanggal_selesai, pg.status
		 FROM potongan_kasbon pk
		 JOIN slip_gaji sg ON sg.id = pk.slip_id
		 JOIN periode_gaji pg ON pg.id = sg.periode_id
		 WHERE pk.kasbon_id = ?
		 ORDER BY pg.tanggal_mulai ASC`,
		kasbonID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasil []PotonganTampil
	for rows.Next() {
		var p PotonganTampil
		var jumlah int64
		var tipe, status string
		if err := rows.Scan(&p.SlipID, &jumlah, &tipe, &p.Mulai, &p.Selesai, &status); err != nil {
			return nil, err
		}
		p.Jumlah = domain.Rupiah(jumlah)
		p.Tipe = domain.TipePeriode(tipe)
		p.StatusPeriode = domain.StatusPeriode(status)
		hasil = append(hasil, p)
	}
	return hasil, rows.Err()
}

// PekerjaAktif memeriksa apakah pekerja ada dan berstatus aktif.
//
// Masukan: context dan ID pekerja.
// Hasil: true bila aktif.
// Aturan: 16.
// Tabel: pekerja.
// Memanggil: database/sql QueryRowContext.
func (r *Repository) PekerjaAktif(ctx context.Context, pekerjaID int64) (bool, error) {
	var ada bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM pekerja WHERE id = ? AND aktif = TRUE)`, pekerjaID,
	).Scan(&ada)
	return ada, err
}

// DaftarPekerja mengambil pekerja untuk pilihan dropdown.
//
// Masukan: context dan penanda hanya pekerja aktif.
// Hasil: pekerja terurut menurut nama.
// Aturan: 16 (pekerja nonaktif tidak muncul di pilihan kasbon baru).
// Tabel: pekerja.
// Memanggil: database/sql QueryContext.
func (r *Repository) DaftarPekerja(ctx context.Context, hanyaAktif bool) ([]domain.Pekerja, error) {
	q := `SELECT id, kode, nama, jenis, aktif FROM pekerja`
	if hanyaAktif {
		q += ` WHERE aktif = TRUE`
	}
	q += ` ORDER BY nama ASC`

	rows, err := r.db.QueryContext(ctx, q)
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

// TanggalTerkunci memeriksa langsung apakah tanggal ada di periode DITUTUP.
// Dipakai sebagai cadangan selama PeriodeProvider Modul 5 belum jadi.
//
// Masukan: context dan tanggal.
// Hasil: true bila terkunci.
// Aturan: 14.
// Tabel: periode_gaji.
// Memanggil: database/sql QueryRowContext.
func (r *Repository) TanggalTerkunci(ctx context.Context, tanggal time.Time) (bool, error) {
	var terkunci bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(
		   SELECT 1 FROM periode_gaji
		   WHERE status = 'DITUTUP' AND ? BETWEEN tanggal_mulai AND tanggal_selesai
		 )`,
		tanggal.Format(formatTanggalSQL),
	).Scan(&terkunci)
	return terkunci, err
}