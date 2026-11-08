// Pemilik: Mahdi | Modul 1 | Aturan: 8, 11, 16
package pekerja

import (
	"context"
	"errors"
	"sort"
	"strings"
	"testing"
	"time"

	"penggajian-konveksi/internal/domain"
)

// fakeRepository adalah implementasi in-memory Repositori untuk tes tanpa MySQL.
type fakeRepository struct {
	pekerja       map[int64]domain.Pekerja
	upah          map[int64][]domain.Upah
	kode          map[string]int64
	nextPekerjaID int64
	nextUpahID    int64
	lastFilter    Filter
	err           error
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		pekerja:       map[int64]domain.Pekerja{},
		upah:          map[int64][]domain.Upah{},
		kode:          map[string]int64{},
		nextPekerjaID: 1,
		nextUpahID:    1,
	}
}

func (f *fakeRepository) SimpanPekerja(_ context.Context, pekerja *domain.Pekerja) error {
	if f.err != nil {
		return f.err
	}
	if id, ada := f.kode[pekerja.Kode]; ada && id != pekerja.ID {
		return errors.New("Kode pekerja sudah dipakai.")
	}
	if pekerja.ID == 0 {
		pekerja.ID = f.nextPekerjaID
		f.nextPekerjaID++
	}
	f.pekerja[pekerja.ID] = *pekerja
	f.kode[pekerja.Kode] = pekerja.ID
	return nil
}

func (f *fakeRepository) SimpanPekerjaDanUpah(_ context.Context, pekerja *domain.Pekerja, upah *domain.Upah) error {
	if err := f.SimpanPekerja(context.Background(), pekerja); err != nil {
		return err
	}
	upah.PekerjaID = pekerja.ID
	return f.SimpanUpah(context.Background(), upah)
}

func (f *fakeRepository) UbahPekerja(_ context.Context, pekerja domain.Pekerja) error {
	if f.err != nil {
		return f.err
	}
	lama, ada := f.pekerja[pekerja.ID]
	if !ada {
		return domain.ErrTidakDitemukan
	}
	if id, dipakai := f.kode[pekerja.Kode]; dipakai && id != pekerja.ID {
		return errors.New("Kode pekerja sudah dipakai.")
	}
	delete(f.kode, lama.Kode)
	f.pekerja[pekerja.ID] = pekerja
	f.kode[pekerja.Kode] = pekerja.ID
	return nil
}

func (f *fakeRepository) NonaktifkanPekerja(_ context.Context, id int64, tanggal time.Time) error {
	if f.err != nil {
		return f.err
	}
	pekerja, ada := f.pekerja[id]
	if !ada {
		return domain.ErrTidakDitemukan
	}
	t := tanggal
	pekerja.TanggalBerhenti = &t
	pekerja.Aktif = false
	f.pekerja[id] = pekerja
	return nil
}

func (f *fakeRepository) DaftarPekerja(_ context.Context, filter Filter) ([]domain.Pekerja, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.lastFilter = filter
	var hasil []domain.Pekerja
	cari := strings.ToLower(strings.TrimSpace(filter.Cari))
	for _, pekerja := range f.pekerja {
		if cari != "" && !strings.Contains(strings.ToLower(pekerja.Kode+" "+pekerja.Nama), cari) {
			continue
		}
		if filter.Jenis != "" && pekerja.Jenis != filter.Jenis {
			continue
		}
		if filter.Aktif != nil && pekerja.Aktif != *filter.Aktif {
			continue
		}
		hasil = append(hasil, pekerja)
	}
	sort.Slice(hasil, func(i, j int) bool { return hasil[i].Kode < hasil[j].Kode })
	return hasil, nil
}

func (f *fakeRepository) CariPekerja(_ context.Context, id int64) (domain.Pekerja, error) {
	if f.err != nil {
		return domain.Pekerja{}, f.err
	}
	pekerja, ada := f.pekerja[id]
	if !ada {
		return domain.Pekerja{}, domain.ErrTidakDitemukan
	}
	return pekerja, nil
}

func (f *fakeRepository) KodeSudahDipakai(_ context.Context, kode string, kecualiID int64) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	id, ada := f.kode[kode]
	return ada && id != kecualiID, nil
}

func (f *fakeRepository) SimpanUpah(_ context.Context, upah *domain.Upah) error {
	if f.err != nil {
		return f.err
	}
	for _, lama := range f.upah[upah.PekerjaID] {
		if samaTanggal(lama.BerlakuMulai, upah.BerlakuMulai) {
			return errors.New("Sudah ada upah dengan tanggal berlaku yang sama.")
		}
	}
	if upah.ID == 0 {
		upah.ID = f.nextUpahID
		f.nextUpahID++
	}
	f.upah[upah.PekerjaID] = append(f.upah[upah.PekerjaID], *upah)
	return nil
}

func (f *fakeRepository) DaftarUpah(_ context.Context, pekerjaID int64) ([]domain.Upah, error) {
	if f.err != nil {
		return nil, f.err
	}
	hasil := append([]domain.Upah(nil), f.upah[pekerjaID]...)
	sort.Slice(hasil, func(i, j int) bool {
		if samaTanggal(hasil[i].BerlakuMulai, hasil[j].BerlakuMulai) {
			return hasil[i].ID > hasil[j].ID
		}
		return hasil[i].BerlakuMulai.After(hasil[j].BerlakuMulai)
	})
	return hasil, nil
}

func (f *fakeRepository) UpahBerlaku(_ context.Context, pekerjaID int64, tanggal time.Time) (domain.Upah, error) {
	if f.err != nil {
		return domain.Upah{}, f.err
	}
	var terpilih domain.Upah
	ada := false
	for _, upah := range f.upah[pekerjaID] {
		if tanggalSaja(upah.BerlakuMulai).After(tanggalSaja(tanggal)) {
			continue
		}
		if !ada || upah.BerlakuMulai.After(terpilih.BerlakuMulai) ||
			(samaTanggal(upah.BerlakuMulai, terpilih.BerlakuMulai) && upah.ID > terpilih.ID) {
			terpilih = upah
			ada = true
		}
	}
	if !ada {
		return domain.Upah{}, domain.ErrTidakDitemukan
	}
	return terpilih, nil
}

func (f *fakeRepository) UpahPadaTanggal(_ context.Context, pekerjaID int64, tanggal time.Time) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	for _, upah := range f.upah[pekerjaID] {
		if samaTanggal(upah.BerlakuMulai, tanggal) {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeRepository) PekerjaAktif(_ context.Context, jenis domain.JenisPekerja, mulai, selesai time.Time) ([]domain.Pekerja, error) {
	if f.err != nil {
		return nil, f.err
	}
	mulai = tanggalSaja(mulai)
	selesai = tanggalSaja(selesai)
	var hasil []domain.Pekerja
	for _, pekerja := range f.pekerja {
		if jenis != "" && pekerja.Jenis != jenis {
			continue
		}
		if tanggalSaja(pekerja.TanggalMulai).After(selesai) {
			continue
		}
		if pekerja.TanggalBerhenti != nil && tanggalSaja(*pekerja.TanggalBerhenti).Before(mulai) {
			continue
		}
		hasil = append(hasil, pekerja)
	}
	sort.Slice(hasil, func(i, j int) bool { return hasil[i].Nama < hasil[j].Nama })
	return hasil, nil
}

func seedPekerja(t *testing.T, repo *fakeRepository, pekerja domain.Pekerja) domain.Pekerja {
	t.Helper()
	if err := repo.SimpanPekerja(context.Background(), &pekerja); err != nil {
		t.Fatalf("gagal menyiapkan pekerja: %v", err)
	}
	return pekerja
}

func tanggal(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

func TestTambahPekerjaValidasi(t *testing.T) {
	svc := NewService(newFakeRepository())
	mulai := tanggal(2026, time.November, 2)
	kasus := []struct {
		nama    string
		pekerja domain.Pekerja
		upah    domain.Rupiah
	}{
		{"kode kosong", domain.Pekerja{Nama: "Budi", Jenis: domain.PekerjaHarian, TanggalMulai: mulai}, 100000},
		{"kode terlalu panjang", domain.Pekerja{Kode: strings.Repeat("P", 21), Nama: "Budi", Jenis: domain.PekerjaHarian, TanggalMulai: mulai}, 100000},
		{"nama kosong", domain.Pekerja{Kode: "P-001", Jenis: domain.PekerjaHarian, TanggalMulai: mulai}, 100000},
		{"nama terlalu panjang", domain.Pekerja{Kode: "P-001", Nama: strings.Repeat("A", 101), Jenis: domain.PekerjaHarian, TanggalMulai: mulai}, 100000},
		{"jenis tidak dikenal", domain.Pekerja{Kode: "P-001", Nama: "Budi", Jenis: domain.JenisPekerja("MINGGUAN"), TanggalMulai: mulai}, 100000},
		{"tanggal mulai kosong", domain.Pekerja{Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian}, 100000},
		{"upah nol", domain.Pekerja{Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian, TanggalMulai: mulai}, 0},
		{"upah negatif", domain.Pekerja{Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian, TanggalMulai: mulai}, -1},
	}

	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			if _, err := svc.TambahPekerja(context.Background(), k.pekerja, k.upah); err == nil {
				t.Fatal("data pekerja tidak valid seharusnya ditolak")
			}
		})
	}
}

func TestTambahPekerjaKodeSudahDipakai(t *testing.T) {
	repo := newFakeRepository()
	seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, time.November, 2), Aktif: true,
	})
	svc := NewService(repo)

	_, err := svc.TambahPekerja(context.Background(), domain.Pekerja{
		Kode: "P-001", Nama: "Siti", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, time.November, 3),
	}, 100000)
	if err == nil || err.Error() != "Kode pekerja sudah dipakai." {
		t.Fatalf("galat = %v, ingin \"Kode pekerja sudah dipakai.\"", err)
	}
}

func TestTambahPekerjaMenyimpanUpahAwal(t *testing.T) {
	repo := newFakeRepository()
	svc := NewService(repo)
	mulai := tanggal(2026, time.November, 2)

	id, err := svc.TambahPekerja(context.Background(), domain.Pekerja{
		Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian, TanggalMulai: mulai,
	}, 100000)
	if err != nil {
		t.Fatalf("TambahPekerja gagal: %v", err)
	}
	if id == 0 {
		t.Fatal("ID pekerja baru tidak boleh nol")
	}
	if !repo.pekerja[id].Aktif {
		t.Fatal("pekerja baru harus aktif")
	}
	upah := repo.upah[id]
	if len(upah) != 1 {
		t.Fatalf("jumlah upah awal = %d, ingin 1", len(upah))
	}
	if upah[0].Nominal != 100000 {
		t.Fatalf("nominal upah awal = %d, ingin 100000", upah[0].Nominal)
	}
	if !samaTanggal(upah[0].BerlakuMulai, mulai) {
		t.Fatalf("tanggal upah awal = %v, ingin %v", upah[0].BerlakuMulai, mulai)
	}
	if upah[0].Alasan == "" {
		t.Fatal("alasan upah awal tidak boleh kosong")
	}
}

func TestUbahPekerjaTolakPerubahanJenis(t *testing.T) {
	repo := newFakeRepository()
	lama := seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, time.November, 2), Aktif: true,
	})
	svc := NewService(repo)

	lama.Jenis = domain.PekerjaBulanan
	err := svc.UbahPekerja(context.Background(), lama)
	if err == nil || !strings.Contains(err.Error(), "Jenis pekerja tidak dapat diubah") {
		t.Fatalf("galat = %v, ingin penolakan perubahan jenis", err)
	}
}

func TestUbahTanggalBerhentiSebelumMulai(t *testing.T) {
	repo := newFakeRepository()
	lama := seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, time.November, 10), Aktif: true,
	})
	berhenti := tanggal(2026, time.November, 1)
	lama.TanggalBerhenti = &berhenti
	svc := NewService(repo)

	err := svc.UbahPekerja(context.Background(), lama)
	if err == nil || err.Error() != "Tanggal berhenti tidak boleh sebelum tanggal mulai." {
		t.Fatalf("galat = %v, ingin \"Tanggal berhenti tidak boleh sebelum tanggal mulai.\"", err)
	}
}

func TestUbahPekerjaMempertahankanJenis(t *testing.T) {
	repo := newFakeRepository()
	lama := seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, time.November, 2), Aktif: true,
	})
	berhenti := tanggal(2026, time.November, 20)
	lama.Nama = "Budi Santoso"
	lama.TanggalBerhenti = &berhenti
	svc := NewService(repo)

	if err := svc.UbahPekerja(context.Background(), lama); err != nil {
		t.Fatalf("UbahPekerja gagal: %v", err)
	}
	hasil := repo.pekerja[lama.ID]
	if hasil.Nama != "Budi Santoso" {
		t.Fatalf("nama = %q, ingin %q", hasil.Nama, "Budi Santoso")
	}
	if hasil.Jenis != domain.PekerjaHarian {
		t.Fatalf("jenis berubah menjadi %q", hasil.Jenis)
	}
	if hasil.Aktif {
		t.Fatal("pekerja dengan tanggal berhenti harus nonaktif")
	}
	if hasil.TanggalBerhenti == nil || !samaTanggal(*hasil.TanggalBerhenti, berhenti) {
		t.Fatalf("tanggal berhenti = %v, ingin %v", hasil.TanggalBerhenti, berhenti)
	}
}

func TestNonaktifkanPekerja(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		repo := newFakeRepository()
		lama := seedPekerja(t, repo, domain.Pekerja{
			Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian,
			TanggalMulai: tanggal(2026, time.November, 2), Aktif: true,
		})
		svc := NewService(repo)
		berhenti := tanggal(2026, time.November, 20)

		if err := svc.NonaktifkanPekerja(context.Background(), lama.ID, berhenti); err != nil {
			t.Fatalf("NonaktifkanPekerja gagal: %v", err)
		}
		hasil := repo.pekerja[lama.ID]
		if hasil.Aktif {
			t.Fatal("pekerja harus nonaktif")
		}
		if hasil.TanggalBerhenti == nil || !samaTanggal(*hasil.TanggalBerhenti, berhenti) {
			t.Fatalf("tanggal berhenti = %v, ingin %v", hasil.TanggalBerhenti, berhenti)
		}
	})

	t.Run("sebelum tanggal mulai", func(t *testing.T) {
		repo := newFakeRepository()
		lama := seedPekerja(t, repo, domain.Pekerja{
			Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian,
			TanggalMulai: tanggal(2026, time.November, 10), Aktif: true,
		})
		svc := NewService(repo)

		err := svc.NonaktifkanPekerja(context.Background(), lama.ID, tanggal(2026, time.November, 1))
		if err == nil || err.Error() != "Tanggal berhenti tidak boleh sebelum tanggal mulai." {
			t.Fatalf("galat = %v, ingin penolakan tanggal berhenti", err)
		}
	})
}

func TestDaftarPekerjaMeneruskanFilter(t *testing.T) {
	repo := newFakeRepository()
	svc := NewService(repo)
	aktif := true
	filter := Filter{Cari: "Budi", Jenis: domain.PekerjaHarian, Aktif: &aktif}

	if _, err := svc.DaftarPekerja(context.Background(), filter); err != nil {
		t.Fatalf("DaftarPekerja gagal: %v", err)
	}
	if repo.lastFilter.Cari != filter.Cari || repo.lastFilter.Jenis != filter.Jenis {
		t.Fatalf("filter diteruskan = %+v, ingin %+v", repo.lastFilter, filter)
	}
	if repo.lastFilter.Aktif == nil || *repo.lastFilter.Aktif != *filter.Aktif {
		t.Fatalf("filter aktif diteruskan = %v, ingin %v", repo.lastFilter.Aktif, *filter.Aktif)
	}
}

func TestCariPekerjaTidakDitemukan(t *testing.T) {
	svc := NewService(newFakeRepository())
	_, err := svc.CariPekerja(context.Background(), 99)
	if !errors.Is(err, domain.ErrTidakDitemukan) {
		t.Fatalf("galat = %v, ingin domain.ErrTidakDitemukan", err)
	}
}

func TestRiwayatUpahTanggalGanda(t *testing.T) {
	repo := newFakeRepository()
	pekerja := seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, time.November, 2), Aktif: true,
	})
	if err := repo.SimpanUpah(context.Background(), &domain.Upah{
		PekerjaID: pekerja.ID, Nominal: 100000, BerlakuMulai: tanggal(2026, time.November, 2),
	}); err != nil {
		t.Fatalf("gagal menyiapkan upah: %v", err)
	}
	svc := NewService(repo)

	err := svc.TambahRiwayatUpah(context.Background(), domain.Upah{
		PekerjaID: pekerja.ID, Nominal: 120000, BerlakuMulai: tanggal(2026, time.November, 2),
	})
	if err == nil || err.Error() != "Sudah ada upah dengan tanggal berlaku yang sama." {
		t.Fatalf("galat = %v, ingin \"Sudah ada upah dengan tanggal berlaku yang sama.\"", err)
	}
}

func TestTambahRiwayatUpahSebelumMulai(t *testing.T) {
	repo := newFakeRepository()
	pekerja := seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, time.November, 10), Aktif: true,
	})
	svc := NewService(repo)

	err := svc.TambahRiwayatUpah(context.Background(), domain.Upah{
		PekerjaID: pekerja.ID, Nominal: 120000, BerlakuMulai: tanggal(2026, time.November, 1),
	})
	if err == nil || err.Error() != "Tanggal berlaku upah tidak boleh sebelum pekerja mulai bekerja." {
		t.Fatalf("galat = %v, ingin penolakan tanggal sebelum mulai bekerja", err)
	}
}

func TestTambahRiwayatUpahValid(t *testing.T) {
	repo := newFakeRepository()
	pekerja := seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, time.November, 2), Aktif: true,
	})
	svc := NewService(repo)

	err := svc.TambahRiwayatUpah(context.Background(), domain.Upah{
		PekerjaID: pekerja.ID, Nominal: 120000,
		BerlakuMulai: tanggal(2026, time.November, 5), Alasan: "  Kenaikan upah  ",
	})
	if err != nil {
		t.Fatalf("TambahRiwayatUpah gagal: %v", err)
	}
	riwayat := repo.upah[pekerja.ID]
	if len(riwayat) != 1 {
		t.Fatalf("jumlah riwayat = %d, ingin 1", len(riwayat))
	}
	if riwayat[0].Nominal != 120000 {
		t.Fatalf("nominal = %d, ingin 120000", riwayat[0].Nominal)
	}
	if riwayat[0].Alasan != "Kenaikan upah" {
		t.Fatalf("alasan = %q, ingin %q", riwayat[0].Alasan, "Kenaikan upah")
	}
}

func TestTambahRiwayatUpahNominalTidakValid(t *testing.T) {
	repo := newFakeRepository()
	pekerja := seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, time.November, 2), Aktif: true,
	})
	svc := NewService(repo)

	err := svc.TambahRiwayatUpah(context.Background(), domain.Upah{
		PekerjaID: pekerja.ID, Nominal: 0, BerlakuMulai: tanggal(2026, time.November, 5),
	})
	if err == nil || err.Error() != "Nominal upah harus lebih dari 0." {
		t.Fatalf("galat = %v, ingin penolakan nominal nol", err)
	}
}

func TestTambahRiwayatUpahPekerjaTidakAda(t *testing.T) {
	svc := NewService(newFakeRepository())
	err := svc.TambahRiwayatUpah(context.Background(), domain.Upah{
		PekerjaID: 99, Nominal: 120000, BerlakuMulai: tanggal(2026, time.November, 5),
	})
	if !errors.Is(err, domain.ErrTidakDitemukan) {
		t.Fatalf("galat = %v, ingin domain.ErrTidakDitemukan", err)
	}
}

func TestCariUpahBerlakuMemilihTerbaru(t *testing.T) {
	repo := newFakeRepository()
	pekerja := seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, time.November, 1), Aktif: true,
	})
	for _, upah := range []domain.Upah{
		{PekerjaID: pekerja.ID, Nominal: 100000, BerlakuMulai: tanggal(2026, time.November, 1)},
		{PekerjaID: pekerja.ID, Nominal: 120000, BerlakuMulai: tanggal(2026, time.November, 5)},
	} {
		if err := repo.SimpanUpah(context.Background(), &upah); err != nil {
			t.Fatalf("gagal menyiapkan upah: %v", err)
		}
	}
	svc := NewService(repo)

	kasus := []struct {
		tanggal time.Time
		ingin   domain.Rupiah
	}{
		{tanggal(2026, time.November, 1), 100000},
		{tanggal(2026, time.November, 4), 100000},
		{tanggal(2026, time.November, 5), 120000},
		{tanggal(2026, time.November, 6), 120000},
	}
	for _, k := range kasus {
		nominal, err := svc.CariUpahBerlaku(context.Background(), pekerja.ID, k.tanggal)
		if err != nil {
			t.Fatalf("CariUpahBerlaku(%v) gagal: %v", k.tanggal, err)
		}
		if nominal != k.ingin {
			t.Fatalf("CariUpahBerlaku(%v) = %d, ingin %d", k.tanggal, nominal, k.ingin)
		}
	}
}

func TestCariUpahBerlakuSebelumRiwayatPertama(t *testing.T) {
	repo := newFakeRepository()
	pekerja := seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, time.November, 1), Aktif: true,
	})
	if err := repo.SimpanUpah(context.Background(), &domain.Upah{
		PekerjaID: pekerja.ID, Nominal: 100000, BerlakuMulai: tanggal(2026, time.November, 2),
	}); err != nil {
		t.Fatalf("gagal menyiapkan upah: %v", err)
	}
	svc := NewService(repo)

	_, err := svc.CariUpahBerlaku(context.Background(), pekerja.ID, tanggal(2026, time.November, 1))
	if !errors.Is(err, domain.ErrTidakDitemukan) {
		t.Fatalf("galat = %v, ingin domain.ErrTidakDitemukan", err)
	}
}

func TestCariUpahBerlakuMengabaikanJam(t *testing.T) {
	repo := newFakeRepository()
	pekerja := seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, time.November, 1), Aktif: true,
	})
	if err := repo.SimpanUpah(context.Background(), &domain.Upah{
		PekerjaID: pekerja.ID, Nominal: 100000,
		BerlakuMulai: time.Date(2026, time.November, 2, 12, 30, 0, 0, time.Local),
	}); err != nil {
		t.Fatalf("gagal menyiapkan upah: %v", err)
	}
	svc := NewService(repo)

	nominal, err := svc.CariUpahBerlaku(context.Background(), pekerja.ID,
		time.Date(2026, time.November, 2, 23, 59, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("CariUpahBerlaku gagal: %v", err)
	}
	if nominal != 100000 {
		t.Fatalf("nominal = %d, ingin 100000", nominal)
	}
}

func TestCariUpahBerlakuPekerjaTanpaRiwayat(t *testing.T) {
	repo := newFakeRepository()
	pekerja := seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, time.November, 1), Aktif: true,
	})
	svc := NewService(repo)

	_, err := svc.CariUpahBerlaku(context.Background(), pekerja.ID, tanggal(2026, time.November, 2))
	if !errors.Is(err, domain.ErrTidakDitemukan) {
		t.Fatalf("galat = %v, ingin domain.ErrTidakDitemukan", err)
	}
}

func TestDaftarPekerjaAktifMenghormatiRentangDanJenis(t *testing.T) {
	repo := newFakeRepository()
	berhentiB := tanggal(2026, time.November, 10)
	berhentiC := tanggal(2026, time.September, 30)
	seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-001", Nama: "A", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, time.November, 1), Aktif: true,
	})
	seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-002", Nama: "B", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, time.November, 1), TanggalBerhenti: &berhentiB, Aktif: false,
	})
	seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-003", Nama: "C", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, time.September, 1), TanggalBerhenti: &berhentiC, Aktif: false,
	})
	seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-004", Nama: "D", Jenis: domain.PekerjaBulanan,
		TanggalMulai: tanggal(2026, time.November, 1), Aktif: true,
	})
	svc := NewService(repo)

	hasil, err := svc.DaftarPekerjaAktif(context.Background(), domain.PekerjaHarian,
		tanggal(2026, time.November, 10), tanggal(2026, time.November, 16))
	if err != nil {
		t.Fatalf("DaftarPekerjaAktif gagal: %v", err)
	}
	if len(hasil) != 2 || hasil[0].Kode != "P-001" || hasil[1].Kode != "P-002" {
		t.Fatalf("hasil = %+v, ingin P-001 dan P-002", hasil)
	}

	semua, err := svc.DaftarPekerjaAktif(context.Background(), "",
		tanggal(2026, time.November, 10), tanggal(2026, time.November, 16))
	if err != nil {
		t.Fatalf("DaftarPekerjaAktif semua jenis gagal: %v", err)
	}
	if len(semua) != 3 {
		t.Fatalf("jumlah semua jenis = %d, ingin 3", len(semua))
	}
}

func TestDaftarPekerjaAktifRentangTidakValid(t *testing.T) {
	svc := NewService(newFakeRepository())
	_, err := svc.DaftarPekerjaAktif(context.Background(), domain.PekerjaHarian,
		tanggal(2026, time.November, 16), tanggal(2026, time.November, 10))
	if err == nil || err.Error() != "Rentang tanggal periode tidak valid." {
		t.Fatalf("galat = %v, ingin penolakan rentang terbalik", err)
	}
}

func TestDaftarRiwayatUpahTerurut(t *testing.T) {
	repo := newFakeRepository()
	pekerja := seedPekerja(t, repo, domain.Pekerja{
		Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, time.November, 1), Aktif: true,
	})
	for _, upah := range []domain.Upah{
		{PekerjaID: pekerja.ID, Nominal: 100000, BerlakuMulai: tanggal(2026, time.November, 2)},
		{PekerjaID: pekerja.ID, Nominal: 120000, BerlakuMulai: tanggal(2026, time.November, 5)},
	} {
		if err := repo.SimpanUpah(context.Background(), &upah); err != nil {
			t.Fatalf("gagal menyiapkan upah: %v", err)
		}
	}
	svc := NewService(repo)

	riwayat, err := svc.DaftarRiwayatUpah(context.Background(), pekerja.ID)
	if err != nil {
		t.Fatalf("DaftarRiwayatUpah gagal: %v", err)
	}
	if len(riwayat) != 2 || riwayat[0].Nominal != 120000 {
		t.Fatalf("riwayat = %+v, ingin upah terbaru lebih dulu", riwayat)
	}
}
