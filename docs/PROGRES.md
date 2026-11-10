# Rekap Progres Modul

Rekap ini untuk laporan kampus. Sumber kebenaran harian tetap GitHub Projects.

Terakhir diperbarui: 11 November 2026

Repo: <https://github.com/MahdiHbb1/penggajian-konveksi> (privat)
Project: <https://github.com/users/MahdiHbb1/projects/2>

| Modul | PIC | Branch | Issue | PR | Status | Tes | Halaman |
|---|---|---|---|---|---|---|---|
| 1 Pekerja dan riwayat upah | Mahdi | `modul-1-pekerja` | [#1](https://github.com/MahdiHbb1/penggajian-konveksi/issues/1) | belum ada | In Progress | 33 tes lulus, 0 skip | 3, 4, 5 |
| 2 Jadwal shift dan kasbon | Edgar | `modul-2-jadwal-kasbon` | [#2](https://github.com/MahdiHbb1/penggajian-konveksi/issues/2) | belum ada | Todo | 9 subtes skip | 6, 7, 8 |
| 3 Absensi | Zahra | `modul-3-absensi` | [#3](https://github.com/MahdiHbb1/penggajian-konveksi/issues/3) | belum ada | Todo | 8 subtes skip | 9, 10 |
| 4 Perhitungan gaji | Farah | `modul-4-penggajian` | [#4](https://github.com/MahdiHbb1/penggajian-konveksi/issues/4) | [#9](https://github.com/MahdiHbb1/penggajian-konveksi/pull/9) | In Progress | 24 tes lulus, 0 skip | 11, 12 |
| 5 Periode, penutupan, koreksi | Anya | `modul-5-periode-koreksi` | [#5](https://github.com/MahdiHbb1/penggajian-konveksi/issues/5) | belum ada | Todo | 5 subtes skip | 13, 14, 15 |
| 6 Slip, rekap, autentikasi, UI | Fadli | `modul-6-laporan-antarmuka` | [#6](https://github.com/MahdiHbb1/penggajian-konveksi/issues/6) | belum ada | Todo | 7 subtes skip | 1, 2, 16, 17 |

Jumlah subtes skip dapat berubah saat tes mulai diisi. Perbarui angka ini setelah `go test ./... -v`.

## Definisi status

Project memakai tiga opsi Status bawaan GitHub.

| Status | Arti |
|---|---|
| Todo | Belum dikerjakan atau baru ditugaskan |
| In Progress | Sedang dikerjakan, sudah ada commit lokal |
| Done | Pull request sudah di-merge, tes lulus, halaman sudah dicek |

Review tidak punya status sendiri. Saat pull request dibuka, issue tetap `In Progress` dan PR-nya ditautkan lewat kolom **Linked pull requests** di Project. Setelah PR di-merge dan verifikasi lulus, baru pindahkan ke `Done`.

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
