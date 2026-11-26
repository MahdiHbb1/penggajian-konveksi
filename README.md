# Sistem Penggajian Konveksi Mitra Jaya

Proyek akhir mata kuliah Pemrograman Web. Aplikasi Go dengan enam modul, 17 halaman, 12 tabel, dan kontrak antarmodul.

## Stack

- Go 1.22 atau lebih baru
- MySQL 8
- `html/template` dan CSS sederhana
- Tanpa framework frontend

## Menjalankan

```bash
cp .env.example .env
# Sesuaikan DB_DSN di .env
go run ./cmd/server
```

Buka `http://localhost:8080`.

## Menyiapkan MySQL

```bash
mysql -u root -p < database/schema.sql
mysql -u root -p < database/seed.sql
```

`seed.sql` berisi dua shift: Pagi 07.00-15.00 dan Siang 15.00-23.00.

## Membuat Akun Pengguna

Autentikasi menggunakan bcrypt (cost 12). Untuk membuat akun:

1. Buat hash bcrypt dari password lokal:

```bash
go run internal/auth/genhash/main.go <password-anda>
```

2. Masukkan akun ke database:

```sql
INSERT INTO pengguna (username, password_hash, role)
VALUES ('admin', '<hash-bcrypt-dari-langkah-1>', 'ADMIN');
```

Jangan menyimpan password asli di source code, seed, atau README.

## Konfigurasi

| Variabel | Default | Keterangan |
|---|---|---|
| `PORT` | `8080` | Port HTTP |
| `DB_DSN` | `root:@tcp(127.0.0.1:3306)/penggajian_konveksi?parseTime=true&loc=Local` | DSN MySQL |
| `SESSION_SECRET` | kosong | Diisi lokal, jangan di-commit |
| `TEMPLATE_DIR` | `web/templates` | Folder template |
| `STATIC_DIR` | `web/static` | Folder aset statis |
| `SECURE_COOKIE` | `false` | Set `true` untuk production/HTTPS |

## Autentikasi dan Keamanan

- Password di-hash dengan **bcrypt cost 12** (`golang.org/x/crypto/bcrypt`) dengan perbandingan username exact (case-sensitive) dan dummy compare untuk mencegah enumerasi akun.
- Session cookie: `HttpOnly`, `SameSite=Lax`, `Secure` (diatur lewat `SECURE_COOKIE=true` di `.env` atau environment). Sesi disimpan in-memory (pengguna login ulang bila server restart).
- Proteksi **CSRF** penuh: Double-Submit Cookie (HttpOnly) pada `/login` dan Synchronizer Token terikat sesi pada seluruh permintaan mutasi (`POST`/`PUT`/`PATCH`/`DELETE`) di semua modul (fail-closed).
- **Rate limiting** login atomik (anti-TOCTOU, case-insensitive, maks 5 percobaan gagal per 5 menit, kunci 15 menit) dan pembatas konkurensi bcrypt.
- Middleware `MiddlewareAuth` melindungi seluruh halaman kecuali `/login` dan `/static/`, serta memasang header keamanan (`X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`).
- Middleware `MiddlewarePeran` membatasi seluruh modul operasional (Modul 1–5) khusus `ADMIN`, sedangkan `/dashboard`, `/rekap`, dan `/slip` dapat diakses `ADMIN` dan `PEMILIK`.

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

- Repo: <https://github.com/MahdiHbb1/penggajian-konveksi> (privat).
- Project: <https://github.com/users/MahdiHbb1/projects/2>.
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

## Alur kerja kelompok

- Satu pemilik per folder atau file. Lihat `docs/OWNERSHIP.md`.
- File bersama: Mahdi memegang `cmd/server`, `internal/domain`, `internal/kontrak`, `internal/config`, `internal/database`; Fadli memegang `internal/web`.
- Jangan mengubah file milik modul lain tanpa persetujuan pemiliknya.
- Bekerja di branch modul masing-masing, lalu buka pull request ke `main`. Jangan push langsung ke `main`.
- Semua nominal uang memakai `int64` di Go dan `BIGINT` di MySQL.
