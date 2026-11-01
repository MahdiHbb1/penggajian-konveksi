# [Modul 3] Absensi

- **PIC:** Zahra
- **Branch:** `modul-3-absensi`
- **Halaman:** 9 `/absensi`, 10 `/absensi/belum-lengkap`
- **Aturan:** 2, 3, 4, 5, 6, 14

## File utama

- `internal/absensi/repository.go`
- `internal/absensi/service.go`
- `internal/absensi/handler.go`
- `internal/absensi/absensi_test.go`
- `web/templates/absensi.html`
- `web/templates/absensi_belum_lengkap.html`

## Pekerjaan

- [ ] Implementasi `CatatAbsensi`, `UbahAbsensi`, `DaftarAbsensi`
- [ ] Implementasi `DaftarAbsensiBelumLengkap`, `AbsensiDalamPeriode`
- [ ] Implementasi `AdaAbsensiBelumLengkap`
- [ ] Implementasi `HitungMenitTerlambat` dan `HitungMenitLembur`
- [ ] Tolak absensi ganda untuk satu jadwal
- [ ] Validasi jam keluar setelah jam masuk, termasuk shift yang lewat tengah malam
- [ ] Tolak absensi pada periode tertutup
- [ ] Tampilkan pratinjau menit terlambat dan lembur di halaman input
- [ ] Ganti semua `t.Skip` di `absensi_test.go`

## Kontrak yang dijaga

- `kontrak.AbsensiProvider` dipakai Modul 4.
- `AbsensiDalamPeriode` hanya mengembalikan absensi yang sudah lengkap.
- `AdaAbsensiBelumLengkap` dipakai sebagai prasyarat hitung gaji.
- Format `time.Time` harus disepakati dengan Farah sebelum integrasi.

## Definition of Done

- [ ] Tidak ada `ErrBelumDiimplementasi` di `internal/absensi`
- [ ] Tes `absensi_test.go` lulus tanpa skip
- [ ] Kedua halaman mengembalikan 200 dengan data nyata
- [ ] Kasus tepat 15 menit dan 61 menit lembur terbukti benar
- [ ] Pull request direview minimal satu anggota lain
