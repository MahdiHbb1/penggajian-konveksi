// Pemilik: Farah | Modul 4 | Aturan: 5, 6, 7, 9, 10, 12
package penggajian

import (
	"testing"
	"time"

	"penggajian-konveksi/internal/domain"
)

func TestHitungUpahPerJam(t *testing.T) {
	kasus := []struct {
		nama    string
		nominal domain.Rupiah
		jenis   domain.JenisPekerja
		ingin   domain.Rupiah
	}{
		{"harian 100000", 100000, domain.PekerjaHarian, 12500},
		{"bulanan 3460000", 3460000, domain.PekerjaBulanan, 20000},
		{"nominal nol", 0, domain.PekerjaHarian, 0},
		{"jenis tidak dikenal", 100000, "", 0},
	}

	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			if hasil := HitungUpahPerJam(k.nominal, k.jenis); hasil != k.ingin {
				t.Fatalf("upah per jam = %d, ingin %d", hasil, k.ingin)
			}
		})
	}
}

func TestHitungPotonganTerlambat(t *testing.T) {
	kasus := []struct {
		nama       string
		menitTelat int
		ingin      domain.Rupiah
	}{
		{"tepat 15 menit", 15, 0},
		{"16 menit", 16, 12500},
		{"20 menit", 20, 12500},
		{"tidak terlambat", 0, 0},
	}

	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			if hasil := HitungPotonganTerlambat(k.menitTelat, 12500); hasil != k.ingin {
				t.Fatalf("potongan = %d, ingin %d", hasil, k.ingin)
			}
		})
	}
}

func TestHitungUpahLembur(t *testing.T) {
	kasus := []struct {
		nama        string
		menitLembur int
		upahPerJam  domain.Rupiah
		ingin       domain.Rupiah
	}{
		{"60 menit", 60, 12500, 18750},
		{"61 menit", 61, 12500, 19166},
		{"90 menit", 90, 12500, 31250},
		{"150 menit", 150, 12500, 56250},
		{"bulanan 180 menit", 180, 20000, 110000},
		{"tidak lembur", 0, 12500, 0},
	}

	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			if hasil := HitungUpahLembur(k.menitLembur, k.upahPerJam); hasil != k.ingin {
				t.Fatalf("lembur = %d, ingin %d", hasil, k.ingin)
			}
		})
	}
}

func TestHitungHariKerja(t *testing.T) {
	kasus := []struct {
		nama    string
		mulai   time.Time
		selesai time.Time
		ingin   int
	}{
		{"November 2026", tanggal(2026, 11, 1), tanggal(2026, 11, 30), 25},
		{"16 sampai 30 November 2026", tanggal(2026, 11, 16), tanggal(2026, 11, 30), 13},
		{"hanya hari Minggu", tanggal(2026, 11, 1), tanggal(2026, 11, 1), 0},
		{"satu pekan penuh", tanggal(2026, 11, 2), tanggal(2026, 11, 7), 6},
		{"rentang terbalik", tanggal(2026, 11, 30), tanggal(2026, 11, 1), 0},
	}

	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			if hasil := HitungHariKerja(k.mulai, k.selesai); hasil != k.ingin {
				t.Fatalf("hari kerja = %d, ingin %d", hasil, k.ingin)
			}
		})
	}
}

func TestHitungUpahPokokHarian(t *testing.T) {
	rincian := []domain.SlipGajiDetail{
		{UpahPokokHari: 100000},
		{UpahPokokHari: 100000},
		{UpahPokokHari: 100000},
		{UpahPokokHari: 100000},
		{UpahPokokHari: 100000},
		{UpahPokokHari: 100000},
	}
	if hasil := HitungUpahPokokHarian(rincian); hasil != 600000 {
		t.Fatalf("upah pokok harian = %d, ingin 600000", hasil)
	}
}

func TestHitungUpahPokokBulanan(t *testing.T) {
	segmen := []SegmenUpah{{Nominal: 3460000, HariDijalani: 13, HariKerjaBulan: 25}}
	if hasil := HitungUpahPokokBulanan(segmen); hasil != 1799200 {
		t.Fatalf("upah proporsional = %d, ingin 1799200", hasil)
	}
}

func TestHitungUpahPokokBulananContohC(t *testing.T) {
	periode := domain.PeriodeGaji{
		Tipe:           domain.PeriodeBulanan,
		TanggalMulai:   tanggal(2026, 11, 1),
		TanggalSelesai: tanggal(2026, 11, 30),
	}
	riwayat := []domain.Upah{
		{Nominal: 3000000, BerlakuMulai: tanggal(2026, 1, 1)},
		{Nominal: 3460000, BerlakuMulai: tanggal(2026, 11, 16)},
	}

	segmen := segmenUpahBulanan(
		periode,
		tanggal(2026, 11, 1),
		tanggal(2026, 11, 30),
		riwayat,
	)
	if len(segmen) != 2 {
		t.Fatalf("jumlah segmen = %d, ingin 2", len(segmen))
	}
	if segmen[0].HariDijalani != 12 || segmen[1].HariDijalani != 13 {
		t.Fatalf("hari segmen = %d dan %d, ingin 12 dan 13",
			segmen[0].HariDijalani, segmen[1].HariDijalani)
	}
	// 3.000.000 x 12/25 = 1.440.000; 3.460.000 x 13/25 = 1.799.200.
	if hasil := HitungUpahPokokBulanan(segmen); hasil != 3239200 {
		t.Fatalf("upah pokok contoh C = %d, ingin 3239200", hasil)
	}
}

func TestHitungGajiBruto(t *testing.T) {
	if hasil := HitungGajiBruto(600000, 56250, 12500, 0); hasil != 643750 {
		t.Fatalf("bruto harian = %d, ingin 643750", hasil)
	}
	if hasil := HitungGajiBruto(3460000, 110000, 20000, 0); hasil != 3550000 {
		t.Fatalf("bruto bulanan = %d, ingin 3550000", hasil)
	}
	if hasil := HitungGajiBruto(600000, 0, 0, -50000); hasil != 550000 {
		t.Fatalf("bruto dengan koreksi negatif = %d, ingin 550000", hasil)
	}
}

func TestAlokasiPotonganKasbon(t *testing.T) {
	kasus := []struct {
		nama   string
		bruto  domain.Rupiah
		kasbon []domain.Kasbon
		ingin  domain.Rupiah
	}{
		{"contoh A", 643750, []domain.Kasbon{{ID: 1, Sisa: 300000}}, 193125},
		{"contoh B", 3550000, []domain.Kasbon{{ID: 2, Sisa: 1500000}}, 1065000},
		{"kasbon lebih kecil dari batas", 1000000, []domain.Kasbon{{ID: 3, Sisa: 100000}}, 100000},
		{"tanpa kasbon", 1000000, nil, 0},
		{"bruto nol", 0, []domain.Kasbon{{ID: 4, Sisa: 100000}}, 0},
	}

	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			_, total := AlokasiPotonganKasbon(k.bruto, k.kasbon)
			if total != k.ingin {
				t.Fatalf("potongan kasbon = %d, ingin %d", total, k.ingin)
			}
		})
	}
}

func TestAlokasiPotonganKasbonUrutTertua(t *testing.T) {
	kasbon := []domain.Kasbon{
		{ID: 10, Sisa: 200000},
		{ID: 11, Sisa: 200000},
	}
	potongan, total := AlokasiPotonganKasbon(1000000, kasbon)
	if total != 300000 {
		t.Fatalf("total potongan = %d, ingin 300000", total)
	}
	if len(potongan) != 2 {
		t.Fatalf("jumlah potongan = %d, ingin 2", len(potongan))
	}
	if potongan[0].KasbonID != 10 || potongan[0].Jumlah != 200000 {
		t.Fatalf("potongan pertama = %+v, ingin kasbon 10 sebesar 200000", potongan[0])
	}
	if potongan[1].KasbonID != 11 || potongan[1].Jumlah != 100000 {
		t.Fatalf("potongan kedua = %+v, ingin kasbon 11 sebesar 100000", potongan[1])
	}
}

func TestHitungMenitAbsensiLewatTengahMalam(t *testing.T) {
	shiftMulai := time.Date(2026, 11, 2, 15, 0, 0, 0, time.Local)
	shiftSelesai := time.Date(2026, 11, 2, 23, 0, 0, 0, time.Local)
	masukTelat := time.Date(2026, 11, 2, 15, 16, 0, 0, time.Local)
	keluarLewat := time.Date(2026, 11, 3, 0, 30, 0, 0, time.Local)

	if hasil := hitungMenitTerlambat(masukTelat, shiftMulai); hasil != 16 {
		t.Fatalf("menit terlambat = %d, ingin 16", hasil)
	}
	if hasil := hitungMenitLembur(keluarLewat, shiftSelesai); hasil != 90 {
		t.Fatalf("menit lembur lewat tengah malam = %d, ingin 90", hasil)
	}
}

func tanggal(tahun int, bulan time.Month, hari int) time.Time {
	return time.Date(tahun, bulan, hari, 0, 0, 0, 0, time.Local)
}
