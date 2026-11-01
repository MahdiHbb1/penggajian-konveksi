// Pemilik: Zahra | Modul 3 | Aturan: 3, 5, 6
package absensi

import (
	"testing"
	"time"
)

func TestHitungMenitTerlambat(t *testing.T) {
	mulai := time.Date(2026, 11, 2, 7, 0, 0, 0, time.Local)
	kasus := []struct {
		nama     string
		jamMasuk time.Time
		ingin    int
	}{
		{"tepat waktu", mulai, 0},
		{"tepat 15 menit", mulai.Add(15 * time.Minute), 15},
		{"16 menit", mulai.Add(16 * time.Minute), 16},
		{"datang lebih awal", mulai.Add(-20 * time.Minute), 0},
	}

	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			t.Skip("belum diimplementasi: HitungMenitTerlambat")
			svc := NewService(NewRepository(nil))
			if hasil := svc.HitungMenitTerlambat(k.jamMasuk, mulai); hasil != k.ingin {
				t.Fatalf("menit terlambat = %d, ingin %d", hasil, k.ingin)
			}
		})
	}
}

func TestHitungMenitLembur(t *testing.T) {
	selesai := time.Date(2026, 11, 2, 15, 0, 0, 0, time.Local)
	kasus := []struct {
		nama      string
		jamKeluar time.Time
		ingin     int
	}{
		{"tepat selesai", selesai, 0},
		{"60 menit", selesai.Add(60 * time.Minute), 60},
		{"150 menit", selesai.Add(150 * time.Minute), 150},
		{"shift siang lewat tengah malam", time.Date(2026, 11, 3, 0, 30, 0, 0, time.Local), 90},
	}

	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			t.Skip("belum diimplementasi: HitungMenitLembur")
			svc := NewService(NewRepository(nil))
			jamSelesai := time.Date(2026, 11, 2, 23, 0, 0, 0, time.Local)
			if hasil := svc.HitungMenitLembur(k.jamKeluar, jamSelesai); hasil != k.ingin {
				t.Fatalf("menit lembur = %d, ingin %d", hasil, k.ingin)
			}
		})
	}
}
