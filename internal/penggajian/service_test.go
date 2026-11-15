// Pemilik: Farah | Modul 4 | Aturan: 5, 6, 7, 9, 10, 12, 17
package penggajian

import (
	"context"
	"errors"
	"testing"
	"time"

	"penggajian-konveksi/internal/domain"
)

type fakeRepo struct {
	periode     domain.PeriodeGaji
	sebelumnya  *domain.PeriodeGaji
	workers     map[domain.JenisPekerja][]domain.Pekerja
	absensi     []domain.AbsensiLengkap
	upah        map[int64][]domain.Upah
	kasbon      map[int64][]domain.Kasbon
	koreksi     map[int64][]domain.KoreksiGaji
	diterapkan  map[int64][]domain.KoreksiGaji
	slips       map[int64]domain.SlipGaji
	rincian     map[int64][]RincianHarian
	status      domain.StatusPeriode
	hasilHitung []HasilHitung
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		workers:    make(map[domain.JenisPekerja][]domain.Pekerja),
		upah:       make(map[int64][]domain.Upah),
		kasbon:     make(map[int64][]domain.Kasbon),
		koreksi:    make(map[int64][]domain.KoreksiGaji),
		diterapkan: make(map[int64][]domain.KoreksiGaji),
		slips:      make(map[int64]domain.SlipGaji),
		rincian:    make(map[int64][]RincianHarian),
	}
}

func (f *fakeRepo) SimpanSlipDanRincian(_ context.Context, slip domain.SlipGaji, rincian []domain.SlipGajiDetail, potongan []domain.PotonganKasbon) error {
	f.slips[slip.PekerjaID] = slip
	hasil := make([]RincianHarian, 0, len(rincian))
	for _, d := range rincian {
		hasil = append(hasil, RincianHarian{SlipGajiDetail: d})
	}
	f.rincian[slip.PekerjaID] = hasil
	return nil
}

func (f *fakeRepo) SimpanHasilPeriode(_ context.Context, periodeID int64, hasil []HasilHitung) error {
	f.slips = make(map[int64]domain.SlipGaji)
	f.rincian = make(map[int64][]RincianHarian)
	f.hasilHitung = hasil
	for _, h := range hasil {
		slip := h.Slip
		slip.PeriodeID = periodeID
		f.slips[slip.PekerjaID] = slip
		rincian := make([]RincianHarian, 0, len(h.Rincian))
		for _, d := range h.Rincian {
			rincian = append(rincian, RincianHarian{SlipGajiDetail: d})
		}
		f.rincian[slip.PekerjaID] = rincian
	}
	return nil
}

func (f *fakeRepo) AmbilSlip(_ context.Context, periodeID, pekerjaID int64) (domain.SlipGaji, []domain.SlipGajiDetail, error) {
	slip, rincian, err := f.AmbilRincian(context.Background(), periodeID, pekerjaID)
	if err != nil {
		return domain.SlipGaji{}, nil, err
	}
	detail := make([]domain.SlipGajiDetail, 0, len(rincian))
	for _, d := range rincian {
		detail = append(detail, d.SlipGajiDetail)
	}
	return slip, detail, nil
}

func (f *fakeRepo) AmbilRincian(_ context.Context, periodeID, pekerjaID int64) (domain.SlipGaji, []RincianHarian, error) {
	slip, ada := f.slips[pekerjaID]
	if !ada || slip.PeriodeID != periodeID {
		return domain.SlipGaji{}, nil, domain.ErrTidakDitemukan
	}
	return slip, f.rincian[pekerjaID], nil
}

func (f *fakeRepo) DaftarSlip(_ context.Context, periodeID int64) ([]domain.SlipGaji, error) {
	var hasil []domain.SlipGaji
	for _, slip := range f.slips {
		if slip.PeriodeID == periodeID {
			hasil = append(hasil, slip)
		}
	}
	return hasil, nil
}

func (f *fakeRepo) DaftarPeriode(_ context.Context) ([]domain.PeriodeGaji, error) {
	if f.periode.ID == 0 {
		return nil, nil
	}
	return []domain.PeriodeGaji{f.periode}, nil
}

func (f *fakeRepo) AmbilPeriode(_ context.Context, id int64) (domain.PeriodeGaji, error) {
	if f.periode.ID != id {
		return domain.PeriodeGaji{}, domain.ErrTidakDitemukan
	}
	return f.periode, nil
}

func (f *fakeRepo) PeriodeSebelumnya(_ context.Context, _ domain.PeriodeGaji) (domain.PeriodeGaji, bool, error) {
	if f.sebelumnya == nil {
		return domain.PeriodeGaji{}, false, nil
	}
	return *f.sebelumnya, true, nil
}

func (f *fakeRepo) PekerjaAktif(_ context.Context, jenis domain.JenisPekerja, _, _ time.Time) ([]domain.Pekerja, error) {
	return f.workers[jenis], nil
}

func (f *fakeRepo) AbsensiPeriode(_ context.Context, _ int64) ([]domain.AbsensiLengkap, error) {
	return f.absensi, nil
}

func (f *fakeRepo) AbsensiBelumLengkap(_ context.Context, _ int64) (int, error) {
	jumlah := 0
	for _, a := range f.absensi {
		if a.Absensi.JamKeluar == nil {
			jumlah++
		}
	}
	return jumlah, nil
}

func (f *fakeRepo) UpahSampai(_ context.Context, pekerjaID int64, _ time.Time) ([]domain.Upah, error) {
	return f.upah[pekerjaID], nil
}

func (f *fakeRepo) KasbonSampai(_ context.Context, pekerjaID int64, _ time.Time) ([]domain.Kasbon, error) {
	return f.kasbon[pekerjaID], nil
}

func (f *fakeRepo) SisaKasbon(_ context.Context, pekerjaID int64) (domain.Rupiah, error) {
	var total domain.Rupiah
	for _, k := range f.kasbon[pekerjaID] {
		total += k.Sisa
	}
	return total, nil
}

func (f *fakeRepo) KoreksiUntukPeriode(_ context.Context, _, pekerjaID int64) ([]domain.KoreksiGaji, error) {
	hasil := append([]domain.KoreksiGaji{}, f.koreksi[pekerjaID]...)
	hasil = append(hasil, f.diterapkan[pekerjaID]...)
	return hasil, nil
}

func (f *fakeRepo) KoreksiDiterapkan(_ context.Context, _, pekerjaID int64) ([]domain.KoreksiGaji, error) {
	return f.diterapkan[pekerjaID], nil
}

func (f *fakeRepo) TandaiKoreksiPeriode(_ context.Context, _ int64) error {
	for pekerjaID, daftar := range f.koreksi {
		f.diterapkan[pekerjaID] = append(f.diterapkan[pekerjaID], daftar...)
		delete(f.koreksi, pekerjaID)
	}
	return nil
}

func (f *fakeRepo) TanggalTerkunci(_ context.Context, _ time.Time) (bool, error) {
	return false, nil
}

func (f *fakeRepo) UbahStatusPeriode(_ context.Context, _ int64, status domain.StatusPeriode) error {
	f.status = status
	f.periode.Status = status
	return nil
}

func serviceUji(repo Repositori) *Service {
	return NewService(repo, nil, nil, nil, nil, nil, nil)
}

func absensiHarian(id, pekerjaID int64, pekerja domain.Pekerja, tanggal time.Time, masuk, keluar time.Time) domain.AbsensiLengkap {
	shift := domain.Shift{ID: 1, Nama: "Pagi", JamMulai: "07:00", JamSelesai: "15:00"}
	jadwal := domain.JadwalShift{
		ID: id, PekerjaID: pekerjaID, ShiftID: 1, Tanggal: tanggal,
		Shift: shift, Pekerja: pekerja,
	}
	jamKeluar := keluar
	return domain.AbsensiLengkap{
		Absensi: domain.Absensi{
			ID: id, JadwalShiftID: id, JamMasuk: masuk, JamKeluar: &jamKeluar,
		},
		Jadwal:  jadwal,
		Pekerja: pekerja,
	}
}

func absensiBulanan(id, pekerjaID int64, pekerja domain.Pekerja, tanggal time.Time, masuk, keluar time.Time) domain.AbsensiLengkap {
	shift := domain.Shift{ID: 2, Nama: "Siang", JamMulai: "15:00", JamSelesai: "23:00"}
	jadwal := domain.JadwalShift{
		ID: id, PekerjaID: pekerjaID, ShiftID: 2, Tanggal: tanggal,
		Shift: shift, Pekerja: pekerja,
	}
	jamKeluar := keluar
	return domain.AbsensiLengkap{
		Absensi: domain.Absensi{
			ID: id, JadwalShiftID: id, JamMasuk: masuk, JamKeluar: &jamKeluar,
		},
		Jadwal:  jadwal,
		Pekerja: pekerja,
	}
}

func TestHitungGajiPeriodeHarianContohA(t *testing.T) {
	repo := newFakeRepo()
	periode := domain.PeriodeGaji{
		ID: 1, Tipe: domain.PeriodeMingguan,
		TanggalMulai: tanggal(2026, 11, 2), TanggalSelesai: tanggal(2026, 11, 7),
		Status: domain.PeriodeTerbuka,
	}
	repo.periode = periode
	pekerja := domain.Pekerja{
		ID: 1, Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, 11, 1), Aktif: true,
	}
	repo.workers[domain.PekerjaHarian] = []domain.Pekerja{pekerja}
	repo.upah[pekerja.ID] = []domain.Upah{{Nominal: 100000, BerlakuMulai: tanggal(2026, 11, 1)}}
	repo.kasbon[pekerja.ID] = []domain.Kasbon{{ID: 1, PekerjaID: pekerja.ID, Sisa: 300000}}

	for i := 0; i < 6; i++ {
		tgl := tanggal(2026, 11, 2+i)
		masuk := time.Date(tgl.Year(), tgl.Month(), tgl.Day(), 7, 0, 0, 0, time.Local)
		keluar := time.Date(tgl.Year(), tgl.Month(), tgl.Day(), 15, 0, 0, 0, time.Local)
		// Nilai acuan kerangka memakai 150 menit lembur (60 menit x1,5 +
		// 90 menit x2) dan satu hari terlambat 20 menit.
		switch i {
		case 4:
			keluar = keluar.Add(150 * time.Minute)
		case 5:
			masuk = masuk.Add(20 * time.Minute)
		}
		repo.absensi = append(repo.absensi,
			absensiHarian(int64(i+1), pekerja.ID, pekerja, tgl, masuk, keluar))
	}

	if err := serviceUji(repo).HitungGajiPeriode(context.Background(), periode.ID); err != nil {
		t.Fatalf("hitung gaji: %v", err)
	}
	slip, ada := repo.slips[pekerja.ID]
	if !ada {
		t.Fatal("slip tidak tersimpan")
	}
	if slip.UpahPokok != 600000 {
		t.Fatalf("upah pokok = %d, ingin 600000", slip.UpahPokok)
	}
	if slip.TotalLembur != 56250 {
		t.Fatalf("total lembur = %d, ingin 56250", slip.TotalLembur)
	}
	if slip.PotonganTelat != 12500 {
		t.Fatalf("potongan telat = %d, ingin 12500", slip.PotonganTelat)
	}
	if slip.GajiBruto != 643750 {
		t.Fatalf("gaji bruto = %d, ingin 643750", slip.GajiBruto)
	}
	if slip.PotonganKasbon != 193125 {
		t.Fatalf("potongan kasbon = %d, ingin 193125", slip.PotonganKasbon)
	}
	if slip.GajiBersih != 450625 {
		t.Fatalf("gaji bersih = %d, ingin 450625", slip.GajiBersih)
	}
	if slip.HariKerjaDibayar != 6 || slip.HariKerjaPeriode != 0 {
		t.Fatalf("hari kerja = %d/%d, ingin 6/0", slip.HariKerjaDibayar, slip.HariKerjaPeriode)
	}
	if repo.status != domain.PeriodeDihitung {
		t.Fatalf("status periode = %s, ingin DIHITUNG", repo.status)
	}
	if len(repo.hasilHitung) != 1 || len(repo.hasilHitung[0].Rincian) != 6 {
		t.Fatalf("hasil hitung = %d slip, ingin 1 slip dengan 6 rincian", len(repo.hasilHitung))
	}
	if len(repo.hasilHitung[0].Potongan) != 1 ||
		repo.hasilHitung[0].Potongan[0].Jumlah != 193125 {
		t.Fatalf("potongan kasbon tersimpan = %+v", repo.hasilHitung[0].Potongan)
	}
}

func TestHitungGajiPeriodeBulananProporsionalMasukPertengahanBulan(t *testing.T) {
	repo := newFakeRepo()
	periode := domain.PeriodeGaji{
		ID: 2, Tipe: domain.PeriodeBulanan,
		TanggalMulai: tanggal(2026, 11, 1), TanggalSelesai: tanggal(2026, 11, 30),
		Status: domain.PeriodeTerbuka,
	}
	repo.periode = periode
	pekerja := domain.Pekerja{
		ID: 2, Kode: "P-002", Nama: "Siti", Jenis: domain.PekerjaBulanan,
		TanggalMulai: tanggal(2026, 11, 16), Aktif: true,
	}
	repo.workers[domain.PekerjaBulanan] = []domain.Pekerja{pekerja}
	repo.upah[pekerja.ID] = []domain.Upah{{Nominal: 3460000, BerlakuMulai: tanggal(2026, 11, 16)}}
	repo.kasbon[pekerja.ID] = []domain.Kasbon{{ID: 2, PekerjaID: pekerja.ID, Sisa: 1500000}}

	id := int64(1)
	for hari := 16; hari <= 30; hari++ {
		tgl := tanggal(2026, 11, hari)
		if tgl.Weekday() == time.Sunday {
			continue
		}
		masuk := time.Date(2026, 11, hari, 15, 0, 0, 0, time.Local)
		keluar := time.Date(2026, 11, hari, 23, 0, 0, 0, time.Local)
		switch hari {
		case 18:
			masuk = masuk.Add(20 * time.Minute)
		case 20:
			keluar = time.Date(2026, 11, 21, 2, 0, 0, 0, time.Local)
		}
		repo.absensi = append(repo.absensi,
			absensiBulanan(id, pekerja.ID, pekerja, tgl, masuk, keluar))
		id++
	}

	if err := serviceUji(repo).HitungGajiPeriode(context.Background(), periode.ID); err != nil {
		t.Fatalf("hitung gaji: %v", err)
	}
	slip := repo.slips[pekerja.ID]
	if slip.UpahPokok != 1799200 {
		t.Fatalf("upah pokok = %d, ingin 1799200", slip.UpahPokok)
	}
	if slip.TotalLembur != 110000 {
		t.Fatalf("total lembur = %d, ingin 110000", slip.TotalLembur)
	}
	if slip.PotonganTelat != 20000 {
		t.Fatalf("potongan telat = %d, ingin 20000", slip.PotonganTelat)
	}
	if slip.GajiBruto != 1889200 {
		t.Fatalf("gaji bruto = %d, ingin 1889200", slip.GajiBruto)
	}
	if slip.PotonganKasbon != 566760 {
		t.Fatalf("potongan kasbon = %d, ingin 566760", slip.PotonganKasbon)
	}
	if slip.GajiBersih != 1322440 {
		t.Fatalf("gaji bersih = %d, ingin 1322440", slip.GajiBersih)
	}
	if slip.HariKerjaDibayar != 13 || slip.HariKerjaPeriode != 25 {
		t.Fatalf("hari kerja = %d/%d, ingin 13/25", slip.HariKerjaDibayar, slip.HariKerjaPeriode)
	}
}

func TestHitungGajiPeriodeBulananContohC(t *testing.T) {
	repo := newFakeRepo()
	periode := domain.PeriodeGaji{
		ID: 9, Tipe: domain.PeriodeBulanan,
		TanggalMulai: tanggal(2026, 11, 1), TanggalSelesai: tanggal(2026, 11, 30),
		Status: domain.PeriodeTerbuka,
	}
	repo.periode = periode
	pekerja := domain.Pekerja{
		ID: 9, Kode: "P-009", Nama: "Tari", Jenis: domain.PekerjaBulanan,
		TanggalMulai: tanggal(2026, 1, 1), Aktif: true,
	}
	repo.workers[domain.PekerjaBulanan] = []domain.Pekerja{pekerja}
	repo.upah[pekerja.ID] = []domain.Upah{
		{Nominal: 3000000, BerlakuMulai: tanggal(2026, 1, 1)},
		{Nominal: 3460000, BerlakuMulai: tanggal(2026, 11, 16)},
	}

	if err := serviceUji(repo).HitungGajiPeriode(context.Background(), periode.ID); err != nil {
		t.Fatalf("hitung gaji: %v", err)
	}
	slip := repo.slips[pekerja.ID]
	// 3.000.000 x 12/25 = 1.440.000; 3.460.000 x 13/25 = 1.799.200.
	if slip.UpahPokok != 3239200 || slip.GajiBruto != 3239200 || slip.GajiBersih != 3239200 {
		t.Fatalf("slip contoh C = pokok %d bruto %d bersih %d",
			slip.UpahPokok, slip.GajiBruto, slip.GajiBersih)
	}
	if slip.HariKerjaDibayar != 25 || slip.HariKerjaPeriode != 25 {
		t.Fatalf("hari kerja = %d/%d, ingin 25/25",
			slip.HariKerjaDibayar, slip.HariKerjaPeriode)
	}
	if len(repo.hasilHitung) != 1 {
		t.Fatalf("jumlah hasil hitung = %d, ingin 1", len(repo.hasilHitung))
	}
	if len(repo.hasilHitung[0].Rincian) != 0 {
		t.Fatalf("rincian = %d, ingin 0", len(repo.hasilHitung[0].Rincian))
	}
	if repo.status != domain.PeriodeDihitung {
		t.Fatalf("status periode = %s, ingin DIHITUNG", repo.status)
	}
}

func TestHitungGajiPeriodeBulananContohB(t *testing.T) {
	repo := newFakeRepo()
	periode := domain.PeriodeGaji{
		ID: 8, Tipe: domain.PeriodeBulanan,
		TanggalMulai: tanggal(2026, 11, 1), TanggalSelesai: tanggal(2026, 11, 30),
		Status: domain.PeriodeTerbuka,
	}
	repo.periode = periode
	pekerja := domain.Pekerja{
		ID: 8, Kode: "P-008", Nama: "Maya", Jenis: domain.PekerjaBulanan,
		TanggalMulai: tanggal(2026, 11, 1), Aktif: true,
	}
	repo.workers[domain.PekerjaBulanan] = []domain.Pekerja{pekerja}
	repo.upah[pekerja.ID] = []domain.Upah{{Nominal: 3460000, BerlakuMulai: tanggal(2026, 11, 1)}}
	repo.kasbon[pekerja.ID] = []domain.Kasbon{{ID: 8, PekerjaID: pekerja.ID, Sisa: 1500000}}

	// 13 hari hadir dari 25 hari kerja November. Pekerja bulanan penuh tetap
	// dibayar pokok sebulan; kehadiran memengaruhi lembur dan potongan telat.
	hariHadir := []int{2, 3, 4, 5, 6, 7, 9, 10, 11, 12, 13, 14, 16}
	for i, hari := range hariHadir {
		tgl := tanggal(2026, 11, hari)
		masuk := time.Date(2026, 11, hari, 15, 0, 0, 0, time.Local)
		keluar := time.Date(2026, 11, hari, 23, 0, 0, 0, time.Local)
		switch i {
		case 0:
			masuk = masuk.Add(20 * time.Minute)
		case len(hariHadir) - 1:
			keluar = time.Date(2026, 11, hari+1, 2, 0, 0, 0, time.Local)
		}
		repo.absensi = append(repo.absensi,
			absensiBulanan(int64(i+1), pekerja.ID, pekerja, tgl, masuk, keluar))
	}

	if err := serviceUji(repo).HitungGajiPeriode(context.Background(), periode.ID); err != nil {
		t.Fatalf("hitung gaji: %v", err)
	}
	slip := repo.slips[pekerja.ID]
	if slip.UpahPokok != 3460000 {
		t.Fatalf("upah pokok = %d, ingin 3460000", slip.UpahPokok)
	}
	// 180 menit lembur: 60 x 1,5 + 120 x 2 = 330 menit setara, 330/60 x 20.000.
	if slip.TotalLembur != 110000 {
		t.Fatalf("total lembur = %d, ingin 110000", slip.TotalLembur)
	}
	if slip.PotonganTelat != 20000 {
		t.Fatalf("potongan telat = %d, ingin 20000", slip.PotonganTelat)
	}
	if slip.GajiBruto != 3550000 {
		t.Fatalf("gaji bruto = %d, ingin 3550000", slip.GajiBruto)
	}
	if slip.PotonganKasbon != 1065000 {
		t.Fatalf("potongan kasbon = %d, ingin 1065000", slip.PotonganKasbon)
	}
	if slip.GajiBersih != 2485000 {
		t.Fatalf("gaji bersih = %d, ingin 2485000", slip.GajiBersih)
	}
	if slip.HariKerjaDibayar != 25 || slip.HariKerjaPeriode != 25 {
		t.Fatalf("hari kerja = %d/%d, ingin 25/25",
			slip.HariKerjaDibayar, slip.HariKerjaPeriode)
	}
	if len(repo.hasilHitung) != 1 || len(repo.hasilHitung[0].Rincian) != 13 {
		t.Fatalf("hasil hitung = %d slip, ingin 1 slip dengan 13 rincian", len(repo.hasilHitung))
	}
	if len(repo.hasilHitung[0].Potongan) != 1 ||
		repo.hasilHitung[0].Potongan[0].Jumlah != 1065000 {
		t.Fatalf("potongan kasbon tersimpan = %+v", repo.hasilHitung[0].Potongan)
	}
}

func TestHitungUlangTidakMenghitungKoreksiDuaKali(t *testing.T) {
	repo := newFakeRepo()
	periode := domain.PeriodeGaji{
		ID: 3, Tipe: domain.PeriodeMingguan,
		TanggalMulai: tanggal(2026, 11, 2), TanggalSelesai: tanggal(2026, 11, 7),
		Status: domain.PeriodeTerbuka,
	}
	repo.periode = periode
	pekerja := domain.Pekerja{
		ID: 3, Kode: "P-003", Nama: "Andi", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, 11, 1), Aktif: true,
	}
	repo.workers[domain.PekerjaHarian] = []domain.Pekerja{pekerja}
	repo.upah[pekerja.ID] = []domain.Upah{{Nominal: 100000, BerlakuMulai: tanggal(2026, 11, 1)}}
	repo.koreksi[pekerja.ID] = []domain.KoreksiGaji{{
		ID: 1, PekerjaID: pekerja.ID, Jumlah: 50000, Keterangan: "Koreksi lembur",
		DibuatPada: tanggal(2026, 11, 1),
	}}
	for i := 0; i < 6; i++ {
		tgl := tanggal(2026, 11, 2+i)
		repo.absensi = append(repo.absensi, absensiHarian(
			int64(i+1), pekerja.ID, pekerja, tgl,
			time.Date(tgl.Year(), tgl.Month(), tgl.Day(), 7, 0, 0, 0, time.Local),
			time.Date(tgl.Year(), tgl.Month(), tgl.Day(), 15, 0, 0, 0, time.Local),
		))
	}

	svc := serviceUji(repo)
	if err := svc.HitungGajiPeriode(context.Background(), periode.ID); err != nil {
		t.Fatalf("hitung pertama: %v", err)
	}
	pertama := repo.slips[pekerja.ID]
	if pertama.TotalKoreksi != 50000 || pertama.GajiBruto != 650000 {
		t.Fatalf("slip pertama = koreksi %d bruto %d", pertama.TotalKoreksi, pertama.GajiBruto)
	}

	if err := svc.HitungGajiPeriode(context.Background(), periode.ID); err != nil {
		t.Fatalf("hitung ulang: %v", err)
	}
	kedua := repo.slips[pekerja.ID]
	if kedua.TotalKoreksi != 50000 || kedua.GajiBruto != 650000 {
		t.Fatalf("slip hitung ulang = koreksi %d bruto %d, koreksi terhitung dua kali",
			kedua.TotalKoreksi, kedua.GajiBruto)
	}
}

func TestHitungGajiPeriodeDitolakSaatPeriodeDitutup(t *testing.T) {
	repo := newFakeRepo()
	repo.periode = domain.PeriodeGaji{ID: 4, Tipe: domain.PeriodeMingguan, Status: domain.PeriodeDitutup}
	err := serviceUji(repo).HitungGajiPeriode(context.Background(), repo.periode.ID)
	if !errors.Is(err, ErrPeriodeTertutup) {
		t.Fatalf("galat = %v, ingin ErrPeriodeTertutup", err)
	}
}

func TestHitungGajiPeriodeBulananTanpaAbsensi(t *testing.T) {
	repo := newFakeRepo()
	periode := domain.PeriodeGaji{
		ID: 5, Tipe: domain.PeriodeBulanan,
		TanggalMulai: tanggal(2026, 11, 1), TanggalSelesai: tanggal(2026, 11, 30),
		Status: domain.PeriodeTerbuka,
	}
	repo.periode = periode
	pekerja := domain.Pekerja{
		ID: 5, Kode: "P-005", Nama: "Rina", Jenis: domain.PekerjaBulanan,
		TanggalMulai: tanggal(2026, 11, 1), Aktif: true,
	}
	repo.workers[domain.PekerjaBulanan] = []domain.Pekerja{pekerja}
	repo.upah[pekerja.ID] = []domain.Upah{{Nominal: 3460000, BerlakuMulai: tanggal(2026, 11, 1)}}

	if err := serviceUji(repo).HitungGajiPeriode(context.Background(), periode.ID); err != nil {
		t.Fatalf("hitung gaji: %v", err)
	}
	slip := repo.slips[pekerja.ID]
	if slip.UpahPokok != 3460000 || slip.GajiBruto != 3460000 || slip.GajiBersih != 3460000 {
		t.Fatalf("slip bulanan tanpa absensi = pokok %d bruto %d bersih %d",
			slip.UpahPokok, slip.GajiBruto, slip.GajiBersih)
	}
	if slip.HariKerjaDibayar != 25 || slip.HariKerjaPeriode != 25 {
		t.Fatalf("hari kerja = %d/%d, ingin 25/25",
			slip.HariKerjaDibayar, slip.HariKerjaPeriode)
	}
	if len(repo.rincian[pekerja.ID]) != 0 {
		t.Fatalf("rincian harian = %d, ingin 0", len(repo.rincian[pekerja.ID]))
	}
}

func TestPeriksaPrasyaratHitungPeriodeSebelumnyaBelumDitutup(t *testing.T) {
	repo := newFakeRepo()
	repo.periode = domain.PeriodeGaji{
		ID: 6, Tipe: domain.PeriodeMingguan, Status: domain.PeriodeTerbuka,
		TanggalMulai: tanggal(2026, 11, 9), TanggalSelesai: tanggal(2026, 11, 14),
	}
	repo.sebelumnya = &domain.PeriodeGaji{
		ID: 5, Tipe: domain.PeriodeMingguan, Status: domain.PeriodeTerbuka,
		TanggalMulai: tanggal(2026, 11, 2), TanggalSelesai: tanggal(2026, 11, 7),
	}

	prasyarat, err := serviceUji(repo).PeriksaPrasyaratHitung(context.Background(), repo.periode.ID)
	if err != nil {
		t.Fatalf("periksa prasyarat: %v", err)
	}
	if !prasyarat.StatusPeriodeOK {
		t.Fatal("status periode seharusnya lolos")
	}
	if prasyarat.PeriodeSebelumnyaOK {
		t.Fatal("periode sebelumnya yang masih terbuka seharusnya menggagalkan prasyarat")
	}
	if prasyarat.Pesan == "" {
		t.Fatal("pesan prasyarat seharusnya berisi alasan gagal")
	}
}

func TestPeriksaPrasyaratHitungAbsensiBelumLengkap(t *testing.T) {
	repo := newFakeRepo()
	repo.periode = domain.PeriodeGaji{
		ID: 7, Tipe: domain.PeriodeMingguan, Status: domain.PeriodeTerbuka,
		TanggalMulai: tanggal(2026, 11, 2), TanggalSelesai: tanggal(2026, 11, 7),
	}
	pekerja := domain.Pekerja{
		ID: 7, Kode: "P-007", Nama: "Dewi", Jenis: domain.PekerjaHarian,
		TanggalMulai: tanggal(2026, 11, 1), Aktif: true,
	}
	tanpaKeluar := absensiHarian(
		1, pekerja.ID, pekerja, tanggal(2026, 11, 2),
		time.Date(2026, 11, 2, 7, 0, 0, 0, time.Local),
		time.Date(2026, 11, 2, 15, 0, 0, 0, time.Local),
	)
	tanpaKeluar.Absensi.JamKeluar = nil
	repo.absensi = []domain.AbsensiLengkap{tanpaKeluar}

	prasyarat, err := serviceUji(repo).PeriksaPrasyaratHitung(context.Background(), repo.periode.ID)
	if err != nil {
		t.Fatalf("periksa prasyarat: %v", err)
	}
	if prasyarat.AbsensiLengkap {
		t.Fatal("absensi tanpa jam keluar seharusnya menggagalkan prasyarat")
	}
	if prasyarat.JumlahAbsensiKurang != 1 {
		t.Fatalf("jumlah absensi kurang = %d, ingin 1", prasyarat.JumlahAbsensiKurang)
	}
}
