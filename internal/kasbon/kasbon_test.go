// Pemilik: Edgar | Modul 2 | Aturan: 12, 14
package kasbon

import (
	"errors"
	"strings"
	"testing"
	"time"

	"penggajian-konveksi/internal/domain"
)

func TestValidasiKasbon(t *testing.T) {
	hari := time.Date(2026, 11, 2, 0, 0, 0, 0, time.Local)
	kasus := []struct {
		nama   string
		kasbon domain.Kasbon
		galat  error
	}{
		{"valid", domain.Kasbon{PekerjaID: 1, Tanggal: hari, Jumlah: 100000}, nil},
		{"jumlah nol", domain.Kasbon{PekerjaID: 1, Tanggal: hari, Jumlah: 0}, ErrJumlahTidakValid},
		{"jumlah negatif", domain.Kasbon{PekerjaID: 1, Tanggal: hari, Jumlah: -10000}, ErrJumlahTidakValid},
		{"tanggal kosong", domain.Kasbon{PekerjaID: 1, Jumlah: 1000}, ErrTanggalKosong},
		{"pekerja kosong", domain.Kasbon{Tanggal: hari, Jumlah: 1000}, ErrPekerjaTidakAktif},
		{"keterangan 150 karakter", domain.Kasbon{PekerjaID: 1, Tanggal: hari, Jumlah: 1000, Keterangan: strings.Repeat("a", 150)}, nil},
		{"keterangan 151 karakter", domain.Kasbon{PekerjaID: 1, Tanggal: hari, Jumlah: 1000, Keterangan: strings.Repeat("a", 151)}, ErrKeteranganPanjang},
	}

	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			err := validasiKasbon(k.kasbon)
			if !errors.Is(err, k.galat) {
				t.Fatalf("galat = %v, harapan %v", err, k.galat)
			}
		})
	}
}

func TestFormatRupiah(t *testing.T) {
	kasus := map[domain.Rupiah]string{
		0:       "0",
		999:     "999",
		1000:    "1.000",
		193125:  "193.125",
		1500000: "1.500.000",
		-12500:  "-12.500",
	}
	for n, harapan := range kasus {
		if got := formatRupiah(n); got != harapan {
			t.Errorf("formatRupiah(%d) = %q, harapan %q", n, got, harapan)
		}
	}
}

func TestPesanGalatInput(t *testing.T) {
	if PesanGalatInput(ErrJumlahTidakValid) != "Jumlah kasbon harus lebih dari 0." {
		t.Fatal("pesan jumlah tidak sesuai dokumen rancangan")
	}
	if PesanGalatInput(errors.New("koneksi putus")) != "" {
		t.Fatal("galat sistem tidak boleh ditampilkan ke pengguna")
	}
}

// Tes di bawah butuh MySQL berisi data uji, jadi dilewati pada tes otomatis.

func TestCatatKasbonPeriodeTertutup(t *testing.T) {
	t.Skip("butuh MySQL: periode DITUTUP yang mencakup tanggal kasbon harus ditolak (ErrPeriodeTertutup)")
}

func TestKasbonDiprosesDariYangTertua(t *testing.T) {
	t.Skip("butuh MySQL: hasil KasbonDiprosesSampai harus terurut dari tanggal tertua")
}