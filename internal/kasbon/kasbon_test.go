// Pemilik: Edgar | Modul 2 | Aturan: 12, 14
package kasbon

import (
	"context"
	"testing"
	"time"

	"penggajian-konveksi/internal/domain"
)

func TestCatatKasbonValidasi(t *testing.T) {
	svc := NewService(NewRepository(nil))
	kasus := []struct {
		nama    string
		jumlah  domain.Rupiah
		tanggal time.Time
	}{
		{"jumlah nol", 0, time.Now()},
		{"jumlah negatif", -10000, time.Now()},
		{"tanggal periode tertutup", 100000, time.Date(2026, 10, 5, 0, 0, 0, 0, time.Local)},
	}

	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			t.Skip("belum diimplementasi: CatatKasbon")
			kasbon := domain.Kasbon{PekerjaID: 1, Tanggal: k.tanggal, Jumlah: k.jumlah, Sisa: k.jumlah}
			if _, err := svc.CatatKasbon(context.Background(), kasbon); err == nil {
				t.Fatal("kasbon tidak valid seharusnya ditolak")
			}
		})
	}
}

func TestKasbonDiprosesDariYangTertua(t *testing.T) {
	t.Skip("belum diimplementasi: KasbonDiprosesSampai")
	svc := NewService(NewRepository(nil))
	daftar, err := svc.KasbonDiprosesSampai(context.Background(), 1, time.Now())
	if err != nil {
		t.Fatalf("ambil kasbon: %v", err)
	}
	for i := 1; i < len(daftar); i++ {
		if daftar[i].Tanggal.Before(daftar[i-1].Tanggal) {
			t.Fatal("kasbon harus terurut dari yang tertua")
		}
	}
}
