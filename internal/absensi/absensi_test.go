// Pemilik: Zahra | Modul 3 | Aturan: 3, 5, 6
package absensi

import (
	"errors"
	"testing"
	"time"
)

func waktu(t *testing.T, tanggal, jam string) time.Time {
	t.Helper()
	hasil, err := time.ParseInLocation("2006-01-02 15:04", tanggal+" "+jam, time.Local)
	if err != nil {
		t.Fatal(err)
	}
	return hasil
}

func TestHitungMenitTerlambat(t *testing.T) {
	mulai := waktu(t, "2026-11-02", "07:00")
	svc := NewService(nil)
	kasus := []struct {
		nama     string
		jamMasuk time.Time
		ingin    int
	}{
		{"tepat waktu", mulai, 0},
		{"datang lebih awal", mulai.Add(-20 * time.Minute), 0},
		{"tepat 15 menit", mulai.Add(15 * time.Minute), 15},
		{"16 menit", mulai.Add(16 * time.Minute), 16},
		{"61 menit", mulai.Add(61 * time.Minute), 61},
	}
	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			if hasil := svc.HitungMenitTerlambat(k.jamMasuk, mulai); hasil != k.ingin {
				t.Fatalf("menit terlambat = %d, ingin %d", hasil, k.ingin)
			}
		})
	}
}

func TestHitungMenitLembur(t *testing.T) {
	selesai := waktu(t, "2026-11-02", "15:00")
	svc := NewService(nil)

	kasus := []struct {
		nama      string
		jamKeluar time.Time
		akhir     time.Time
		ingin     int
	}{
		{"tepat selesai", selesai, selesai, 0},
		{"60 menit", selesai.Add(60 * time.Minute), selesai, 60},
		{"61 menit", selesai.Add(61 * time.Minute), selesai, 61},
		{"150 menit", selesai.Add(150 * time.Minute), selesai, 150},
		{
			"keluar lewat tengah malam",
			waktu(t, "2026-11-03", "00:30"),
			waktu(t, "2026-11-02", "23:00"),
			90,
		},
		{"pulang lebih awal", selesai.Add(-30 * time.Minute), selesai, 0},
	}
	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			if hasil := svc.HitungMenitLembur(k.jamKeluar, k.akhir); hasil != k.ingin {
				t.Fatalf("menit lembur = %d, ingin %d", hasil, k.ingin)
			}
		})
	}
}

func TestNormalisasiJamKeluar(t *testing.T) {
	masuk := waktu(t, "2026-11-02", "22:00")
	keluar := waktu(t, "2026-11-02", "00:30")

	hasil, err := normalisasiJamKeluar(masuk, &keluar)
	if err != nil {
		t.Fatalf("tidak mengharapkan error: %v", err)
	}
	ingin := waktu(t, "2026-11-03", "00:30")
	if !hasil.Equal(ingin) {
		t.Fatalf("jam keluar = %v, ingin %v", hasil, ingin)
	}
}

func TestNormalisasiJamKeluarMenolakWaktuSama(t *testing.T) {
	masuk := waktu(t, "2026-11-02", "08:00")
	keluar := masuk

	_, err := normalisasiJamKeluar(masuk, &keluar)
	if !errors.Is(err, ErrJamKeluarTidakValid) {
		t.Fatalf("error = %v, ingin ErrJamKeluarTidakValid", err)
	}
}

func TestNormalisasiJamKeluarBolehKosong(t *testing.T) {
	masuk := waktu(t, "2026-11-02", "08:00")
	hasil, err := normalisasiJamKeluar(masuk, nil)
	if err != nil {
		t.Fatalf("tidak mengharapkan error: %v", err)
	}
	if hasil != nil {
		t.Fatalf("jam keluar = %v, ingin nil", hasil)
	}
}

func TestTanggalSama(t *testing.T) {
	a := waktu(t, "2026-11-02", "07:00")
	b := waktu(t, "2026-11-02", "23:59")
	c := waktu(t, "2026-11-03", "00:00")

	if !tanggalSama(a, b) {
		t.Fatal("dua waktu pada tanggal sama seharusnya cocok")
	}
	if tanggalSama(a, c) {
		t.Fatal("dua waktu pada tanggal berbeda seharusnya tidak cocok")
	}
}
