// Pemilik: Mahdi | Modul 1 | Aturan: 8, 9, 16
package pekerja

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"penggajian-konveksi/internal/domain"
	"penggajian-konveksi/internal/web"
)

func newTestMux(t *testing.T) (*http.ServeMux, *fakeRepository) {
	t.Helper()
	repo := newFakeRepository()
	renderer := &web.Render{Dir: filepath.Join("..", "..", "web", "templates")}
	mux := http.NewServeMux()
	DaftarkanRute(mux, renderer, NewService(repo))
	return mux, repo
}

func minta(mux *http.ServeMux, target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rekaman := httptest.NewRecorder()
	mux.ServeHTTP(rekaman, req)
	return rekaman
}

func kirimForm(mux *http.ServeMux, target string, nilai url.Values) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(nilai.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rekaman := httptest.NewRecorder()
	mux.ServeHTTP(rekaman, req)
	return rekaman
}

func TestHalamanDaftarPekerjaMenampilkanData(t *testing.T) {
	mux, repo := newTestMux(t)
	seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, time.November, 2), Aktif: true,
	})
	seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-002", Nama: "Siti", Jenis: domain.PekerjaBulanan,
		TanggalMulai: tanggal(2026, time.November, 3), Aktif: true,
	})

	hasil := minta(mux, "/pekerja")
	if hasil.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", hasil.Code)
	}
	body := hasil.Body.String()
	for _, teks := range []string{"P-001", "Budi", "P-002", "Siti", "02 Nov 2026"} {
		if !strings.Contains(body, teks) {
			t.Fatalf("halaman tidak memuat %q", teks)
		}
	}
}

func TestHalamanDaftarPekerjaMenerapkanFilter(t *testing.T) {
	mux, repo := newTestMux(t)
	seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, time.November, 2), Aktif: true,
	})
	seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-002", Nama: "Siti", Jenis: domain.PekerjaBulanan,
		TanggalMulai: tanggal(2026, time.November, 3), Aktif: true,
	})

	hasil := minta(mux, "/pekerja?cari=Budi&jenis=HARIAN&aktif=1")
	if hasil.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", hasil.Code)
	}
	body := hasil.Body.String()
	if !strings.Contains(body, "Budi") {
		t.Fatal("hasil filter harus memuat Budi")
	}
	if strings.Contains(body, "Siti") {
		t.Fatal("hasil filter tidak boleh memuat Siti")
	}
	if !strings.Contains(body, `value="Budi"`) {
		t.Fatal("nilai filter cari harus dipertahankan")
	}
	if repo.lastFilter.Cari != "Budi" || repo.lastFilter.Jenis != domain.PekerjaHarian {
		t.Fatalf("filter diteruskan = %+v", repo.lastFilter)
	}
	if repo.lastFilter.Aktif == nil || !*repo.lastFilter.Aktif {
		t.Fatalf("filter aktif diteruskan = %v", repo.lastFilter.Aktif)
	}
}

func TestHalamanDaftarPekerjaMenerimaFilterJenisTidakValid(t *testing.T) {
	mux, _ := newTestMux(t)
	hasil := minta(mux, "/pekerja?jenis=XYZ")
	if hasil.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", hasil.Code)
	}
}

func TestHalamanFormBaruDanSimpan(t *testing.T) {
	mux, repo := newTestMux(t)

	halaman := minta(mux, "/pekerja/baru")
	if halaman.Code != http.StatusOK {
		t.Fatalf("status GET = %d, ingin 200", halaman.Code)
	}
	if !strings.Contains(halaman.Body.String(), `action="/pekerja/baru"`) {
		t.Fatal("form baru harus mengarah ke /pekerja/baru")
	}

	simpan := kirimForm(mux, "/pekerja/baru", url.Values{
		"kode":             {"P-009"},
		"nama":             {"Rina"},
		"jenis":            {"HARIAN"},
		"tanggal_mulai":    {"2026-11-10"},
		"upah_awal":        {"150000"},
		"tanggal_berhenti": {""},
	})
	if simpan.Code != http.StatusSeeOther {
		t.Fatalf("status POST = %d, ingin 303; body: %s", simpan.Code, simpan.Body.String())
	}
	if !strings.HasPrefix(simpan.Header().Get("Location"), "/pekerja?") {
		t.Fatalf("Location = %q, ingin /pekerja?info=...", simpan.Header().Get("Location"))
	}

	if len(repo.pekerja) != 1 {
		t.Fatalf("jumlah pekerja = %d, ingin 1", len(repo.pekerja))
	}
	var baru domain.Pekerja
	for _, pekerja := range repo.pekerja {
		baru = pekerja
	}
	if baru.Kode != "P-009" || baru.Nama != "Rina" || baru.Jenis != domain.PekerjaHarian {
		t.Fatalf("pekerja tersimpan = %+v", baru)
	}
	if !baru.Aktif {
		t.Fatal("pekerja baru harus aktif")
	}
	upah := repo.upah[baru.ID]
	if len(upah) != 1 || upah[0].Nominal != 150000 {
		t.Fatalf("upah awal tersimpan = %+v", upah)
	}
}

func TestHalamanFormUbahDanSimpan(t *testing.T) {
	mux, repo := newTestMux(t)
	lama := seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, time.November, 2), Aktif: true,
	})

	halaman := minta(mux, "/pekerja/1/ubah")
	if halaman.Code != http.StatusOK {
		t.Fatalf("status GET = %d, ingin 200", halaman.Code)
	}
	body := halaman.Body.String()
	for _, teks := range []string{"P-001", "Budi", `value="2026-11-02"`} {
		if !strings.Contains(body, teks) {
			t.Fatalf("form ubah tidak memuat %q", teks)
		}
	}
	if !strings.Contains(body, `action="/pekerja/1/ubah"`) {
		t.Fatal("form ubah harus mengarah ke /pekerja/1/ubah")
	}

	simpan := kirimForm(mux, "/pekerja/1/ubah", url.Values{
		"kode":             {"P-001"},
		"nama":             {"Budi Santoso"},
		"tanggal_mulai":    {"2026-11-02"},
		"tanggal_berhenti": {""},
	})
	if simpan.Code != http.StatusSeeOther {
		t.Fatalf("status POST = %d, ingin 303; body: %s", simpan.Code, simpan.Body.String())
	}
	hasil := repo.pekerja[lama.ID]
	if hasil.Nama != "Budi Santoso" {
		t.Fatalf("nama = %q, ingin %q", hasil.Nama, "Budi Santoso")
	}
	if hasil.Jenis != domain.PekerjaHarian {
		t.Fatalf("jenis berubah menjadi %q", hasil.Jenis)
	}
}

func TestHalamanFormUbahPekerjaTidakAda(t *testing.T) {
	mux, _ := newTestMux(t)
	hasil := minta(mux, "/pekerja/99/ubah")
	if hasil.Code != http.StatusNotFound {
		t.Fatalf("status = %d, ingin 404", hasil.Code)
	}
}

func TestHalamanFormUbahDenganTanggalBerhenti(t *testing.T) {
	mux, repo := newTestMux(t)
	pekerja := seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, time.November, 2), Aktif: true,
	})

	simpan := kirimForm(mux, "/pekerja/1/ubah", url.Values{
		"kode":             {"P-001"},
		"nama":             {"Budi"},
		"jenis":            {"HARIAN"},
		"tanggal_mulai":    {"2026-11-02"},
		"tanggal_berhenti": {"2026-11-20"},
	})
	if simpan.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, ingin 303; body: %s", simpan.Code, simpan.Body.String())
	}
	hasil := repo.pekerja[pekerja.ID]
	if hasil.Aktif {
		t.Fatal("pekerja dengan tanggal berhenti harus nonaktif")
	}
	if hasil.TanggalBerhenti == nil ||
		!samaTanggal(*hasil.TanggalBerhenti, tanggal(2026, time.November, 20)) {
		t.Fatalf("tanggal berhenti = %v", hasil.TanggalBerhenti)
	}
}

func TestHalamanRiwayatUpahDanSimpan(t *testing.T) {
	mux, repo := newTestMux(t)
	pekerja := seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, time.November, 2), Aktif: true,
	})
	if err := repo.SimpanUpah(t.Context(), &domain.Upah{
		PekerjaID: pekerja.ID, Nominal: 100000,
		BerlakuMulai: tanggal(2026, time.November, 2), Alasan: "Upah awal",
	}); err != nil {
		t.Fatalf("gagal menyiapkan upah: %v", err)
	}

	halaman := minta(mux, "/pekerja/1/upah")
	if halaman.Code != http.StatusOK {
		t.Fatalf("status GET = %d, ingin 200", halaman.Code)
	}
	body := halaman.Body.String()
	for _, teks := range []string{"Budi", "P-001", "Rp 100.000", `action="/pekerja/1/upah"`} {
		if !strings.Contains(body, teks) {
			t.Fatalf("halaman riwayat tidak memuat %q", teks)
		}
	}

	simpan := kirimForm(mux, "/pekerja/1/upah", url.Values{
		"nominal":       {"120000"},
		"berlaku_mulai": {"2026-11-05"},
		"alasan":        {"Kenaikan upah"},
	})
	if simpan.Code != http.StatusSeeOther {
		t.Fatalf("status POST = %d, ingin 303; body: %s", simpan.Code, simpan.Body.String())
	}
	if len(repo.upah[pekerja.ID]) != 2 {
		t.Fatalf("jumlah riwayat = %d, ingin 2", len(repo.upah[pekerja.ID]))
	}
}

func TestFormMenampilkanGalatValidasi(t *testing.T) {
	mux, repo := newTestMux(t)
	seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, time.November, 2), Aktif: true,
	})

	hasil := kirimForm(mux, "/pekerja/1/upah", url.Values{
		"nominal":       {"0"},
		"berlaku_mulai": {"2026-11-05"},
	})
	if hasil.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", hasil.Code)
	}
	if !strings.Contains(hasil.Body.String(), "Nominal upah harus lebih dari 0.") {
		t.Fatalf("galat validasi tidak tampil; body: %s", hasil.Body.String())
	}
}

func TestHalamanRiwayatUpahPekerjaTidakAda(t *testing.T) {
	mux, _ := newTestMux(t)
	hasil := minta(mux, "/pekerja/99/upah")
	if hasil.Code != http.StatusNotFound {
		t.Fatalf("status = %d, ingin 404", hasil.Code)
	}
}

func TestNonaktifkanPekerjaDariDaftar(t *testing.T) {
	mux, repo := newTestMux(t)
	pekerja := seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, time.November, 2), Aktif: true,
	})

	hasil := kirimForm(mux, "/pekerja/1/nonaktif", url.Values{
		"tanggal_berhenti": {"2026-11-20"},
	})
	if hasil.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, ingin 303; body: %s", hasil.Code, hasil.Body.String())
	}
	terbaru := repo.pekerja[pekerja.ID]
	if terbaru.Aktif {
		t.Fatal("pekerja harus nonaktif")
	}
	if terbaru.TanggalBerhenti == nil ||
		!samaTanggal(*terbaru.TanggalBerhenti, tanggal(2026, time.November, 20)) {
		t.Fatalf("tanggal berhenti = %v", terbaru.TanggalBerhenti)
	}
}

func TestNonaktifkanTanpaTanggalMemakaiHariIni(t *testing.T) {
	mux, repo := newTestMux(t)
	pekerja := seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2020, time.January, 1), Aktif: true,
	})

	hasil := kirimForm(mux, "/pekerja/1/nonaktif", url.Values{})
	if hasil.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, ingin 303; body: %s", hasil.Code, hasil.Body.String())
	}
	terbaru := repo.pekerja[pekerja.ID]
	if terbaru.Aktif || terbaru.TanggalBerhenti == nil {
		t.Fatalf("pekerja harus nonaktif dengan tanggal berhenti, dapat %+v", terbaru)
	}
	if terbaru.TanggalBerhenti.Before(tanggal(2020, time.January, 1)) {
		t.Fatalf("tanggal berhenti = %v, tidak boleh sebelum mulai", terbaru.TanggalBerhenti)
	}
}
