// Pemilik: Anya | Modul 5 | Aturan: 1, 13, 14, 15, 18
package periode

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"penggajian-konveksi/internal/domain"
)

func tanggal(tahun int, bulan time.Month, hari int) time.Time {
	return time.Date(tahun, bulan, hari, 0, 0, 0, 0, time.Local)
}

// q mengubah SQL menjadi regexp literal agar karakter seperti * dan ? tidak
// dianggap sebagai bagian dari pola pencocokan sqlmock.
func q(sql string) string { return regexp.QuoteMeta(sql) }

func dbMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("membuat sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, mock
}

func svcMock(t *testing.T) (*Service, sqlmock.Sqlmock) {
	t.Helper()
	db, mock := dbMock(t)
	return NewService(NewRepository(db)), mock
}

func pastikanEkspektasi(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

func rowsPeriode() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "tipe", "tanggal_mulai", "tanggal_selesai", "status", "ditutup_pada"})
}

func TestBuatPeriodeRentang(t *testing.T) {
	kasus := []struct {
		nama    string
		tipe    domain.TipePeriode
		mulai   time.Time
		selesai time.Time
	}{
		{"mingguan Senin sampai Sabtu", domain.PeriodeMingguan, tanggal(2026, time.November, 2), tanggal(2026, time.November, 7)},
		{"bulanan November", domain.PeriodeBulanan, tanggal(2026, time.November, 1), tanggal(2026, time.November, 30)},
		{"bulanan Februari 2026", domain.PeriodeBulanan, tanggal(2026, time.February, 1), tanggal(2026, time.February, 28)},
		{"bulanan Februari kabisat", domain.PeriodeBulanan, tanggal(2028, time.February, 1), tanggal(2028, time.February, 29)},
	}

	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			svc, mock := svcMock(t)
			mock.ExpectQuery(q("SELECT COUNT(*) FROM periode_gaji WHERE tipe=? AND tanggal_mulai")).
				WithArgs(k.tipe, k.selesai.Format("2006-01-02"), k.mulai.Format("2006-01-02")).
				WillReturnRows(sqlmock.NewRows([]string{"jumlah"}).AddRow(0))
			mock.ExpectExec(q("INSERT INTO periode_gaji")).
				WithArgs(k.tipe, k.mulai.Format("2006-01-02"), k.selesai.Format("2006-01-02")).
				WillReturnResult(sqlmock.NewResult(12, 1))

			id, err := svc.BuatPeriode(context.Background(), k.tipe, k.mulai)
			if err != nil {
				t.Fatalf("BuatPeriode: %v", err)
			}
			if id != 12 {
				t.Fatalf("ID periode = %d, ingin 12", id)
			}
			pastikanEkspektasi(t, mock)
		})
	}
}

func TestBuatPeriodeValidasi(t *testing.T) {
	svc := NewService(NewRepository(nil))
	kasus := []struct {
		nama  string
		tipe  domain.TipePeriode
		mulai time.Time
	}{
		{"mingguan bukan Senin", domain.PeriodeMingguan, tanggal(2026, time.November, 3)},
		{"bulanan bukan tanggal 1", domain.PeriodeBulanan, tanggal(2026, time.November, 2)},
		{"tipe tidak dikenal", domain.TipePeriode("TAHUNAN"), tanggal(2026, time.November, 1)},
		{"tanggal mulai kosong", domain.PeriodeMingguan, time.Time{}},
	}
	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			if _, err := svc.BuatPeriode(context.Background(), k.tipe, k.mulai); err == nil {
				t.Fatal("periode tidak valid seharusnya ditolak")
			}
		})
	}
}

func TestBuatPeriodeMenolakTumpangTindih(t *testing.T) {
	svc, mock := svcMock(t)
	mock.ExpectQuery(q("SELECT COUNT(*) FROM periode_gaji WHERE tipe=? AND tanggal_mulai")).
		WithArgs(domain.PeriodeMingguan, "2026-11-07", "2026-11-02").
		WillReturnRows(sqlmock.NewRows([]string{"jumlah"}).AddRow(1))

	if _, err := svc.BuatPeriode(context.Background(), domain.PeriodeMingguan, tanggal(2026, time.November, 2)); err == nil {
		t.Fatal("periode yang tumpang tindih seharusnya ditolak")
	} else if !strings.Contains(err.Error(), "tumpang tindih") {
		t.Fatalf("pesan galat = %q, ingin memuat tumpang tindih", err.Error())
	}
	pastikanEkspektasi(t, mock)
}

func TestBuatPeriodeBolehTumpangTindihBedaTipe(t *testing.T) {
	svc, mock := svcMock(t)
	mock.ExpectQuery(q("SELECT COUNT(*) FROM periode_gaji WHERE tipe=? AND tanggal_mulai")).
		WithArgs(domain.PeriodeMingguan, "2026-11-07", "2026-11-02").
		WillReturnRows(sqlmock.NewRows([]string{"jumlah"}).AddRow(0))
	mock.ExpectExec(q("INSERT INTO periode_gaji")).
		WithArgs(domain.PeriodeMingguan, "2026-11-02", "2026-11-07").
		WillReturnResult(sqlmock.NewResult(13, 1))

	id, err := svc.BuatPeriode(context.Background(), domain.PeriodeMingguan, tanggal(2026, time.November, 2))
	if err != nil {
		t.Fatalf("BuatPeriode: %v", err)
	}
	if id != 13 {
		t.Fatalf("ID periode = %d, ingin 13", id)
	}
	pastikanEkspektasi(t, mock)
}

func TestBuatPeriodeTanpaDatabase(t *testing.T) {
	svc := NewService(NewRepository(nil))
	_, err := svc.BuatPeriode(context.Background(), domain.PeriodeMingguan, tanggal(2026, time.November, 2))
	if !errors.Is(err, ErrDatabasePeriodeTidakTersedia) {
		t.Fatalf("galat = %v, ingin %v", err, ErrDatabasePeriodeTidakTersedia)
	}
}

func TestPeriodeSebelumnyaDitutup(t *testing.T) {
	p := domain.PeriodeGaji{Tipe: domain.PeriodeMingguan, TanggalMulai: tanggal(2026, time.November, 9)}
	kasus := []struct {
		nama       string
		rows       *sqlmock.Rows
		ingin      bool
		keterangan string
	}{
		{"periode pertama", rowsPeriode(), true, "tidak ada periode sebelumnya"},
		{"periode sebelumnya ditutup", rowsPeriode().AddRow(1, "MINGGUAN", tanggal(2026, time.November, 2), tanggal(2026, time.November, 7), "DITUTUP", nil), true, ""},
		{"periode sebelumnya belum ditutup", rowsPeriode().AddRow(1, "MINGGUAN", tanggal(2026, time.November, 2), tanggal(2026, time.November, 7), "DIHITUNG", nil), false, ""},
	}

	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			svc, mock := svcMock(t)
			mock.ExpectQuery(q("SELECT id,tipe,tanggal_mulai,tanggal_selesai,status,ditutup_pada FROM periode_gaji WHERE tipe=? AND tanggal_selesai<?")).
				WithArgs(domain.PeriodeMingguan, "2026-11-09").
				WillReturnRows(k.rows)

			hasil, err := svc.PeriodeSebelumnyaDitutup(context.Background(), p)
			if err != nil {
				t.Fatalf("PeriodeSebelumnyaDitutup: %v", err)
			}
			if hasil != k.ingin {
				t.Fatalf("hasil = %v, ingin %v (%s)", hasil, k.ingin, k.keterangan)
			}
			pastikanEkspektasi(t, mock)
		})
	}
}

func TestPeriksaPeriodeTerkunci(t *testing.T) {
	for _, k := range []struct {
		nama   string
		jumlah int
		ingin  bool
	}{{"terkunci", 1, true}, {"belum terkunci", 0, false}} {
		t.Run(k.nama, func(t *testing.T) {
			svc, mock := svcMock(t)
			mock.ExpectQuery(q("SELECT COUNT(*) FROM periode_gaji WHERE status='DITUTUP'")).
				WithArgs("2026-11-04").
				WillReturnRows(sqlmock.NewRows([]string{"jumlah"}).AddRow(k.jumlah))

			hasil, err := svc.PeriksaPeriodeTerkunci(context.Background(), tanggal(2026, time.November, 4))
			if err != nil {
				t.Fatalf("PeriksaPeriodeTerkunci: %v", err)
			}
			if hasil != k.ingin {
				t.Fatalf("hasil = %v, ingin %v", hasil, k.ingin)
			}
			pastikanEkspektasi(t, mock)
		})
	}
}

func TestUbahStatusPeriode(t *testing.T) {
	svc, mock := svcMock(t)
	mock.ExpectBegin()
	mock.ExpectQuery(q("SELECT status FROM periode_gaji WHERE id=? FOR UPDATE")).
		WithArgs(int64(4)).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("TERBUKA"))
	mock.ExpectExec(q("UPDATE periode_gaji SET status=? WHERE id=?")).
		WithArgs(domain.PeriodeDihitung, int64(4)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := svc.UbahStatusPeriode(context.Background(), 4, domain.PeriodeDihitung); err != nil {
		t.Fatalf("UbahStatusPeriode: %v", err)
	}
	pastikanEkspektasi(t, mock)
}

func TestUbahStatusPeriodeDitolakSetelahDitutup(t *testing.T) {
	svc, mock := svcMock(t)
	mock.ExpectBegin()
	mock.ExpectQuery(q("SELECT status FROM periode_gaji WHERE id=? FOR UPDATE")).
		WithArgs(int64(4)).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("DITUTUP"))
	mock.ExpectRollback()

	err := svc.UbahStatusPeriode(context.Background(), 4, domain.PeriodeDihitung)
	if !errors.Is(err, ErrPeriodeSudahDitutup) {
		t.Fatalf("galat = %v, ingin %v", err, ErrPeriodeSudahDitutup)
	}
	pastikanEkspektasi(t, mock)
}

func TestUbahStatusTidakBolehLangsungDitutup(t *testing.T) {
	svc := NewService(NewRepository(nil))
	if err := svc.UbahStatusPeriode(context.Background(), 1, domain.PeriodeDitutup); err == nil {
		t.Fatal("status DITUTUP harus ditolak lewat UbahStatusPeriode")
	}
	if err := svc.UbahStatusPeriode(context.Background(), 1, domain.StatusPeriode("SALAH")); err == nil {
		t.Fatal("status tidak dikenal seharusnya ditolak")
	}
}

func TestTutupPeriodeBerhasil(t *testing.T) {
	svc, mock := svcMock(t)
	mulai := tanggal(2026, time.November, 2)
	selesai := tanggal(2026, time.November, 7)

	mock.ExpectQuery(q("SELECT id,tipe,tanggal_mulai,tanggal_selesai,status,ditutup_pada FROM periode_gaji WHERE id=?")).
		WithArgs(int64(5)).
		WillReturnRows(rowsPeriode().AddRow(5, "MINGGUAN", mulai, selesai, "DIHITUNG", nil))
	mock.ExpectQuery(q("SELECT id,tipe,tanggal_mulai,tanggal_selesai,status,ditutup_pada FROM periode_gaji WHERE tipe=? AND tanggal_selesai<?")).
		WithArgs(domain.PeriodeMingguan, "2026-11-02").
		WillReturnRows(rowsPeriode().AddRow(4, "MINGGUAN", tanggal(2026, time.October, 26), tanggal(2026, time.October, 31), "DITUTUP", tanggal(2026, time.October, 31)))

	mock.ExpectBegin()
	mock.ExpectQuery(q("SELECT status,tipe,tanggal_mulai FROM periode_gaji WHERE id=? FOR UPDATE")).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "tipe", "tanggal_mulai"}).AddRow("DIHITUNG", "MINGGUAN", mulai))
	mock.ExpectQuery(q("SELECT status FROM periode_gaji WHERE tipe=? AND tanggal_selesai<? ORDER BY tanggal_selesai DESC LIMIT 1 FOR UPDATE")).
		WithArgs(domain.PeriodeMingguan, "2026-11-02").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("DITUTUP"))
	mock.ExpectQuery(q("SELECT COALESCE(SUM(total_koreksi),0) FROM slip_gaji WHERE periode_id=?")).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(0))
	mock.ExpectQuery(q("SELECT COALESCE(SUM(jumlah),0) FROM koreksi_gaji WHERE periode_terapan_id=?")).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(0))
	mock.ExpectExec(q("UPDATE kasbon k JOIN")).
		WithArgs(int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(q("UPDATE periode_gaji SET status='DITUTUP',ditutup_pada=NOW() WHERE id=? AND status='DIHITUNG'")).
		WithArgs(int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := svc.TutupPeriode(context.Background(), 5); err != nil {
		t.Fatalf("TutupPeriode: %v", err)
	}
	pastikanEkspektasi(t, mock)
}

func TestTutupPeriodeTidakBisaDibatalkan(t *testing.T) {
	svc, mock := svcMock(t)
	mock.ExpectQuery(q("SELECT id,tipe,tanggal_mulai,tanggal_selesai,status,ditutup_pada FROM periode_gaji WHERE id=?")).
		WithArgs(int64(5)).
		WillReturnRows(rowsPeriode().AddRow(5, "MINGGUAN", tanggal(2026, time.November, 2), tanggal(2026, time.November, 7), "DITUTUP", tanggal(2026, time.November, 8)))

	err := svc.TutupPeriode(context.Background(), 5)
	if !errors.Is(err, ErrPeriodeSudahDitutup) {
		t.Fatalf("galat = %v, ingin %v", err, ErrPeriodeSudahDitutup)
	}
	pastikanEkspektasi(t, mock)
}

func TestTutupPeriodeHanyaDariDihitung(t *testing.T) {
	svc, mock := svcMock(t)
	mock.ExpectQuery(q("SELECT id,tipe,tanggal_mulai,tanggal_selesai,status,ditutup_pada FROM periode_gaji WHERE id=?")).
		WithArgs(int64(5)).
		WillReturnRows(rowsPeriode().AddRow(5, "MINGGUAN", tanggal(2026, time.November, 2), tanggal(2026, time.November, 7), "TERBUKA", nil))

	if err := svc.TutupPeriode(context.Background(), 5); err == nil {
		t.Fatal("periode TERBUKA seharusnya tidak dapat ditutup")
	}
	pastikanEkspektasi(t, mock)
}

func TestTutupPeriodeMenolakPeriodeSebelumnyaBelumDitutup(t *testing.T) {
	svc, mock := svcMock(t)
	mock.ExpectQuery(q("SELECT id,tipe,tanggal_mulai,tanggal_selesai,status,ditutup_pada FROM periode_gaji WHERE id=?")).
		WithArgs(int64(5)).
		WillReturnRows(rowsPeriode().AddRow(5, "MINGGUAN", tanggal(2026, time.November, 2), tanggal(2026, time.November, 7), "DIHITUNG", nil))
	mock.ExpectQuery(q("SELECT id,tipe,tanggal_mulai,tanggal_selesai,status,ditutup_pada FROM periode_gaji WHERE tipe=? AND tanggal_selesai<?")).
		WithArgs(domain.PeriodeMingguan, "2026-11-02").
		WillReturnRows(rowsPeriode().AddRow(4, "MINGGUAN", tanggal(2026, time.October, 26), tanggal(2026, time.October, 31), "TERBUKA", nil))

	err := svc.TutupPeriode(context.Background(), 5)
	if err == nil {
		t.Fatal("periode sebelumnya yang belum ditutup seharusnya menolak penutupan")
	}
	pastikanEkspektasi(t, mock)
}

func TestTutupPeriodeTanpaDatabase(t *testing.T) {
	svc := NewService(NewRepository(nil))
	if err := svc.TutupPeriode(context.Background(), 1); !errors.Is(err, ErrDatabasePeriodeTidakTersedia) {
		t.Fatalf("galat = %v, ingin %v", err, ErrDatabasePeriodeTidakTersedia)
	}
}

func TestRingkasanTutupSebelumPenutupan(t *testing.T) {
	svc, mock := svcMock(t)
	mock.ExpectQuery(q("SELECT COUNT(*),COALESCE(SUM(gaji_bersih),0),COALESCE(SUM(potongan_kasbon),0) FROM slip_gaji WHERE periode_id=?")).
		WithArgs(int64(4)).
		WillReturnRows(sqlmock.NewRows([]string{"jumlah", "bersih", "potongan"}).AddRow(2, 500000, 100000))
	mock.ExpectQuery(q("SELECT COUNT(*) FROM koreksi_gaji WHERE periode_terapan_id=?")).
		WithArgs(int64(4)).
		WillReturnRows(sqlmock.NewRows([]string{"jumlah"}).AddRow(1))

	ring, err := svc.RingkasanTutupPeriode(context.Background(), 4)
	if err != nil {
		t.Fatalf("RingkasanTutupPeriode: %v", err)
	}
	if ring.JumlahSlip != 2 || ring.TotalBersih != 500000 || ring.TotalPotonganKasbon != 100000 || ring.JumlahKoreksi != 1 {
		t.Fatalf("ringkasan = %+v", ring)
	}
	pastikanEkspektasi(t, mock)
}

func TestCatatKoreksiValidasi(t *testing.T) {
	svc := NewService(NewRepository(nil))
	kasus := []struct {
		nama string
		data domain.KoreksiGaji
	}{
		{"pekerja kosong", domain.KoreksiGaji{PeriodeAsalID: 1, Jumlah: -1000, Keterangan: "x"}},
		{"nominal nol", domain.KoreksiGaji{PekerjaID: 1, PeriodeAsalID: 1, Jumlah: 0, Keterangan: "x"}},
		{"alasan kosong", domain.KoreksiGaji{PekerjaID: 1, PeriodeAsalID: 1, Jumlah: -1000, Keterangan: "   "}},
		{"alasan terlalu panjang", domain.KoreksiGaji{PekerjaID: 1, PeriodeAsalID: 1, Jumlah: -1000, Keterangan: strings.Repeat("a", 201)}},
	}
	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			if _, err := svc.CatatKoreksi(context.Background(), k.data); err == nil {
				t.Fatal("koreksi tidak valid seharusnya ditolak")
			}
		})
	}
}

func TestCatatKoreksiHanyaUntukPeriodeDitutup(t *testing.T) {
	svc, mock := svcMock(t)
	mock.ExpectQuery(q("SELECT id,tipe,tanggal_mulai,tanggal_selesai,status,ditutup_pada FROM periode_gaji WHERE id=?")).
		WithArgs(int64(1)).
		WillReturnRows(rowsPeriode().AddRow(1, "MINGGUAN", tanggal(2026, time.November, 2), tanggal(2026, time.November, 7), "DIHITUNG", nil))

	_, err := svc.CatatKoreksi(context.Background(), domain.KoreksiGaji{PekerjaID: 1, PeriodeAsalID: 1, Jumlah: -50000, Keterangan: "Koreksi upah"})
	if err == nil || !strings.Contains(err.Error(), "sudah ditutup") {
		t.Fatalf("galat = %v, ingin penolakan periode belum ditutup", err)
	}
	pastikanEkspektasi(t, mock)
}

func TestCatatKoreksiBerhasil(t *testing.T) {
	svc, mock := svcMock(t)
	mock.ExpectQuery(q("SELECT id,tipe,tanggal_mulai,tanggal_selesai,status,ditutup_pada FROM periode_gaji WHERE id=?")).
		WithArgs(int64(1)).
		WillReturnRows(rowsPeriode().AddRow(1, "MINGGUAN", tanggal(2026, time.October, 26), tanggal(2026, time.October, 31), "DITUTUP", tanggal(2026, time.November, 1)))
	mock.ExpectQuery(q("SELECT COUNT(*) FROM pekerja WHERE id=?")).
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"jumlah"}).AddRow(1))
	mock.ExpectExec(q("INSERT INTO koreksi_gaji")).
		WithArgs(int64(2), int64(1), int64(-50000), "Koreksi upah").
		WillReturnResult(sqlmock.NewResult(9, 1))

	id, err := svc.CatatKoreksi(context.Background(), domain.KoreksiGaji{
		PekerjaID:     2,
		PeriodeAsalID: 1,
		Jumlah:        -50000,
		Keterangan:    "  Koreksi upah  ",
	})
	if err != nil {
		t.Fatalf("CatatKoreksi: %v", err)
	}
	if id != 9 {
		t.Fatalf("ID koreksi = %d, ingin 9", id)
	}
	pastikanEkspektasi(t, mock)
}

func TestCatatKoreksiPekerjaTidakAda(t *testing.T) {
	svc, mock := svcMock(t)
	mock.ExpectQuery(q("SELECT id,tipe,tanggal_mulai,tanggal_selesai,status,ditutup_pada FROM periode_gaji WHERE id=?")).
		WithArgs(int64(1)).
		WillReturnRows(rowsPeriode().AddRow(1, "MINGGUAN", tanggal(2026, time.October, 26), tanggal(2026, time.October, 31), "DITUTUP", tanggal(2026, time.November, 1)))
	mock.ExpectQuery(q("SELECT COUNT(*) FROM pekerja WHERE id=?")).
		WithArgs(int64(99)).
		WillReturnRows(sqlmock.NewRows([]string{"jumlah"}).AddRow(0))

	_, err := svc.CatatKoreksi(context.Background(), domain.KoreksiGaji{PekerjaID: 99, PeriodeAsalID: 1, Jumlah: -50000, Keterangan: "Koreksi"})
	if err == nil || !strings.Contains(err.Error(), "pekerja tidak ditemukan") {
		t.Fatalf("galat = %v, ingin pekerja tidak ditemukan", err)
	}
	pastikanEkspektasi(t, mock)
}

func TestDaftarKoreksiMenunggu(t *testing.T) {
	svc, mock := svcMock(t)
	dibuat := tanggal(2026, time.November, 1)
	mock.ExpectQuery(q("SELECT id,pekerja_id,periode_asal_id,periode_terapan_id,jumlah,keterangan,dibuat_pada FROM koreksi_gaji WHERE pekerja_id=? AND periode_terapan_id IS NULL")).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "pekerja_id", "periode_asal_id", "periode_terapan_id", "jumlah", "keterangan", "dibuat_pada"}).
			AddRow(1, 7, 2, nil, -10000, "Potong upah", dibuat).
			AddRow(2, 7, 2, nil, 5000, "Bonus", dibuat))

	daftar, err := svc.DaftarKoreksiMenunggu(context.Background(), 7)
	if err != nil {
		t.Fatalf("DaftarKoreksiMenunggu: %v", err)
	}
	if len(daftar) != 2 {
		t.Fatalf("jumlah koreksi = %d, ingin 2", len(daftar))
	}
	if daftar[0].Status() != "Menunggu" || daftar[1].PeriodeTerapanID != nil {
		t.Fatalf("status koreksi menunggu tidak sesuai: %+v", daftar)
	}
	pastikanEkspektasi(t, mock)
}

func TestTandaiKoreksiDiterapkan(t *testing.T) {
	svc, mock := svcMock(t)
	mock.ExpectExec(q("UPDATE koreksi_gaji k JOIN slip_gaji s ON s.pekerja_id=k.pekerja_id AND s.periode_id=?")).
		WithArgs(int64(3), int64(3), int64(3)).
		WillReturnResult(sqlmock.NewResult(0, 2))

	if err := svc.TandaiKoreksiDiterapkan(context.Background(), 3); err != nil {
		t.Fatalf("TandaiKoreksiDiterapkan: %v", err)
	}
	pastikanEkspektasi(t, mock)
}

func TestPeriksaPeriodeTerkunciJenis(t *testing.T) {
	svc, mock := svcMock(t)
	mock.ExpectQuery(q("SELECT COUNT(*) FROM periode_gaji WHERE status='DITUTUP' AND tipe=? AND ? BETWEEN tanggal_mulai AND tanggal_selesai")).
		WithArgs(domain.PeriodeBulanan, "2026-11-05").
		WillReturnRows(sqlmock.NewRows([]string{"jumlah"}).AddRow(0))

	terkunci, err := svc.PeriksaPeriodeTerkunciJenis(context.Background(), domain.PekerjaBulanan, tanggal(2026, time.November, 5))
	if err != nil {
		t.Fatalf("PeriksaPeriodeTerkunciJenis: %v", err)
	}
	if terkunci {
		t.Fatal("pekerja bulanan seharusnya tidak terkunci ketika hanya periode mingguan yang ditutup")
	}
	pastikanEkspektasi(t, mock)
}

func TestTutupPeriodeMenolakKoreksiTidakSinkron(t *testing.T) {
	svc, mock := svcMock(t)
	mulai := tanggal(2026, time.November, 9)
	selesai := tanggal(2026, time.November, 14)

	mock.ExpectQuery(q("SELECT id,tipe,tanggal_mulai,tanggal_selesai,status,ditutup_pada FROM periode_gaji WHERE id=?")).
		WithArgs(int64(6)).
		WillReturnRows(rowsPeriode().AddRow(6, "MINGGUAN", mulai, selesai, "DIHITUNG", nil))
	mock.ExpectQuery(q("SELECT id,tipe,tanggal_mulai,tanggal_selesai,status,ditutup_pada FROM periode_gaji WHERE tipe=? AND tanggal_selesai<?")).
		WithArgs(domain.PeriodeMingguan, "2026-11-09").
		WillReturnRows(rowsPeriode())
	mock.ExpectBegin()
	mock.ExpectQuery(q("SELECT status,tipe,tanggal_mulai FROM periode_gaji WHERE id=? FOR UPDATE")).
		WithArgs(int64(6)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "tipe", "tanggal_mulai"}).AddRow("DIHITUNG", "MINGGUAN", mulai))
	mock.ExpectQuery(q("SELECT status FROM periode_gaji WHERE tipe=? AND tanggal_selesai<? ORDER BY tanggal_selesai DESC LIMIT 1 FOR UPDATE")).
		WithArgs(domain.PeriodeMingguan, "2026-11-09").
		WillReturnRows(sqlmock.NewRows([]string{"status"}))
	mock.ExpectQuery(q("SELECT COALESCE(SUM(total_koreksi),0) FROM slip_gaji WHERE periode_id=?")).
		WithArgs(int64(6)).
		WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(50000))
	mock.ExpectQuery(q("SELECT COALESCE(SUM(jumlah),0) FROM koreksi_gaji WHERE periode_terapan_id=?")).
		WithArgs(int64(6)).
		WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(0))
	mock.ExpectRollback()

	if err := svc.TutupPeriode(context.Background(), 6); !errors.Is(err, ErrKoreksiTidakSinkron) {
		t.Fatalf("galat = %v, ingin ErrKoreksiTidakSinkron", err)
	}
	pastikanEkspektasi(t, mock)
}

func TestKoreksiTidakDihitungDuaKali(t *testing.T) {
	k := domain.KoreksiGaji{PeriodeTerapanID: nil}
	if k.Status() != "Menunggu" {
		t.Fatal("koreksi baru harus berstatus Menunggu")
	}
	id := int64(2)
	k.PeriodeTerapanID = &id
	if k.Status() != "Diterapkan" {
		t.Fatal("koreksi yang memiliki periode penerapan harus Diterapkan")
	}
}
