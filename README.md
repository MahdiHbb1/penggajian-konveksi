# Sistem Penggajian Konveksi Mitra Jaya

Kerangka proyek akhir mata kuliah Pemrograman Web. Aplikasi Go ini menyiapkan struktur enam modul, 17 halaman, 12 tabel, dan kontrak antarmodul. Badan fungsi bisnis masih placeholder dan akan diisi oleh pemilik modul masing-masing.

## Stack

- Go 1.22 atau lebih baru
- MySQL 8 (opsional saat pengembangan kerangka)
- `html/template` dan CSS sederhana
- Tanpa framework frontend

## Menjalankan tanpa MySQL

Kerangka harus bisa diperiksa tanpa database.

```bash
cp .env.example .env
go run ./cmd/server
```

Buka `http://localhost:8080`. Aplikasi akan mencatat peringatan bila MySQL tidak tersedia, tetapi tetap berjalan. Semua halaman memakai data kosong.

## Menyiapkan MySQL

```bash
mysql -u root -p < database/schema.sql
mysql -u root -p < database/seed.sql
```

`seed.sql` hanya berisi dua shift: Pagi 07.00-15.00 dan Siang 15.00-23.00. Tidak ada akun contoh.

Setelah Modul 6 mengimplementasikan autentikasi, buat akun admin dengan hash bcrypt. Jangan menulis password asli di repo, README, atau seed. Contoh bentuk data:

```sql
INSERT INTO pengguna (username, password_hash, role)
VALUES ('admin', '<hash-bcrypt-dari-password-lokal>', 'ADMIN');
```

## Konfigurasi

| Variabel | Default | Keterangan |
|---|---|---|
| `PORT` | `8080` | Port HTTP |
| `DB_DSN` | `root:@tcp(127.0.0.1:3306)/penggajian_konveksi?parseTime=true&loc=Local` | DSN MySQL |
| `SESSION_SECRET` | kosong | Diisi lokal, jangan di-commit |
| `TEMPLATE_DIR` | `web/templates` | Folder template |
| `STATIC_DIR` | `web/static` | Folder aset statis |

## Struktur

```text
cmd/server/              titik masuk aplikasi
internal/config/         konstanta aturan dan env
internal/database/       koneksi MySQL yang tidak fatal
internal/domain/         tipe domain dan error bersama
internal/kontrak/        interface antarmodul
internal/pekerja/        Modul 1
internal/jadwal/         Modul 2, jadwal shift
internal/kasbon/         Modul 2, kasbon
internal/absensi/        Modul 3
internal/penggajian/     Modul 4
internal/periode/        Modul 5
internal/laporan/        Modul 6, slip dan rekap
internal/auth/           Modul 6, login dan middleware
internal/web/            router dan renderer
web/templates/           layout dan 17 halaman
web/static/css/          style.css
database/                schema.sql dan seed.sql
docs/                    ERD, kepemilikan, laporan, panduan GitHub
.github/                 template issue dan pull request
scripts/                 pembuat label dan issue GitHub
```

## Tracking dan GitHub

- Panduan repo, kolaborator, dan GitHub Projects: `docs/GITHUB_PROJECT.md`.
- Rekap progres untuk laporan: `docs/PROGRES.md`.
- Pembuat label dan issue: `scripts/siapkan-github.ps1`.
- Satu issue per modul, satu branch per modul, dan pull request ke `main`.

## Pemeriksaan

```bash
gofmt -l .
go build ./...
go vet ./...
go test ./...
```

Tes kerangka memakai `t.Skip` karena fungsi belum diisi. Nilai harapan tes sudah ditulis dan mengacu pada contoh A/B/C di dokumen rancangan.

## Mode kerangka dan autentikasi

`MiddlewareAuth` pada versi kerangka meloloskan semua request dengan pengguna palsu berperan Admin. Ini hanya agar 17 halaman bisa diperiksa. Bypass ini wajib diganti Modul 6 sebelum aplikasi dipakai. Jangan mengklaim login sudah berjalan.

## Alur kerja kelompok

- Satu pemilik per folder atau file. Lihat `docs/OWNERSHIP.md`.
- File bersama: Mahdi memegang `cmd/server`, `internal/domain`, `internal/kontrak`, `internal/config`, `internal/database`; Fadli memegang `internal/web`.
- Jangan mengubah file milik modul lain tanpa persetujuan pemiliknya.
- Bekerja di branch modul masing-masing, lalu buka pull request ke `main`. Jangan push langsung ke `main`.
- Semua nominal uang memakai `int64` di Go dan `BIGINT` di MySQL.
