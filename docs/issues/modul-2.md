# [Modul 2] Jadwal shift dan kasbon

- **PIC:** Edgar
- **Branch:** `modul-2-jadwal-kasbon`
- **Halaman:** 6 `/jadwal` dan `/jadwal/salin`, 7 `/kasbon`, 8 `/kasbon/baru` dan `/kasbon/{id}`
- **Aturan:** 3, 12, 14, 16, 19

## File utama

- `internal/jadwal/repository.go`, `service.go`, `handler.go`
- `internal/kasbon/repository.go`, `service.go`, `handler.go`
- `internal/jadwal/jadwal_test.go`
- `internal/kasbon/kasbon_test.go`
- `web/templates/jadwal.html`
- `web/templates/kasbon_daftar.html`
- `web/templates/kasbon.html`

## Pekerjaan jadwal

- [ ] Implementasi `DaftarShift`, `BuatJadwal`, `UbahJadwal`, `BatalkanJadwal`
- [ ] Implementasi `SalinJadwalMingguLalu` tanpa menimpa sel yang sudah terisi
- [ ] Implementasi `AmbilJadwal(pekerjaID, tanggal)`
- [ ] Tolak hari Minggu
- [ ] Tolak jadwal ganda pada pekerja dan tanggal yang sama
- [ ] Tolak perubahan jadwal yang sudah punya absensi
- [ ] Tolak perubahan jadwal pada periode tertutup
- [ ] Isi grid mingguan dari data nyata

## Pekerjaan kasbon

- [ ] Implementasi `CatatKasbon`, `DaftarKasbon`, `AmbilKasbon`
- [ ] Implementasi `SisaKasbon` dan `KasbonDiprosesSampai`
- [ ] Implementasi `RiwayatPotongan`
- [ ] Validasi jumlah lebih dari 0 dan keterangan maksimal 150 karakter
- [ ] Set `sisa` awal sama dengan `jumlah`
- [ ] Tolak kasbon pada periode tertutup
- [ ] Tampilkan status Berjalan atau Lunas dari kolom `sisa`

## Kontrak yang dijaga

- `kontrak.JadwalProvider` dipakai Modul 3 dan Modul 4.
- `kontrak.KasbonProvider` dipakai Modul 4 dan Modul 5.
- `KasbonDiprosesSampai` harus mengembalikan kasbon terurut dari yang tertua.
- `SisaKasbon` harus konsisten dengan potongan yang sudah tercatat.

## Definition of Done

- [ ] Tidak ada `ErrBelumDiimplementasi` di `internal/jadwal` dan `internal/kasbon`
- [ ] Kedua file tes lulus tanpa skip
- [ ] Semua halaman mengembalikan 200 dengan data nyata
- [ ] Salin jadwal tidak menghasilkan duplikat
- [ ] Pull request direview minimal satu anggota lain
