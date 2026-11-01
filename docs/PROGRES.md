# Rekap Progres Modul

Rekap ini untuk laporan kampus. Sumber kebenaran harian tetap GitHub Projects.

Terakhir diperbarui: 1 November 2026

| Modul | PIC | Branch | Issue | PR | Status | Tes | Halaman |
|---|---|---|---|---|---|---|---|
| 1 Pekerja dan riwayat upah | Mahdi | `modul-1-pekerja` | belum dibuat | belum ada | Backlog | 5 subtes skip | 3, 4, 5 |
| 2 Jadwal shift dan kasbon | Edgar | `modul-2-jadwal-kasbon` | belum dibuat | belum ada | Backlog | 9 subtes skip | 6, 7, 8 |
| 3 Absensi | Zahra | `modul-3-absensi` | belum dibuat | belum ada | Backlog | 8 subtes skip | 9, 10 |
| 4 Perhitungan gaji | Farah | `modul-4-penggajian` | belum dibuat | belum ada | Backlog | 17 subtes skip | 11, 12 |
| 5 Periode, penutupan, koreksi | Anya | `modul-5-periode-koreksi` | belum dibuat | belum ada | Backlog | 5 subtes skip | 13, 14, 15 |
| 6 Slip, rekap, autentikasi, UI | Fadli | `modul-6-laporan-antarmuka` | belum dibuat | belum ada | Backlog | 7 subtes skip | 1, 2, 16, 17 |

Jumlah subtes skip dapat berubah saat tes mulai diisi. Perbarui angka ini setelah `go test ./... -v`.

## Definisi status

| Status | Arti |
|---|---|
| Backlog | Belum dikerjakan, desain masih dibahas |
| Ready | Siap dikerjakan, file dan kontrak sudah jelas |
| In progress | Sedang dikerjakan, sudah ada commit lokal |
| In review | Pull request ke `main` sudah dibuka dan menunggu review |
| Done | Pull request sudah di-merge, tes lulus, halaman sudah dicek |

## Cara memperbarui

1. Pindahkan kartu issue di GitHub Projects.
2. Isi kolom `Issue` dengan nomor issue, misalnya `#12`.
3. Isi kolom `PR` dengan nomor pull request, misalnya `#18`.
4. Perbarui kolom `Status` dan `Tes` di file ini setiap akhir pekan.
5. Tulis hambatan di komentar issue, bukan hanya di grup chat.

## Checklist koordinasi lintas modul

- [ ] Format tanggal dan timezone disepakati sebelum Modul 3 dan 4 terintegrasi.
- [ ] Kontrak `PekerjaProvider` tidak berubah tanpa persetujuan Mahdi dan Farah.
- [ ] Kontrak `AbsensiProvider` tidak berubah tanpa persetujuan Zahra dan Farah.
- [ ] Kontrak `PeriodeProvider` tidak berubah tanpa persetujuan Anya dan Farah.
- [ ] Penutupan periode diuji bersama Modul 4 sebelum demo.
- [ ] Login sungguhan menggantikan bypass sebelum demo.
