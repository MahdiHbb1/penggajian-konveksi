# Kepemilikan File dan Tugas

Satu file atau folder hanya punya satu pemilik review. Tabel ini adalah default; bila kelompok mengubah pembagian, ubah tabel ini sebelum coding.

## Ringkasan

| Anggota | GitHub | Modul | Branch | Folder utama | Halaman |
|---|---|---|---|---|---|
| Mahdi | `@MahdiHbb1` | 1. Pekerja dan riwayat upah | `modul-1-pekerja` | `internal/pekerja`, `internal/config`, `internal/database`, `internal/domain`, `internal/kontrak`, `database`, `cmd/server` | 3, 4, 5 |
| Edgar | `@edgar8989` | 2. Jadwal shift dan kasbon | `modul-2-jadwal-kasbon` | `internal/jadwal`, `internal/kasbon` | 6, 7, 8 |
| Zahra | `@Raatsum` | 3. Absensi | `modul-3-absensi` | `internal/absensi` | 9, 10 |
| Farah | `@ffaraahw` | 4. Perhitungan gaji | `modul-4-penggajian` | `internal/penggajian` | 11, 12 |
| Anya | `@anyaa-k` | 5. Periode, penutupan, koreksi | `modul-5-periode-koreksi` | `internal/periode` | 13, 14, 15 |
| Fadli | `@fadlinugraha15` | 6. Slip, rekap, antarmuka | `modul-6-laporan-antarmuka` | `internal/laporan`, `internal/auth`, `internal/web`, `web` | 1, 2, 16, 17 |

## Mahdi

- [ ] `internal/pekerja/handler.go`, `service.go`, `repository.go`
- [ ] `internal/domain/tipe.go`
- [ ] `internal/kontrak/kontrak.go`
- [ ] `internal/config/config.go`
- [ ] `internal/database/koneksi.go`
- [ ] `database/schema.sql`, `database/seed.sql`
- [ ] `docs/erd.dbml`
- [ ] `cmd/server/main.go`
- [ ] `internal/pekerja/pekerja_test.go`
- [ ] Halaman 3 `/pekerja`, 4 `/pekerja/baru` dan `/pekerja/{id}/ubah`, 5 `/pekerja/{id}/upah`

Yang BUKAN tugas Mahdi: logika hitung gaji, jadwal, absensi, kasbon, periode, slip, autentikasi.

## Edgar

- [ ] `internal/jadwal/handler.go`, `service.go`, `repository.go`
- [ ] `internal/kasbon/handler.go`, `service.go`, `repository.go`
- [ ] `internal/jadwal/jadwal_test.go`
- [ ] `internal/kasbon/kasbon_test.go`
- [ ] Halaman 6 `/jadwal` dan `/jadwal/salin`, 7 `/kasbon`, 8 `/kasbon/baru` dan `/kasbon/{id}`

Yang BUKAN tugas Edgar: upah, absensi, hitung gaji, periode, slip, autentikasi.

## Zahra

- [ ] `internal/absensi/handler.go`, `service.go`, `repository.go`
- [ ] `internal/absensi/absensi_test.go`
- [ ] Halaman 9 `/absensi`, 10 `/absensi/belum-lengkap`

Yang BUKAN tugas Zahra: jadwal, kasbon, upah, perhitungan rupiah, periode, slip, autentikasi.

## Farah

- [ ] `internal/penggajian/handler.go`, `service.go`, `repository.go`, `kalkulasi.go`
- [ ] `internal/penggajian/kalkulasi_test.go`
- [ ] Halaman 11 `/penggajian`, 12 `/penggajian/{periodeID}/rincian/{pekerjaID}`

Yang BUKAN tugas Farah: CRUD pekerja, jadwal, kasbon, absensi, membuat periode, menutup periode, autentikasi.

Beban modul ini paling berat. Jangan memindahkan tugas tanpa persetujuan manusia.

## Anya

- [ ] `internal/periode/handler.go`, `service.go`, `repository.go`
- [ ] `internal/periode/periode_test.go`
- [ ] Halaman 13 `/periode` dan `/periode/baru`, 14 `/periode/{id}/tutup`, 15 `/koreksi` dan `/koreksi/baru`

Yang BUKAN tugas Anya: menghitung gaji, mengisi absensi, membuat jadwal, mencatat kasbon, autentikasi.

## Fadli

- [ ] `internal/laporan/handler.go`, `service.go`, `repository.go`
- [ ] `internal/auth/handler.go`, `service.go`, `middleware.go`
- [ ] `internal/web/router.go`, `render.go`
- [ ] `web/templates/*`, `web/static/css/style.css`
- [ ] `internal/auth/auth_test.go`, `internal/laporan/laporan_test.go`
- [ ] Halaman 1 `/login`, 2 `/dashboard`, 16 `/slip/{periodeID}/{pekerjaID}`, 17 `/rekap/{periodeID}`

Yang BUKAN tugas Fadli: aturan bisnis enam modul lain dan perhitungan gaji.

## File bersama

| File | Pemilik review |
|---|---|
| `cmd/server/main.go` | Mahdi |
| `internal/domain/tipe.go` | Mahdi |
| `internal/kontrak/kontrak.go` | Mahdi |
| `internal/config/config.go` | Mahdi |
| `internal/database/koneksi.go` | Mahdi |
| `internal/web/router.go`, `render.go` | Fadli |
| `README.md` | Mahdi |
| `docs/LAPORAN_AGENT.md` | Mahdi |

## Aturan review

- Setiap file Go diawali `// Pemilik: <nama> | Modul <n> | Aturan: <nomor>`.
- Setiap template diawali `<!-- Pemilik: <nama> | Modul <n> -->`.
- Perubahan file bersama harus disetujui pemilik review.
- Jangan menghapus placeholder milik modul lain saat mengisi fungsi sendiri.
