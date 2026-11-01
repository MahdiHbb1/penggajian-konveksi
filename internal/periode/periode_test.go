// Pemilik: Anya | Modul 5 | Aturan: 1, 13, 14, 15, 18
package periode

import (
	"context"
	"testing"
	"time"

	"penggajian-konveksi/internal/domain"
)

func TestBuatPeriodeValidasi(t *testing.T) {
	svc := NewService(NewRepository(nil))
	kasus := []struct {
		nama  string
		tipe  domain.TipePeriode
		mulai time.Time
	}{
		{"mingguan bukan Senin", domain.PeriodeMingguan, tanggal(2026, 11, 3)},
		{"bulanan bukan tanggal 1", domain.PeriodeBulanan, tanggal(2026, 11, 2)},
	}

	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			t.Skip("belum diimplementasi: BuatPeriode")
			if _, err := svc.BuatPeriode(context.Background(), k.tipe, k.mulai); err == nil {
				t.Fatal("periode tidak valid seharusnya ditolak")
			}
		})
	}
}

func TestTutupPeriodeHanyaDariDihitung(t *testing.T) {
	t.Skip("belum diimplementasi: TutupPeriode")
	svc := NewService(NewRepository(nil))
	if err := svc.TutupPeriode(context.Background(), 1); err == nil {
		t.Fatal("periode Terbuka seharusnya tidak dapat ditutup")
	}
}

func TestCatatKoreksiHanyaUntukPeriodeDitutup(t *testing.T) {
	t.Skip("belum diimplementasi: CatatKoreksi")
	svc := NewService(NewRepository(nil))
	koreksi := domain.KoreksiGaji{PeriodeAsalID: 1, Jumlah: -50000, Keterangan: "Koreksi upah"}
	if _, err := svc.CatatKoreksi(context.Background(), koreksi); err == nil {
		t.Fatal("koreksi untuk periode terbuka seharusnya ditolak")
	}
}

func TestKoreksiTidakDihitungDuaKali(t *testing.T) {
	t.Skip("belum diimplementasi: TandaiKoreksiDiterapkan")
	svc := NewService(NewRepository(nil))
	if err := svc.TandaiKoreksiDiterapkan(context.Background(), 1); err != nil {
		t.Fatalf("menandai koreksi: %v", err)
	}
}

func tanggal(tahun int, bulan time.Month, hari int) time.Time {
	return time.Date(tahun, bulan, hari, 0, 0, 0, 0, time.Local)
}
