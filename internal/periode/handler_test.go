// Pemilik: Anya | Modul 5 | Aturan: 13, 15
package periode

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"penggajian-konveksi/internal/domain"
	"penggajian-konveksi/internal/web"
)

func muxMock(t *testing.T) (*http.ServeMux, sqlmock.Sqlmock) {
	t.Helper()
	db, mock := dbMock(t)
	renderer := &web.Render{Dir: filepath.Join("..", "..", "web", "templates")}
	mux := http.NewServeMux()
	DaftarkanRute(mux, renderer, NewService(NewRepository(db)))
	return mux, mock
}

func panggil(mux http.Handler, metode, target string, form url.Values) *httptest.ResponseRecorder {
	body := strings.NewReader("")
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req := httptest.NewRequest(metode, target, body)
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func harapanDaftarPeriode(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(q("SELECT id,tipe,tanggal_mulai,tanggal_selesai,status,ditutup_pada FROM periode_gaji ORDER BY tanggal_mulai DESC,id DESC")).
		WillReturnRows(rowsPeriode().
			AddRow(3, "MINGGUAN", tanggal(2026, time.November, 9), tanggal(2026, time.November, 14), "TERBUKA", nil).
			AddRow(2, "BULANAN", tanggal(2026, time.October, 1), tanggal(2026, time.October, 31), "DITUTUP", tanggal(2026, time.November, 1)).
			AddRow(1, "MINGGUAN", tanggal(2026, time.November, 2), tanggal(2026, time.November, 7), "DIHITUNG", nil))
}

func harapanDaftarKoreksi(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(q("SELECT id,pekerja_id,periode_asal_id,periode_terapan_id,jumlah,keterangan,dibuat_pada FROM koreksi_gaji ORDER BY dibuat_pada DESC,id DESC")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "pekerja_id", "periode_asal_id", "periode_terapan_id", "jumlah", "keterangan", "dibuat_pada"}).
			AddRow(1, 2, 2, nil, -10000, "Koreksi upah", tanggal(2026, time.November, 2)))
}

func harapanDaftarPekerja(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(q("SELECT id,kode,nama,jenis,tanggal_mulai,tanggal_berhenti,aktif FROM pekerja ORDER BY nama,id")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "kode", "nama", "jenis", "tanggal_mulai", "tanggal_berhenti", "aktif"}).
			AddRow(2, "P-001", "Budi", "HARIAN", tanggal(2026, time.January, 1), nil, true))
}

func TestHalamanPeriodeMenampilkanData(t *testing.T) {
	mux, mock := muxMock(t)
	harapanDaftarPeriode(mock)

	rr := panggil(mux, http.MethodGet, "/periode", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200; body: %s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	for _, ingin := range []string{"Daftar periode", "MINGGUAN", "02-11-2026", "Tutup periode", "Buat periode", `action="/periode/3/hitung"`, "Hitung gaji", `action="/periode/1/hitung"`, "Hitung ulang"} {
		if !strings.Contains(body, ingin) {
			t.Fatalf("halaman tidak memuat %q", ingin)
		}
	}
	pastikanEkspektasi(t, mock)
}

func TestHalamanFormPeriodeBaru(t *testing.T) {
	mux, mock := muxMock(t)
	harapanDaftarPeriode(mock)

	rr := panggil(mux, http.MethodGet, "/periode/baru", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `action="/periode/baru"`) {
		t.Fatal("form periode baru tidak ditemukan")
	}
	pastikanEkspektasi(t, mock)
}

func TestHalamanTutupPeriodeMenampilkanData(t *testing.T) {
	mux, mock := muxMock(t)
	mock.ExpectQuery(q("SELECT id,tipe,tanggal_mulai,tanggal_selesai,status,ditutup_pada FROM periode_gaji WHERE id=?")).
		WithArgs(int64(1)).
		WillReturnRows(rowsPeriode().AddRow(1, "MINGGUAN", tanggal(2026, time.November, 2), tanggal(2026, time.November, 7), "DIHITUNG", nil))
	mock.ExpectQuery(q("SELECT COUNT(*),COALESCE(SUM(gaji_bersih),0),COALESCE(SUM(potongan_kasbon),0) FROM slip_gaji WHERE periode_id=?")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"jumlah", "bersih", "potongan"}).AddRow(2, 500000, 100000))
	mock.ExpectQuery(q("SELECT COUNT(*) FROM koreksi_gaji WHERE periode_terapan_id=?")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"jumlah"}).AddRow(1))
	mock.ExpectQuery(q("SELECT id,tipe,tanggal_mulai,tanggal_selesai,status,ditutup_pada FROM periode_gaji WHERE tipe=? AND tanggal_selesai<?")).
		WithArgs(domain.PeriodeMingguan, "2026-11-02").
		WillReturnRows(rowsPeriode())

	rr := panggil(mux, http.MethodGet, "/periode/1/tutup", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200; body: %s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	for _, ingin := range []string{"Ringkasan penutupan periode 1", "Rp 500000", "Koreksi diterapkan", "Saya sudah membayarkan gaji periode ini"} {
		if !strings.Contains(body, ingin) {
			t.Fatalf("halaman tidak memuat %q", ingin)
		}
	}
	pastikanEkspektasi(t, mock)
}

func TestHalamanTutupPeriodeSudahDitutup(t *testing.T) {
	mux, mock := muxMock(t)
	mock.ExpectQuery(q("SELECT id,tipe,tanggal_mulai,tanggal_selesai,status,ditutup_pada FROM periode_gaji WHERE id=?")).
		WithArgs(int64(1)).
		WillReturnRows(rowsPeriode().AddRow(1, "MINGGUAN", tanggal(2026, time.November, 2), tanggal(2026, time.November, 7), "DITUTUP", tanggal(2026, time.November, 8)))
	mock.ExpectQuery(q("SELECT COUNT(*),COALESCE(SUM(gaji_bersih),0),COALESCE(SUM(potongan_kasbon),0) FROM slip_gaji WHERE periode_id=?")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"jumlah", "bersih", "potongan"}).AddRow(2, 500000, 100000))
	mock.ExpectQuery(q("SELECT COUNT(*) FROM koreksi_gaji WHERE periode_terapan_id=?")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"jumlah"}).AddRow(1))

	rr := panggil(mux, http.MethodGet, "/periode/1/tutup", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "tidak dapat ditutup") {
		t.Fatal("halaman seharusnya menyatakan periode sudah ditutup")
	}
	if strings.Contains(body, `action="/periode/1/tutup"`) {
		t.Fatal("form penutupan tidak boleh muncul untuk periode yang sudah ditutup")
	}
	pastikanEkspektasi(t, mock)
}

func TestHalamanTutupPeriodeTidakDitemukan(t *testing.T) {
	mux, mock := muxMock(t)
	mock.ExpectQuery(q("SELECT id,tipe,tanggal_mulai,tanggal_selesai,status,ditutup_pada FROM periode_gaji WHERE id=?")).
		WithArgs(int64(99)).
		WillReturnRows(rowsPeriode())

	rr := panggil(mux, http.MethodGet, "/periode/99/tutup", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, ingin 404", rr.Code)
	}
	pastikanEkspektasi(t, mock)
}

func TestHalamanKoreksiMenampilkanData(t *testing.T) {
	mux, mock := muxMock(t)
	harapanDaftarKoreksi(mock)
	harapanDaftarPeriode(mock)
	harapanDaftarPekerja(mock)

	rr := panggil(mux, http.MethodGet, "/koreksi", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200; body: %s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	for _, ingin := range []string{"Koreksi upah", "Budi", "Koreksi baru", "Menunggu"} {
		if !strings.Contains(body, ingin) {
			t.Fatalf("halaman tidak memuat %q", ingin)
		}
	}
	pastikanEkspektasi(t, mock)
}

func TestHalamanFormKoreksiBaru(t *testing.T) {
	mux, mock := muxMock(t)
	harapanDaftarKoreksi(mock)
	harapanDaftarPeriode(mock)
	harapanDaftarPekerja(mock)

	rr := panggil(mux, http.MethodGet, "/koreksi/baru", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `action="/koreksi/baru"`) {
		t.Fatal("form koreksi baru tidak ditemukan")
	}
	pastikanEkspektasi(t, mock)
}

func TestSimpanPeriodeBerhasil(t *testing.T) {
	mux, mock := muxMock(t)
	mock.ExpectQuery(q("SELECT COUNT(*) FROM periode_gaji WHERE tipe=? AND tanggal_mulai")).
		WithArgs(domain.PeriodeMingguan, "2026-11-07", "2026-11-02").
		WillReturnRows(sqlmock.NewRows([]string{"jumlah"}).AddRow(0))
	mock.ExpectExec(q("INSERT INTO periode_gaji")).
		WithArgs(domain.PeriodeMingguan, "2026-11-02", "2026-11-07").
		WillReturnResult(sqlmock.NewResult(1, 1))

	rr := panggil(mux, http.MethodPost, "/periode/baru", url.Values{
		"tipe":          {"MINGGUAN"},
		"tanggal_mulai": {"2026-11-02"},
	})
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, ingin 303; body: %s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get("Location") != "/periode" {
		t.Fatalf("Location = %q, ingin /periode", rr.Header().Get("Location"))
	}
	pastikanEkspektasi(t, mock)
}

func TestAksiHitungMengarahkanKePenggajianDenganPeriode(t *testing.T) {
	mux, _ := muxMock(t)
	rr := panggil(mux, http.MethodPost, "/periode/7/hitung", nil)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, ingin 303", rr.Code)
	}
	if rr.Header().Get("Location") != "/penggajian?periode_id=7" {
		t.Fatalf("Location = %q, ingin /penggajian?periode_id=7", rr.Header().Get("Location"))
	}
}

func TestAksiHitungMenolakIDTidakValid(t *testing.T) {
	mux, _ := muxMock(t)
	rr := panggil(mux, http.MethodPost, "/periode/abc/hitung", nil)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, ingin 400", rr.Code)
	}
}

func TestSimpanTutupMenolakTanpaKonfirmasi(t *testing.T) {
	mux, _ := muxMock(t)
	rr := panggil(mux, http.MethodPost, "/periode/1/tutup", url.Values{})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, ingin 400", rr.Code)
	}
}

func TestSimpanTutupBerhasil(t *testing.T) {
	mux, mock := muxMock(t)
	mulai := tanggal(2026, time.November, 2)
	selesai := tanggal(2026, time.November, 7)
	mock.ExpectQuery(q("SELECT id,tipe,tanggal_mulai,tanggal_selesai,status,ditutup_pada FROM periode_gaji WHERE id=?")).
		WithArgs(int64(5)).
		WillReturnRows(rowsPeriode().AddRow(5, "MINGGUAN", mulai, selesai, "DIHITUNG", nil))
	mock.ExpectQuery(q("SELECT id,tipe,tanggal_mulai,tanggal_selesai,status,ditutup_pada FROM periode_gaji WHERE tipe=? AND tanggal_selesai<?")).
		WithArgs(domain.PeriodeMingguan, "2026-11-02").
		WillReturnRows(rowsPeriode())
	mock.ExpectBegin()
	mock.ExpectQuery(q("SELECT status,tipe,tanggal_mulai FROM periode_gaji WHERE id=? FOR UPDATE")).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "tipe", "tanggal_mulai"}).AddRow("DIHITUNG", "MINGGUAN", mulai))
	mock.ExpectQuery(q("SELECT status FROM periode_gaji WHERE tipe=? AND tanggal_selesai<? ORDER BY tanggal_selesai DESC LIMIT 1 FOR UPDATE")).
		WithArgs(domain.PeriodeMingguan, "2026-11-02").
		WillReturnRows(sqlmock.NewRows([]string{"status"}))
	mock.ExpectQuery(q("SELECT COALESCE(SUM(total_koreksi),0) FROM slip_gaji WHERE periode_id=?")).WithArgs(int64(5)).WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(0))
	mock.ExpectQuery(q("SELECT COALESCE(SUM(jumlah),0) FROM koreksi_gaji WHERE periode_terapan_id=?")).WithArgs(int64(5)).WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(0))
	mock.ExpectExec(q("UPDATE kasbon k JOIN")).WithArgs(int64(5)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(q("UPDATE periode_gaji SET status='DITUTUP',ditutup_pada=NOW() WHERE id=? AND status='DIHITUNG'")).WithArgs(int64(5)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	rr := panggil(mux, http.MethodPost, "/periode/5/tutup", url.Values{"sudah_dibayar": {"1"}})
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, ingin 303; body: %s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get("Location") != "/periode" {
		t.Fatalf("Location = %q, ingin /periode", rr.Header().Get("Location"))
	}
	pastikanEkspektasi(t, mock)
}

func TestSimpanKoreksiBerhasil(t *testing.T) {
	mux, mock := muxMock(t)
	mock.ExpectQuery(q("SELECT id,tipe,tanggal_mulai,tanggal_selesai,status,ditutup_pada FROM periode_gaji WHERE id=?")).
		WithArgs(int64(2)).
		WillReturnRows(rowsPeriode().AddRow(2, "BULANAN", tanggal(2026, time.October, 1), tanggal(2026, time.October, 31), "DITUTUP", tanggal(2026, time.November, 1)))
	mock.ExpectQuery(q("SELECT COUNT(*) FROM pekerja WHERE id=?")).
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"jumlah"}).AddRow(1))
	mock.ExpectExec(q("INSERT INTO koreksi_gaji")).
		WithArgs(int64(2), int64(2), int64(-10000), "Koreksi upah").
		WillReturnResult(sqlmock.NewResult(1, 1))

	rr := panggil(mux, http.MethodPost, "/koreksi/baru", url.Values{
		"pekerja_id":      {"2"},
		"periode_asal_id": {"2"},
		"jumlah":          {"-10000"},
		"keterangan":      {"Koreksi upah"},
	})
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, ingin 303; body: %s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get("Location") != "/koreksi" {
		t.Fatalf("Location = %q, ingin /koreksi", rr.Header().Get("Location"))
	}
	pastikanEkspektasi(t, mock)
}

func TestSimpanKoreksiMenolakPeriodeBelumDitutup(t *testing.T) {
	mux, mock := muxMock(t)
	mock.ExpectQuery(q("SELECT id,tipe,tanggal_mulai,tanggal_selesai,status,ditutup_pada FROM periode_gaji WHERE id=?")).
		WithArgs(int64(2)).
		WillReturnRows(rowsPeriode().AddRow(2, "BULANAN", tanggal(2026, time.November, 1), tanggal(2026, time.November, 30), "DIHITUNG", nil))

	rr := panggil(mux, http.MethodPost, "/koreksi/baru", url.Values{
		"pekerja_id":      {"2"},
		"periode_asal_id": {"2"},
		"jumlah":          {"-10000"},
		"keterangan":      {"Koreksi upah"},
	})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, ingin 400", rr.Code)
	}
	pastikanEkspektasi(t, mock)
}
