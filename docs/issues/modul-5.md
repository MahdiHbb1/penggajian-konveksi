# [Modul 5] Periode, penutupan, dan koreksi

- **PIC:** Anya
- **Branch:** `modul-5-periode-koreksi`
- **Halaman:** 13 `/periode` dan `/periode/baru`, 14 `/periode/{id}/tutup`, 15 `/koreksi` dan `/koreksi/baru`
- **Aturan:** 1, 12, 13, 14, 15, 18

## File utama

- `internal/periode/repository.go`
- `internal/periode/service.go`
- `internal/periode/handler.go`
- `internal/periode/periode_test.go`
- `web/templates/periode.html`
- `web/templates/periode_tutup.html`
- `web/templates/koreksi.html`

## Pekerjaan

- [ ] Implementasi `BuatPeriode` mingguan dan bulanan
- [ ] Mingguan wajib mulai Senin dan selesai Sabtu
- [ ] Bulanan wajib mulai tanggal 1 dan selesai akhir bulan
- [ ] Tolak periode yang tumpang tindih
- [ ] Implementasi `DaftarPeriode`, `AmbilPeriode`, `UbahStatusPeriode`
- [ ] Implementasi `PeriksaPeriodeTerkunci`
- [ ] Implementasi `PeriodeSebelumnyaDitutup`
- [ ] Implementasi `TutupPeriode` dalam satu transaksi
- [ ] Kurangi sisa kasbon saat periode ditutup
- [ ] Tandai koreksi yang diterapkan sebagai selesai
- [ ] Implementasi `CatatKoreksi`, `DaftarKoreksi`, `DaftarKoreksiMenunggu`
- [ ] Tolak koreksi untuk periode yang belum ditutup
- [ ] Ganti semua `t.Skip` di `periode_test.go`

## Kontrak yang dijaga

- `kontrak.PeriodeProvider` dipakai Modul 2, 3, 4, dan 5.
- `kontrak.KoreksiProvider` dipakai Modul 4.
- `PeriodeSebelumnyaDitutup` harus benar untuk periode pertama.
- Koreksi hanya boleh masuk sekali ke slip.

## Definition of Done

- [ ] Tidak ada `ErrBelumDiimplementasi` di `internal/periode`
- [ ] Tes `periode_test.go` lulus tanpa skip
- [ ] Penutupan periode tidak bisa dibatalkan
- [ ] Ketiga halaman mengembalikan 200 dengan data nyata
- [ ] Pull request direview minimal satu anggota lain
