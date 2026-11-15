// Pemilik: Edgar | Modul 2 | Aturan: 3, 14, 19
package jadwal

import (
	"context"
	"errors"
	"testing"
	"time"

	"penggajian-konveksi/internal/domain"
)

func tgl(tahun int, bulan time.Month, hari int) time.Time {
	return time.Date(tahun, bulan, hari, 0, 0, 0, 0, time.Local)
}

func TestBuatJadwalHariMingguDitolak(t *testing.T) {
	// Pemeriksaan hari Minggu terjadi sebelum akses database.
	svc := NewService(NewRepository(nil))
	jadwal := domain.JadwalShift{PekerjaID: 1, ShiftID: 1, Tanggal: tgl(2026, time.November, 1)}
	_, err := svc.BuatJadwal(context.Background(), jadwal)
	if !errors.Is(err, ErrHariMinggu) {
		t.Fatalf("galat = %v, harapan ErrHariMinggu", err)
	}
}

func TestBuatJadwalValidasiMasukan(t *testing.T) {
	svc := NewService(NewRepository(nil))
	senin := tgl(2026, time.November, 2)
	kasus := []struct {
		nama    string
		jadwal  domain.JadwalShift
		harapan error
	}{
		{"pekerja kosong", domain.JadwalShift{ShiftID: 1, Tanggal: senin}, ErrPekerjaTidakAktif},
		{"shift kosong", domain.JadwalShift{PekerjaID: 1, Tanggal: senin}, ErrShiftTidakValid},
		{"tanggal kosong", domain.JadwalShift{PekerjaID: 1, ShiftID: 1}, ErrTanggalKosong},
	}
	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			if _, err := svc.BuatJadwal(context.Background(), k.jadwal); !errors.Is(err, k.harapan) {
				t.Fatalf("galat = %v, harapan %v", err, k.harapan)
			}
		})
	}
}

func TestSeninPekan(t *testing.T) {
	kasus := []struct {
		masukan, harapan time.Time
	}{
		{tgl(2026, time.November, 2), tgl(2026, time.November, 2)}, // Senin
		{tgl(2026, time.November, 7), tgl(2026, time.November, 2)}, // Sabtu
		{tgl(2026, time.November, 8), tgl(2026, time.November, 2)}, // Minggu: pekan sebelumnya
		{tgl(2026, time.November, 9), tgl(2026, time.November, 9)}, // Senin berikutnya
		{tgl(2026, time.March, 1), tgl(2026, time.February, 23)},   // lintas bulan
	}
	for _, k := range kasus {
		if got := SeninPekan(k.masukan); !got.Equal(k.harapan) {
			t.Errorf("SeninPekan(%s) = %s, harapan %s", k.masukan.Format(layoutTanggal), got.Format(layoutTanggal), k.harapan.Format(layoutTanggal))
		}
	}
}

func TestPekanDefaultHariMinggu(t *testing.T) {
	// Minggu 8 Nov 2026 menampilkan pekan depan (Senin 9 Nov).
	if got := pekanDefault(tgl(2026, time.November, 8)); !got.Equal(tgl(2026, time.November, 9)) {
		t.Fatalf("pekanDefault = %s", got.Format(layoutTanggal))
	}
	// Rabu 4 Nov 2026 menampilkan pekan berjalan (Senin 2 Nov).
	if got := pekanDefault(tgl(2026, time.November, 4)); !got.Equal(tgl(2026, time.November, 2)) {
		t.Fatalf("pekanDefault = %s", got.Format(layoutTanggal))
	}
}

func TestWaktuShift(t *testing.T) {
	hari := tgl(2026, time.October, 28)

	siang, err := WaktuShift(hari, "15:00")
	if err != nil {
		t.Fatalf("WaktuShift: %v", err)
	}
	if siang.Hour() != 15 || siang.Minute() != 0 || siang.Day() != 28 {
		t.Fatalf("hasil = %s", siang)
	}

	if _, err := WaktuShift(hari, "07:00:00"); err != nil {
		t.Fatalf("format dengan detik harus diterima: %v", err)
	}
	if _, err := WaktuShift(hari, "bukan jam"); err == nil {
		t.Fatal("jam tidak valid seharusnya ditolak")
	}
}

func TestPesanGalatInput(t *testing.T) {
	if PesanGalatInput(ErrHariMinggu) != "Hari Minggu tidak dapat dijadwalkan." {
		t.Fatal("pesan hari Minggu tidak sesuai dokumen rancangan")
	}
	if PesanGalatInput(errors.New("koneksi putus")) != "" {
		t.Fatal("galat sistem tidak boleh ditampilkan ke pengguna")
	}
}

// Tes di bawah butuh MySQL berisi data uji, jadi dilewati pada tes otomatis.

func TestBuatJadwalGandaDitolak(t *testing.T) {
	t.Skip("butuh MySQL: jadwal kedua pada pekerja dan tanggal yang sama harus ditolak (ErrJadwalGanda)")
}

func TestBuatJadwalPeriodeTertutupDitolak(t *testing.T) {
	t.Skip("butuh MySQL: tanggal di periode DITUTUP harus ditolak (ErrPeriodeTertutup)")
}

func TestUbahJadwalDenganAbsensiDitolak(t *testing.T) {
	t.Skip("butuh MySQL: jadwal yang punya absensi tidak boleh diubah atau dibatalkan (ErrSudahAdaAbsensi)")
}

func TestSalinJadwalTidakMenimpa(t *testing.T) {
	t.Skip("butuh MySQL: sel yang sudah terisi dilewati dan dihitung pada jumlah dilewati")
}