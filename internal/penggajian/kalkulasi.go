// Pemilik: Farah | Modul 4 | Aturan: 5, 6, 7, 9, 10, 12
package penggajian

import (
	"sort"
	"time"

	"penggajian-konveksi/internal/config"
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
func HitungHariKerja(mulai, selesai time.Time) int {
	if mulai.IsZero() || selesai.IsZero() {
		return 0
	}
	mulai = tanggalSaja(mulai)
	selesai = tanggalSaja(selesai)
	if selesai.Before(mulai) {
		return 0
	}

	jumlah := 0
	for tanggal := mulai; !tanggal.After(selesai); tanggal = tanggal.AddDate(0, 0, 1) {
		if tanggal.Weekday() != time.Sunday {
			jumlah++
		}
	}
	return jumlah
}

// HitungUpahPerJam menghitung upah per jam sesuai jenis pekerja.
//
// Masukan: nominal upah dan jenis pekerja.
// Hasil: upah per jam, dibulatkan ke bawah.
// Aturan: 7.
// Tabel: tidak ada.
// Memanggil: tidak ada.
func HitungUpahPerJam(nominal domain.Rupiah, jenis domain.JenisPekerja) domain.Rupiah {
	if nominal <= 0 {
		return 0
	}
	switch jenis {
	case domain.PekerjaHarian:
		return nominal / domain.Rupiah(config.JamKerjaPerHari)
	case domain.PekerjaBulanan:
		return nominal / domain.Rupiah(config.HariKerjaPerBulan)
	default:
		return 0
	}
}

// HitungUpahPokokHarian menjumlahkan upah pokok seluruh hari hadir.
//
// Masukan: rincian harian.
// Hasil: total upah pokok.
// Aturan: 9.
// Tabel: tidak ada.
// Memanggil: tidak ada.
func HitungUpahPokokHarian(rincian []domain.SlipGajiDetail) domain.Rupiah {
	var total domain.Rupiah
	for _, r := range rincian {
		if r.UpahPokokHari > 0 {
			total += r.UpahPokokHari
		}
	}
	return total
}

// HitungUpahPokokBulanan menjumlahkan upah proporsional per segmen.
//
// Masukan: daftar segmen upah.
// Hasil: total upah pokok.
// Aturan: 9, 10.
// Tabel: tidak ada.
// Memanggil: tidak ada.
func HitungUpahPokokBulanan(segmen []SegmenUpah) domain.Rupiah {
	var total domain.Rupiah
	for _, s := range segmen {
		if s.Nominal <= 0 || s.HariDijalani <= 0 || s.HariKerjaBulan <= 0 {
			continue
		}
		hari := s.HariDijalani
		if hari > s.HariKerjaBulan {
			hari = s.HariKerjaBulan
		}
		total += s.Nominal * domain.Rupiah(hari) / domain.Rupiah(s.HariKerjaBulan)
	}
	return total
}

// HitungPotonganTerlambat menghitung potongan satu jam bila terlambat.
//
// Masukan: menit terlambat dan upah per jam.
// Hasil: potongan, 0 bila tidak melewati batas.
// Aturan: 5.
// Tabel: tidak ada.
// Memanggil: tidak ada.
func HitungPotonganTerlambat(menitTelat int, upahPerJam domain.Rupiah) domain.Rupiah {
	if menitTelat <= config.MenitTerlambatDipotong || upahPerJam <= 0 {
		return 0
	}
	return upahPerJam
}

// HitungUpahLembur menghitung upah lembur per menit.
//
// Masukan: menit lembur dan upah per jam.
// Hasil: upah lembur, dibulatkan ke bawah per hari.
// Aturan: 6.
// Tabel: tidak ada.
// Memanggil: tidak ada.
func HitungUpahLembur(menitLembur int, upahPerJam domain.Rupiah) domain.Rupiah {
	if menitLembur <= 0 || upahPerJam <= 0 {
		return 0
	}

	const (
		menitLemburPertama = 60
		penyebutPengali    = 2
	)
	menitPertama := menitLembur
	if menitPertama > menitLemburPertama {
		menitPertama = menitLemburPertama
	}
	menitBerikut := menitLembur - menitPertama

	// Pengali 1,5 dan 2 diubah ke bilangan bulat: 3/2 dan 4/2, lalu penyebutnya
	// disamakan menjadi 120 menit. Tidak ada nominal uang yang menjadi float64.
	pengaliPertama := domain.Rupiah(config.PengaliLemburPertama * penyebutPengali)
	pengaliBerikut := domain.Rupiah(config.PengaliLemburBerikutnya * penyebutPengali)
	pembilang := domain.Rupiah(menitPertama)*pengaliPertama*upahPerJam +
		domain.Rupiah(menitBerikut)*pengaliBerikut*upahPerJam
	return pembilang / domain.Rupiah(menitLemburPertama*penyebutPengali)
}

// HitungGajiBruto menghitung gaji bruto.
//
// Masukan: upah pokok, lembur, potongan terlambat, koreksi.
// Hasil: gaji bruto.
// Aturan: 12.
// Tabel: tidak ada.
// Memanggil: tidak ada.
func HitungGajiBruto(pokok, lembur, potonganTelat, koreksi domain.Rupiah) domain.Rupiah {
	return pokok + lembur - potonganTelat + koreksi
}

// AlokasiPotonganKasbon membagi potongan kasbon dari yang tertua.
//
// Masukan: gaji bruto dan daftar kasbon.
// Hasil: daftar potongan dan total potongan.
// Aturan: 12.
// Tabel: tidak ada.
// Memanggil: tidak ada.
func AlokasiPotonganKasbon(bruto domain.Rupiah, daftar []domain.Kasbon) ([]domain.PotonganKasbon, domain.Rupiah) {
	if bruto <= 0 || len(daftar) == 0 {
		return nil, 0
	}

	batas := bruto * domain.Rupiah(config.PersenBatasKasbon) / 100
	if batas <= 0 {
		return nil, 0
	}

	// Salin agar urutan pemanggil tidak berubah, lalu pastikan kasbon tertua
	// dipotong lebih dulu.
	urut := append([]domain.Kasbon(nil), daftar...)
	sort.SliceStable(urut, func(i, j int) bool {
		if urut[i].Tanggal.Equal(urut[j].Tanggal) {
			return urut[i].ID < urut[j].ID
		}
		return urut[i].Tanggal.Before(urut[j].Tanggal)
	})

	var (
		potongan []domain.PotonganKasbon
		total    domain.Rupiah
	)
	for _, k := range urut {
		if total >= batas {
			break
		}
		if k.ID <= 0 || k.Sisa <= 0 {
			continue
		}
		jumlah := k.Sisa
		if sisaBatas := batas - total; jumlah > sisaBatas {
			jumlah = sisaBatas
		}
		if jumlah <= 0 {
			break
		}
		potongan = append(potongan, domain.PotonganKasbon{
			KasbonID: k.ID,
			Jumlah:   jumlah,
		})
		total += jumlah
	}
	return potongan, total
}

func tanggalSaja(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	tahun, bulan, hari := t.Date()
	return time.Date(tahun, bulan, hari, 0, 0, 0, 0, t.Location())
}
