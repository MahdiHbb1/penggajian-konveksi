// Pemilik: Farah | Modul 4 | Aturan: 5, 6, 7, 8, 9, 10, 12, 17, 18
package penggajian

import (
	"context"

	"penggajian-konveksi/internal/domain"
	"penggajian-konveksi/internal/kontrak"
)

// Prasyarat adalah hasil pemeriksaan sebelum periode dihitung.
type Prasyarat struct {
	StatusPeriodeOK     bool
	PeriodeSebelumnyaOK bool
	AbsensiLengkap      bool
	JumlahAbsensiKurang int
	Pesan               string
}

// Service menghitung gaji satu periode.
type Service struct {
	repo    *Repository
	pekerja kontrak.PekerjaProvider
	jadwal  kontrak.JadwalProvider
	kasbon  kontrak.KasbonProvider
	absensi kontrak.AbsensiProvider
	periode kontrak.PeriodeProvider
	koreksi kontrak.KoreksiProvider
}

// NewService membuat service penggajian dengan seluruh dependensi kontrak.
func NewService(
	repo *Repository,
	pekerja kontrak.PekerjaProvider,
	jadwal kontrak.JadwalProvider,
	kasbon kontrak.KasbonProvider,
	absensi kontrak.AbsensiProvider,
	periode kontrak.PeriodeProvider,
	koreksi kontrak.KoreksiProvider,
) *Service {
	return &Service{
		repo:    repo,
		pekerja: pekerja,
		jadwal:  jadwal,
		kasbon:  kasbon,
		absensi: absensi,
		periode: periode,
		koreksi: koreksi,
	}
}

// HitungGajiPeriode menghitung dan menyimpan gaji seluruh pekerja pada periode.
//
// Masukan: context dan ID periode.
// Hasil: error.
// Aturan: 5, 6, 7, 8, 9, 12, 17, 18.
// Tabel: periode_gaji, pekerja, riwayat_upah, jadwal_shift, absensi, kasbon, koreksi_gaji, slip_gaji, slip_gaji_detail, potongan_kasbon.
// Memanggil: PeriksaPrasyaratHitung, PekerjaProvider, AbsensiProvider, KasbonProvider, KoreksiProvider, Repository.SimpanSlipDanRincian, PeriodeProvider.UbahStatusPeriode.
//
// TODO(Farah):
//  1. Periksa status periode dan seluruh prasyarat.
//  2. Ambil pekerja aktif sesuai tipe periode.
//  3. Kumpulkan absensi, jadwal, upah berlaku, kasbon, dan koreksi per pekerja.
//  4. Hitung upah pokok harian atau bulanan per segmen upah.
//  5. Hitung lembur, potongan terlambat, dan koreksi.
//  6. Hitung bruto, batas kasbon 30%, alokasi kasbon tertua, dan gaji bersih.
//  7. Simpan slip, rincian, dan potongan dalam satu transaksi.
//  8. Ubah status periode menjadi Dihitung.
//  9. Saat hitung ulang, timpa slip lama dan jangan hitung koreksi dua kali.
func (s *Service) HitungGajiPeriode(ctx context.Context, periodeID int64) error {
	return domain.ErrBelumDiimplementasi
}

// PeriksaPrasyaratHitung memeriksa syarat hitung gaji.
//
// Masukan: context dan ID periode.
// Hasil: Prasyarat.
// Aturan: 2, 13.
// Tabel: periode_gaji, absensi, jadwal_shift.
// Memanggil: PeriodeProvider.AmbilPeriode, PeriodeProvider.PeriodeSebelumnyaDitutup, AbsensiProvider.AdaAbsensiBelumLengkap.
//
// TODO(Farah):
//  1. Periode harus Terbuka atau Dihitung.
//  2. Periode sebelumnya dengan tipe sama harus Ditutup; periode pertama lolos.
//  3. Tidak boleh ada absensi tanpa jam keluar.
func (s *Service) PeriksaPrasyaratHitung(ctx context.Context, periodeID int64) (Prasyarat, error) {
	return Prasyarat{}, domain.ErrBelumDiimplementasi
}

// SimpanSlipDanRincian menyimpan hasil hitung.
//
// Masukan: context, slip, rincian, potongan.
// Hasil: error.
// Aturan: 13, 18.
// Tabel: slip_gaji, slip_gaji_detail, potongan_kasbon.
// Memanggil: Repository.SimpanSlipDanRincian.
func (s *Service) SimpanSlipDanRincian(ctx context.Context, slip domain.SlipGaji, rincian []domain.SlipGajiDetail, potongan []domain.PotonganKasbon) error {
	return domain.ErrBelumDiimplementasi
}

// AmbilRincianGaji mengambil slip dan rincian satu pekerja.
//
// Masukan: context, ID periode, ID pekerja.
// Hasil: slip dan rincian harian.
// Aturan: 17.
// Tabel: slip_gaji, slip_gaji_detail.
// Memanggil: Repository.AmbilSlip.
func (s *Service) AmbilRincianGaji(ctx context.Context, periodeID, pekerjaID int64) (domain.SlipGaji, []domain.SlipGajiDetail, error) {
	return domain.SlipGaji{}, nil, domain.ErrBelumDiimplementasi
}
