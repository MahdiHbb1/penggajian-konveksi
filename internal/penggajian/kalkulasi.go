// Pemilik: Farah | Modul 4 | Aturan: 5, 6, 7, 9, 10, 12
package penggajian

import (
	"time"

	"penggajian-konveksi/internal/domain"
)

// SegmenUpah adalah satu potongan masa berlaku upah untuk pekerja bulanan.
type SegmenUpah struct {
	Nominal        domain.Rupiah
	HariDijalani   int
	HariKerjaBulan int
}

// HitungHariKerja menghitung hari Senin sampai Sabtu pada rentang tanggal.
//
// Masukan: tanggal mulai dan selesai.
// Hasil: jumlah hari kerja.
// Aturan: 1, 9.
// Tabel: tidak ada.
// Memanggil: tidak ada.
//
// TODO(Farah):
//  1. Iterasi tanggal mulai sampai selesai.
//  2. Lewati hari Minggu.
func HitungHariKerja(mulai, selesai time.Time) int {
	return 0
}

// HitungUpahPerJam menghitung upah per jam sesuai jenis pekerja.
//
// Masukan: nominal upah dan jenis pekerja.
// Hasil: upah per jam, dibulatkan ke bawah.
// Aturan: 7.
// Tabel: tidak ada.
// Memanggil: tidak ada.
//
// TODO(Farah):
//  1. Harian: nominal dibagi 8.
//  2. Bulanan: nominal dibagi 173.
func HitungUpahPerJam(nominal domain.Rupiah, jenis domain.JenisPekerja) domain.Rupiah {
	return 0
}

// HitungUpahPokokHarian menjumlahkan upah pokok seluruh hari hadir.
//
// Masukan: rincian harian.
// Hasil: total upah pokok.
// Aturan: 9.
// Tabel: tidak ada.
// Memanggil: tidak ada.
//
// TODO(Farah):
//  1. Jumlahkan UpahPokokHari pada setiap rincian.
func HitungUpahPokokHarian(rincian []domain.SlipGajiDetail) domain.Rupiah {
	return 0
}

// HitungUpahPokokBulanan menjumlahkan upah proporsional per segmen.
//
// Masukan: daftar segmen upah.
// Hasil: total upah pokok.
// Aturan: 9, 10.
// Tabel: tidak ada.
// Memanggil: tidak ada.
//
// TODO(Farah):
//  1. Untuk setiap segmen, hitung nominal x hari dijalani / hari kerja bulan.
//  2. Bulatkan ke bawah per segmen, lalu jumlahkan.
//  3. Pekerja bulanan tetap mendapat upah pokok meski tidak masuk.
func HitungUpahPokokBulanan(segmen []SegmenUpah) domain.Rupiah {
	return 0
}

// HitungPotonganTerlambat menghitung potongan satu jam bila terlambat.
//
// Masukan: menit terlambat dan upah per jam.
// Hasil: potongan, 0 bila tidak melewati batas.
// Aturan: 5.
// Tabel: tidak ada.
// Memanggil: tidak ada.
//
// TODO(Farah):
//  1. Terlambat lebih dari 15 menit dipotong satu kali upah per jam.
//  2. Tepat 15 menit tidak dipotong.
func HitungPotonganTerlambat(menitTelat int, upahPerJam domain.Rupiah) domain.Rupiah {
	return 0
}

// HitungUpahLembur menghitung upah lembur per menit.
//
// Masukan: menit lembur dan upah per jam.
// Hasil: upah lembur, dibulatkan ke bawah per hari.
// Aturan: 6.
// Tabel: tidak ada.
// Memanggil: tidak ada.
//
// TODO(Farah):
//  1. 60 menit pertama memakai pengali 1,5.
//  2. Menit berikutnya memakai pengali 2.
//  3. Hitung per menit lalu bulatkan ke bawah.
func HitungUpahLembur(menitLembur int, upahPerJam domain.Rupiah) domain.Rupiah {
	return 0
}

// HitungGajiBruto menghitung gaji bruto.
//
// Masukan: upah pokok, lembur, potongan terlambat, koreksi.
// Hasil: gaji bruto.
// Aturan: 12.
// Tabel: tidak ada.
// Memanggil: tidak ada.
//
// TODO(Farah):
//  1. Jumlahkan pokok dan lembur.
//  2. Kurangi potongan terlambat.
//  3. Tambahkan koreksi yang boleh negatif.
func HitungGajiBruto(pokok, lembur, potonganTelat, koreksi domain.Rupiah) domain.Rupiah {
	return 0
}

// AlokasiPotonganKasbon membagi potongan kasbon dari yang tertua.
//
// Masukan: gaji bruto dan daftar kasbon.
// Hasil: daftar potongan dan total potongan.
// Aturan: 12.
// Tabel: tidak ada.
// Memanggil: tidak ada.
//
// TODO(Farah):
//  1. Hitung batas 30% dari bruto.
//  2. Ambil minimum antara total sisa dan batas.
//  3. Alokasikan berurutan dari kasbon tertua.
func AlokasiPotonganKasbon(bruto domain.Rupiah, daftar []domain.Kasbon) ([]domain.PotonganKasbon, domain.Rupiah) {
	return nil, 0
}
