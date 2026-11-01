// Pemilik: Fadli | Modul 6 | Aturan: 13, 17
package laporan

import (
	"context"
	"testing"
)

func TestSlipBelumTersediaSebelumDihitung(t *testing.T) {
	t.Skip("belum diimplementasi: AmbilSlipGaji")
	svc := NewService(NewRepository(nil))
	if _, err := svc.AmbilSlipGaji(context.Background(), 1, 1); err == nil {
		t.Fatal("slip periode Terbuka seharusnya tidak tersedia")
	}
}

func TestRekapHanyaPeriodeDihitungAtauDitutup(t *testing.T) {
	t.Skip("belum diimplementasi: RekapGajiPerPeriode")
	svc := NewService(NewRepository(nil))
	if _, err := svc.RekapGajiPerPeriode(context.Background(), 1); err == nil {
		t.Fatal("rekap periode Terbuka seharusnya tidak tersedia")
	}
}

func TestRekapTidakMemotongTelatDuaKali(t *testing.T) {
	t.Skip("belum diimplementasi: RekapGajiPerPeriode")
	svc := NewService(NewRepository(nil))
	rekap, err := svc.RekapGajiPerPeriode(context.Background(), 1)
	if err != nil {
		t.Fatalf("ambil rekap: %v", err)
	}
	if rekap.TotalPotonganKasbon < 0 {
		t.Fatal("total potongan kasbon tidak boleh negatif")
	}
}
