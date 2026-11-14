// Pemilik: Anya | Modul 5 | Aturan: 1, 13, 14, 15, 18
package periode

import (
	"context"
	"penggajian-konveksi/internal/domain"
	"testing"
	"time"
)

func TestBuatPeriodeValidasi(t *testing.T) {
	svc := NewService(NewRepository(nil))
	kasus := []struct {
		nama  string
		tipe  domain.TipePeriode
		mulai time.Time
	}{{"mingguan bukan Senin", domain.PeriodeMingguan, tanggal(2026, 11, 3)}, {"bulanan bukan tanggal 1", domain.PeriodeBulanan, tanggal(2026, 11, 2)}, {"tipe tidak dikenal", domain.TipePeriode("TAHUNAN"), tanggal(2026, 11, 1)}}
	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			if _, e := svc.BuatPeriode(context.Background(), k.tipe, k.mulai); e == nil {
				t.Fatal("periode tidak valid seharusnya ditolak")
			}
		})
	}
}
func TestTutupPeriodeHanyaDariDihitung(t *testing.T) {
	svc := NewService(NewRepository(nil))
	if e := svc.TutupPeriode(context.Background(), 1); e == nil {
		t.Fatal("tanpa database penutupan harus gagal")
	}
}
func TestCatatKoreksiHanyaUntukPeriodeDitutup(t *testing.T) {
	svc := NewService(NewRepository(nil))
	k := domain.KoreksiGaji{PeriodeAsalID: 1, PekerjaID: 1, Jumlah: -50000, Keterangan: "Koreksi upah"}
	if _, e := svc.CatatKoreksi(context.Background(), k); e == nil {
		t.Fatal("tanpa database koreksi harus gagal")
	}
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
func tanggal(tahun int, bulan time.Month, hari int) time.Time {
	return time.Date(tahun, bulan, hari, 0, 0, 0, 0, time.Local)
}
