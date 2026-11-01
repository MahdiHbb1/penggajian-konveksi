// Pemilik: Farah | Modul 4 | Aturan: 5, 6, 7, 9, 12
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
	}

	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			t.Skip("belum diimplementasi: HitungUpahPerJam")
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
	}

	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			t.Skip("belum diimplementasi: HitungPotonganTerlambat")
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
	}

	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			t.Skip("belum diimplementasi: HitungUpahLembur")
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
	}

	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			t.Skip("belum diimplementasi: HitungHariKerja")
			if hasil := HitungHariKerja(k.mulai, k.selesai); hasil != k.ingin {
				t.Fatalf("hari kerja = %d, ingin %d", hasil, k.ingin)
			}
		})
	}
}

func TestHitungUpahPokokBulanan(t *testing.T) {
	t.Skip("belum diimplementasi: HitungUpahPokokBulanan")
	segmen := []SegmenUpah{{Nominal: 3460000, HariDijalani: 13, HariKerjaBulan: 25}}
	if hasil := HitungUpahPokokBulanan(segmen); hasil != 1799200 {
		t.Fatalf("upah proporsional = %d, ingin 1799200", hasil)
	}
}

func TestHitungGajiBruto(t *testing.T) {
	t.Skip("belum diimplementasi: HitungGajiBruto")
	if hasil := HitungGajiBruto(600000, 56250, 12500, 0); hasil != 643750 {
		t.Fatalf("bruto harian = %d, ingin 643750", hasil)
	}
	if hasil := HitungGajiBruto(3460000, 110000, 20000, 0); hasil != 3550000 {
		t.Fatalf("bruto bulanan = %d, ingin 3550000", hasil)
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
	}

	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			t.Skip("belum diimplementasi: AlokasiPotonganKasbon")
			_, total := AlokasiPotonganKasbon(k.bruto, k.kasbon)
			if total != k.ingin {
				t.Fatalf("potongan kasbon = %d, ingin %d", total, k.ingin)
			}
		})
	}
}

func tanggal(tahun int, bulan time.Month, hari int) time.Time {
	return time.Date(tahun, bulan, hari, 0, 0, 0, 0, time.Local)
}
