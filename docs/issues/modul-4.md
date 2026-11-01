# [Modul 4] Perhitungan gaji

- **PIC:** Farah
- **Branch:** `modul-4-penggajian`
- **Halaman:** 11 `/penggajian`, 12 `/penggajian/{periodeID}/rincian/{pekerjaID}`
- **Aturan:** 5, 6, 7, 8, 9, 10, 12, 17, 18

## File utama

- `internal/penggajian/repository.go`
- `internal/penggajian/service.go`
- `internal/penggajian/kalkulasi.go`
- `internal/penggajian/handler.go`
- `internal/penggajian/kalkulasi_test.go`
- `web/templates/penggajian.html`
- `web/templates/penggajian_rincian.html`

## Pekerjaan

- [ ] Implementasi `HitungGajiPeriode` dari prasyarat sampai simpan transaksi
- [ ] Implementasi `PeriksaPrasyaratHitung`
- [ ] Implementasi `HitungHariKerja`
- [ ] Implementasi `HitungUpahPerJam`, `HitungUpahPokokHarian`, `HitungUpahPokokBulanan`
- [ ] Implementasi `HitungPotonganTerlambat`
- [ ] Implementasi `HitungUpahLembur` dengan pengali 1,5 dan 2
- [ ] Implementasi `HitungGajiBruto`
- [ ] Implementasi `AlokasiPotonganKasbon` dengan batas 30 persen
- [ ] Implementasi `SimpanSlipDanRincian` dan `AmbilRincianGaji`
- [ ] Tangani hitung ulang tanpa menghitung koreksi dua kali
- [ ] Ganti semua `t.Skip` di `kalkulasi_test.go`

## Nilai acuan

- Contoh A: upah harian 100.000, hadir 6 hari, lembur 90 menit, kasbon 300.000.
- Contoh B: gaji bulanan 3.460.000, 13 dari 25 hari kerja, lembur 180 menit, kasbon 1.500.000.
- Contoh C: upah berubah di tengah periode, hitung proporsional per segmen.
- Bulatkan ke bawah per hari atau per segmen, jangan pakai `float64` untuk uang.

## Kontrak yang dijaga

- Memakai `PekerjaProvider`, `JadwalProvider`, `KasbonProvider`, `AbsensiProvider`, `PeriodeProvider`, dan `KoreksiProvider`.
- Jangan menaruh query SQL di `kalkulasi.go`.
- Sepakati format tanggal dan timezone dengan Zahra dan Anya.

## Definition of Done

- [ ] Tidak ada `ErrBelumDiimplementasi` di `internal/penggajian`
- [ ] Seluruh nilai contoh A, B, dan C terbukti di tes
- [ ] Kasus tepi 15 menit, 16 menit, 61 menit, dan lewat tengah malam lulus
- [ ] Slip tersimpan dalam satu transaksi
- [ ] Kedua halaman mengembalikan 200 dengan data nyata
- [ ] Pull request direview minimal satu anggota lain
