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
- `internal/periode/handler_test.go`
- `web/templates/periode.html`
- `web/templates/periode_tutup.html`
- `web/templates/koreksi.html`

## Pekerjaan

- [x] Implementasi `BuatPeriode` mingguan dan bulanan
- [x] Mingguan wajib mulai Senin dan selesai Sabtu
- [x] Bulanan wajib mulai tanggal 1 dan selesai akhir bulan
- [x] Tolak periode bertipe sama yang tumpang tindih
- [x] Implementasi `DaftarPeriode`, `AmbilPeriode`, `UbahStatusPeriode`
- [x] Implementasi `PeriksaPeriodeTerkunci`
- [x] Implementasi `PeriodeSebelumnyaDitutup`
- [x] Implementasi `TutupPeriode` dalam satu transaksi
- [x] Kurangi sisa kasbon saat periode ditutup
- [x] Tandai koreksi yang diterapkan sebagai selesai
- [x] Implementasi `CatatKoreksi`, `DaftarKoreksi`, `DaftarKoreksiMenunggu`
- [x] Tolak koreksi untuk periode yang belum ditutup
- [x] Ganti semua `t.Skip` di `periode_test.go`
- [x] Aksi hitung mengarahkan ke `/penggajian?periode_id={id}`

## Keputusan aturan

- Tumpang tindih diperiksa per tipe periode. `MINGGUAN` dan `BULANAN` boleh berjalan
  paralel karena dipakai untuk kelompok pekerja berbeda (harian dan bulanan), sedangkan
  dua periode dengan tipe yang sama tetap tidak boleh beririsan.
- Aksi hitung pada halaman periode hanya meneruskan ke
  `/penggajian?periode_id={id}`. Perhitungan gaji tetap menjadi tanggung jawab Modul 4.

## Kontrak yang dijaga

- `kontrak.PeriodeProvider` dipakai Modul 2, 3, 4, dan 5.
- `kontrak.KoreksiProvider` dipakai Modul 4.
- `PeriodeSebelumnyaDitutup` harus benar untuk periode pertama.
- Koreksi hanya boleh masuk sekali ke slip.

Service sekarang memakai penegasan kontrak langsung:

```go
var _ kontrak.PeriodeProvider = (*Service)(nil)
var _ kontrak.KoreksiProvider = (*Service)(nil)
```

## Definition of Done

- [x] Tidak ada `ErrBelumDiimplementasi` di `internal/periode`
- [x] Tes `periode_test.go` lulus tanpa skip
- [x] Penutupan periode tidak bisa dibatalkan
- [x] Ketiga halaman mengembalikan 200 dengan data nyata
- [ ] Pull request direview minimal satu anggota lain

Halaman diuji lewat `handler_test.go` dengan repository yang diarahkan ke database tiruan
(`go-sqlmock`) dan template asli dari `web/templates`. Verifikasi akhir dengan MySQL
tetap perlu dilakukan sekali saat integrasi sebelum demo.

## Verifikasi

```text
gofmt -l internal/periode        -> bersih
go build ./...                   -> lolos
go vet ./...                     -> lolos
go test ./...                    -> lolos
go test ./internal/periode/...   -> lolos, tanpa t.Skip
```

Catatan integrasi:

- `TutupPeriode` menolak periode berstatus `DITUTUP`, sehingga penutupan tidak dapat dibatalkan.
- Penutupan juga menolak bila periode sebelumnya dengan tipe yang sama belum ditutup.
- Periode mingguan dan bulanan dapat berjalan bersamaan; pengecekan tumpang tindih
  memakai filter `tipe`.
- Sisa kasbon dikurangi dari `potongan_kasbon` aktual pada slip, bukan dari perkiraan.
- Koreksi yang masih menunggu untuk pekerja pada periode tersebut ditandai `periode_terapan_id`
  di dalam transaksi penutupan.
