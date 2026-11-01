// Pemilik: Mahdi | Modul 1 | Aturan: 8, 11, 16
package pekerja

import (
	"context"
	"testing"
	"time"

	"penggajian-konveksi/internal/domain"
)

func TestTambahPekerjaValidasi(t *testing.T) {
	svc := NewService(NewRepository(nil))
	mulai := time.Date(2026, 11, 2, 0, 0, 0, 0, time.Local)
	kasus := []struct {
		nama    string
		pekerja domain.Pekerja
		upah    domain.Rupiah
	}{
		{"kode kosong", domain.Pekerja{Nama: "Budi", Jenis: domain.PekerjaHarian, TanggalMulai: mulai}, 100000},
		{"kode terlalu panjang", domain.Pekerja{Kode: "123456789012345678901", Nama: "Budi", Jenis: domain.PekerjaHarian, TanggalMulai: mulai}, 100000},
		{"upah nol", domain.Pekerja{Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian, TanggalMulai: mulai}, 0},
	}

	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			t.Skip("belum diimplementasi: TambahPekerja")
			if _, err := svc.TambahPekerja(context.Background(), k.pekerja, k.upah); err == nil {
				t.Fatal("data pekerja tidak valid seharusnya ditolak")
			}
		})
	}
}

func TestUbahTanggalBerhentiSebelumMulai(t *testing.T) {
	t.Skip("belum diimplementasi: UbahPekerja")
	svc := NewService(NewRepository(nil))
	mulai := time.Date(2026, 11, 10, 0, 0, 0, 0, time.Local)
	berhenti := time.Date(2026, 11, 1, 0, 0, 0, 0, time.Local)
	pekerja := domain.Pekerja{ID: 1, Kode: "P-001", Nama: "Budi", Jenis: domain.PekerjaHarian, TanggalMulai: mulai, TanggalBerhenti: &berhenti}
	if err := svc.UbahPekerja(context.Background(), pekerja); err == nil {
		t.Fatal("tanggal berhenti sebelum mulai seharusnya ditolak")
	}
}

func TestRiwayatUpahTanggalGanda(t *testing.T) {
	t.Skip("belum diimplementasi: TambahRiwayatUpah")
	svc := NewService(NewRepository(nil))
	upah := domain.Upah{PekerjaID: 1, Nominal: 120000, BerlakuMulai: time.Date(2026, 11, 2, 0, 0, 0, 0, time.Local)}
	if err := svc.TambahRiwayatUpah(context.Background(), upah); err == nil {
		t.Fatal("upah pada tanggal yang sama seharusnya ditolak")
	}
}
