// Pemilik: Anya | Modul 5 | Aturan: 1, 12, 13, 14, 15, 18
package periode

import (
	"context"
	"database/sql"
	"errors"
	"penggajian-konveksi/internal/domain"
	"time"
)

type RingkasanTutup struct {
	JumlahSlip          int
	TotalBersih         domain.Rupiah
	TotalPotonganKasbon domain.Rupiah
	JumlahKoreksi       int
}
type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) SimpanPeriode(ctx context.Context, p *domain.PeriodeGaji) error {
	res, e := r.db.ExecContext(ctx, `INSERT INTO periode_gaji(tipe,tanggal_mulai,tanggal_selesai,status) VALUES(?,?,?,'TERBUKA')`, p.Tipe, p.TanggalMulai.Format("2006-01-02"), p.TanggalSelesai.Format("2006-01-02"))
	if e != nil {
		return e
	}
	p.ID, e = res.LastInsertId()
	return e
}
func (r *Repository) AdaTumpangTindih(ctx context.Context, p domain.PeriodeGaji) (bool, error) {
	var n int
	e := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM periode_gaji WHERE tanggal_mulai<=? AND tanggal_selesai>=?`, p.TanggalSelesai.Format("2006-01-02"), p.TanggalMulai.Format("2006-01-02")).Scan(&n)
	return n > 0, e
}
func (r *Repository) DaftarPeriode(ctx context.Context) ([]domain.PeriodeGaji, error) {
	rows, e := r.db.QueryContext(ctx, `SELECT id,tipe,tanggal_mulai,tanggal_selesai,status,ditutup_pada FROM periode_gaji ORDER BY tanggal_mulai DESC,id DESC`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.PeriodeGaji{}
	for rows.Next() {
		p, e := scanPeriode(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
func scanPeriode(s interface{ Scan(...any) error }) (domain.PeriodeGaji, error) {
	var p domain.PeriodeGaji
	var tutup sql.NullTime
	e := s.Scan(&p.ID, &p.Tipe, &p.TanggalMulai, &p.TanggalSelesai, &p.Status, &tutup)
	if tutup.Valid {
		p.DitutupPada = &tutup.Time
	}
	return p, e
}
func (r *Repository) AmbilPeriode(ctx context.Context, id int64) (domain.PeriodeGaji, error) {
	return scanPeriode(r.db.QueryRowContext(ctx, `SELECT id,tipe,tanggal_mulai,tanggal_selesai,status,ditutup_pada FROM periode_gaji WHERE id=?`, id))
}
func (r *Repository) PeriodeTerkunci(ctx context.Context, t time.Time) (bool, error) {
	var n int
	e := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM periode_gaji WHERE status='DITUTUP' AND ? BETWEEN tanggal_mulai AND tanggal_selesai`, t.Format("2006-01-02")).Scan(&n)
	return n > 0, e
}
func (r *Repository) PeriodeSebelumnya(ctx context.Context, p domain.PeriodeGaji) (domain.PeriodeGaji, bool, error) {
	prev, e := scanPeriode(r.db.QueryRowContext(ctx, `SELECT id,tipe,tanggal_mulai,tanggal_selesai,status,ditutup_pada FROM periode_gaji WHERE tipe=? AND tanggal_selesai<? ORDER BY tanggal_selesai DESC LIMIT 1`, p.Tipe, p.TanggalMulai.Format("2006-01-02")))
	if errors.Is(e, sql.ErrNoRows) {
		return domain.PeriodeGaji{}, false, nil
	}
	if e != nil {
		return domain.PeriodeGaji{}, false, e
	}
	return prev, true, nil
}
func (r *Repository) UbahStatus(ctx context.Context, id int64, s domain.StatusPeriode) error {
	_, e := r.db.ExecContext(ctx, `UPDATE periode_gaji SET status=? WHERE id=? AND status<>'DITUTUP'`, s, id)
	return e
}
func (r *Repository) Tutup(ctx context.Context, id int64) error {
	tx, e := r.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var status domain.StatusPeriode
	if e = tx.QueryRowContext(ctx, `SELECT status FROM periode_gaji WHERE id=? FOR UPDATE`, id).Scan(&status); e != nil {
		return e
	}
	if status == domain.PeriodeDitutup {
		return errors.New("periode sudah ditutup dan tidak dapat dibatalkan")
	}
	if status != domain.PeriodeDihitung {
		return errors.New("periode hanya dapat ditutup dari status DIHITUNG")
	}
	// Pengurangan kasbon dilakukan dari rincian potongan aktual setiap slip, tepat sekali.
	_, e = tx.ExecContext(ctx, `UPDATE kasbon k JOIN (SELECT pk.kasbon_id,SUM(pk.jumlah) jumlah FROM potongan_kasbon pk JOIN slip_gaji s ON s.id=pk.slip_id WHERE s.periode_id=? GROUP BY pk.kasbon_id) x ON x.kasbon_id=k.id SET k.sisa=GREATEST(0,k.sisa-x.jumlah)`, id)
	if e != nil {
		return e
	}
	_, e = tx.ExecContext(ctx, `UPDATE koreksi_gaji k JOIN slip_gaji s ON s.pekerja_id=k.pekerja_id SET k.periode_terapan_id=? WHERE s.periode_id=? AND k.periode_terapan_id IS NULL`, id, id)
	if e != nil {
		return e
	}
	res, e := tx.ExecContext(ctx, `UPDATE periode_gaji SET status='DITUTUP',ditutup_pada=NOW() WHERE id=? AND status='DIHITUNG'`, id)
	if e != nil {
		return e
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return errors.New("status periode berubah; penutupan dibatalkan")
	}
	return tx.Commit()
}
func (r *Repository) Ringkasan(ctx context.Context, id int64) (RingkasanTutup, error) {
	var x RingkasanTutup
	e := r.db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(gaji_bersih),0),COALESCE(SUM(potongan_kasbon),0) FROM slip_gaji WHERE periode_id=?`, id).Scan(&x.JumlahSlip, &x.TotalBersih, &x.TotalPotonganKasbon)
	if e != nil {
		return x, e
	}
	e = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM koreksi_gaji WHERE periode_terapan_id=?`, id).Scan(&x.JumlahKoreksi)
	return x, e
}

// DaftarPekerjaKoreksi menyediakan pekerja untuk pilihan formulir koreksi.
func (r *Repository) DaftarPekerjaKoreksi(ctx context.Context) ([]domain.Pekerja, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,kode,nama,jenis,tanggal_mulai,tanggal_berhenti,aktif FROM pekerja ORDER BY nama,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	hasil := []domain.Pekerja{}
	for rows.Next() {
		var p domain.Pekerja
		var berhenti sql.NullTime
		if err := rows.Scan(&p.ID, &p.Kode, &p.Nama, &p.Jenis, &p.TanggalMulai, &berhenti, &p.Aktif); err != nil {
			return nil, err
		}
		if berhenti.Valid {
			p.TanggalBerhenti = &berhenti.Time
		}
		hasil = append(hasil, p)
	}
	return hasil, rows.Err()
}

func (r *Repository) SimpanKoreksi(ctx context.Context, k *domain.KoreksiGaji) error {
	res, e := r.db.ExecContext(ctx, `INSERT INTO koreksi_gaji(pekerja_id,periode_asal_id,jumlah,keterangan,dibuat_pada) VALUES(?,?,?,?,NOW())`, k.PekerjaID, k.PeriodeAsalID, k.Jumlah, k.Keterangan)
	if e != nil {
		return e
	}
	k.ID, e = res.LastInsertId()
	return e
}
func (r *Repository) DaftarKoreksi(ctx context.Context, pid *int64) ([]domain.KoreksiGaji, error) {
	q := `SELECT id,pekerja_id,periode_asal_id,periode_terapan_id,jumlah,keterangan,dibuat_pada FROM koreksi_gaji`
	args := []any{}
	if pid != nil {
		q += ` WHERE pekerja_id=?`
		args = append(args, *pid)
	}
	q += ` ORDER BY dibuat_pada DESC,id DESC`
	return r.queryKoreksi(ctx, q, args...)
}
func (r *Repository) DaftarKoreksiMenunggu(ctx context.Context, pid int64) ([]domain.KoreksiGaji, error) {
	return r.queryKoreksi(ctx, `SELECT id,pekerja_id,periode_asal_id,periode_terapan_id,jumlah,keterangan,dibuat_pada FROM koreksi_gaji WHERE pekerja_id=? AND periode_terapan_id IS NULL ORDER BY dibuat_pada,id`, pid)
}
func (r *Repository) queryKoreksi(ctx context.Context, q string, args ...any) ([]domain.KoreksiGaji, error) {
	rows, e := r.db.QueryContext(ctx, q, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.KoreksiGaji{}
	for rows.Next() {
		var k domain.KoreksiGaji
		var terapan sql.NullInt64
		if e := rows.Scan(&k.ID, &k.PekerjaID, &k.PeriodeAsalID, &terapan, &k.Jumlah, &k.Keterangan, &k.DibuatPada); e != nil {
			return nil, e
		}
		if terapan.Valid {
			v := terapan.Int64
			k.PeriodeTerapanID = &v
		}
		out = append(out, k)
	}
	return out, rows.Err()
}
func (r *Repository) TandaiKoreksiDiterapkan(ctx context.Context, id int64) error {
	_, e := r.db.ExecContext(ctx, `UPDATE koreksi_gaji k JOIN slip_gaji s ON s.pekerja_id=k.pekerja_id AND s.periode_id=? SET k.periode_terapan_id=? WHERE k.periode_terapan_id IS NULL AND k.periode_asal_id<>?`, id, id, id)
	return e
}
