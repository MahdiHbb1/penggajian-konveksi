// Pemilik: Edgar | Modul 2 | Aturan: 3, 14, 19
package jadwal

import (
	"context"
	"testing"
	"time"

	"penggajian-konveksi/internal/domain"
)

func TestBuatJadwalValidasi(t *testing.T) {
	svc := NewService(NewRepository(nil))
	kasus := []struct {
		nama    string
		tanggal time.Time
	}{
		{"hari Minggu", time.Date(2026, 11, 1, 0, 0, 0, 0, time.Local)},
		{"jadwal ganda", time.Date(2026, 11, 2, 0, 0, 0, 0, time.Local)},
		{"periode tertutup", time.Date(2026, 10, 5, 0, 0, 0, 0, time.Local)},
	}

	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			t.Skip("belum diimplementasi: BuatJadwal")
			jadwal := domain.JadwalShift{PekerjaID: 1, ShiftID: 1, Tanggal: k.tanggal}
			if _, err := svc.BuatJadwal(context.Background(), jadwal); err == nil {
				t.Fatal("jadwal tidak valid seharusnya ditolak")
			}
		})
	}
}

func TestUbahJadwalDenganAbsensiDitolak(t *testing.T) {
	t.Skip("belum diimplementasi: UbahJadwal")
	svc := NewService(NewRepository(nil))
	jadwal := domain.JadwalShift{ID: 1, PekerjaID: 1, ShiftID: 2, Tanggal: time.Now()}
	if err := svc.UbahJadwal(context.Background(), jadwal); err == nil {
		t.Fatal("jadwal yang sudah punya absensi seharusnya tidak dapat diubah")
	}
}

func TestSalinJadwalTidakMenimpa(t *testing.T) {
	t.Skip("belum diimplementasi: SalinJadwalMingguLalu")
	svc := NewService(NewRepository(nil))
	disalin, dilewati, err := svc.SalinJadwalMingguLalu(context.Background(), time.Now())
	if err != nil {
		t.Fatalf("salin jadwal: %v", err)
	}
	if disalin < 0 || dilewati < 0 {
		t.Fatalf("jumlah tidak valid: disalin=%d dilewati=%d", disalin, dilewati)
	}
}
