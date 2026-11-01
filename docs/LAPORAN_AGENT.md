# Laporan Agent: Kerangka Web Sistem Penggajian

## 1. Ringkasan

Kerangka proyek Go untuk Sistem Penggajian Konveksi Mitra Jaya sudah dibuat dengan enam modul, 17 kelompok halaman, 12 tabel, dan kontrak antarmodul. Aplikasi dapat di-build dan dijalankan tanpa MySQL; seluruh halaman merender data kosong dengan label pemilik. Badan fungsi bisnis sengaja masih mengembalikan `ErrBelumDiimplementasi`, sesuai batas kerangka. Verifikasi build, vet, tes, 22 URL halaman, 16 aksi POST, dan pemuatan skema pada instance MySQL sementara berhasil.

## 2. Pohon folder akhir

```text
penggajian-konveksi/
  cmd/server/main.go
  internal/
    config/config.go
    database/koneksi.go
    domain/tipe.go
    kontrak/kontrak.go
    pekerja/{handler,service,repository}.go + pekerja_test.go
    jadwal/{handler,service,repository}.go + jadwal_test.go
    kasbon/{handler,service,repository}.go + kasbon_test.go
    absensi/{handler,service,repository}.go + absensi_test.go
    penggajian/{handler,service,repository,kalkulasi}.go + kalkulasi_test.go
    periode/{handler,service,repository}.go + periode_test.go
    laporan/{handler,service,repository}.go + laporan_test.go
    auth/{handler,service,middleware}.go + auth_test.go
    web/{router,render}.go
  web/
    templates/  layout.html + 17 halaman
    static/css/style.css
  database/
    schema.sql
    seed.sql
  docs/
    erd.dbml
    OWNERSHIP.md
    LAPORAN_AGENT.md
  README.md
  .env.example
  .gitignore
  go.mod
  go.sum
```

## 3. Kesesuaian dengan Pengumpulan 2

| Bagian | Target | Status | Bukti |
|---|---|---|---|
| Aturan bisnis | 19 aturan | Tercakup | Komentar aturan di service, handler, dan tes |
| Tabel | 12 tabel | Tercakup | `database/schema.sql` memiliki 12 `CREATE TABLE` |
| Halaman | 17 kelompok halaman | Tercakup | 17 template halaman + layout, 22 URL tampilan |
| Modul | 6 modul | Tercakup | Folder dan pemilik di `docs/OWNERSHIP.md` |
| Contoh A/B/C | Nilai harian, bulanan, proporsional | Tercakup sebagai nilai tes | `internal/penggajian/kalkulasi_test.go` |
| Kasus uji tepi | 15 menit, 61 menit, lewat tengah malam, kasbon, terkunci | Tercakup sebagai nama tes | Tes di setiap modul |

Catatan: 17 kelompok halaman menghasilkan 22 URL GET karena beberapa template dipakai dua route, misalnya form pekerja baru dan ubah. Total seluruh route, termasuk aksi POST, adalah 40.

## 4. Pemilik, fungsi, file, dan halaman

| Anggota | Modul | File utama | Halaman | Tes |
|---|---|---|---|---|
| Mahdi | 1 | 3 file pekerja, domain, kontrak, config, database, schema, seed, ERD, main | 3, 4, 5 | `pekerja_test.go` |
| Edgar | 2 | 3 file jadwal, 3 file kasbon | 6, 7, 8 | `jadwal_test.go`, `kasbon_test.go` |
| Zahra | 3 | 3 file absensi | 9, 10 | `absensi_test.go` |
| Farah | 4 | 3 file penggajian + `kalkulasi.go` | 11, 12 | `kalkulasi_test.go` |
| Anya | 5 | 3 file periode | 13, 14, 15 | `periode_test.go` |
| Fadli | 6 | 3 file laporan, 3 file auth, 2 file web, 18 template, CSS | 1, 2, 16, 17 | `auth_test.go`, `laporan_test.go` |

Rincian checklist ada di `docs/OWNERSHIP.md`.

## 5. Skill, plugin, dan tool

- Skill yang tersedia di harness: executing-plans, verification-before-completion, systematic-debugging, frontend design, dan skill lain yang tidak relevan dengan kerangka backend.
- Skill yang dipakai: `executing-plans` untuk menjalankan spesifikasi secara inline dan menjaga urutan verifikasi.
- Tool yang dipakai: shell PowerShell untuk Go dan pemeriksaan HTTP, `read`/`write` untuk file, `grep` untuk pencarian.
- Tool yang dilewati: MCP dan plugin eksternal, karena tidak ada yang dibutuhkan untuk membuat kerangka Go dan MySQL.
- Tidak ada skill yang dipakai hanya supaya bisa disebut.

## 6. Hasil verifikasi

| Pemeriksaan | Perintah | Hasil |
|---|---|---|
| Format | `gofmt -l .` | Tidak ada output, bersih |
| Build | `go build ./...` | Lolos |
| Vet | `go vet ./...` | Lolos |
| Tes | `go test ./...` | Lolos, 51 subtes di-skip karena fungsi belum diisi |
| Route halaman | `Invoke-WebRequest` ke 22 URL | 22 dari 22 menjawab 200 dan memuat label pemilik |
| Aksi POST | `Invoke-WebRequest -Method Post` ke 16 aksi | 16 dari 16 menjawab 200 |
| Root | `GET /` | 303 ke `/dashboard` |
| CSS | `GET /static/css/style.css` | 200, 6.442 byte |
| Tanpa MySQL | `go run ./cmd/server` | Berjalan, hanya mencatat peringatan koneksi |
| Skema | `mysql < database/schema.sql` | 12 tabel dibuat tanpa error |
| Seed | `mysql < database/seed.sql` | 2 shift, tanpa akun |
| Constraint | Insert jenis tidak valid | Ditolak `chk_pekerja_jenis`, 17 CHECK aktif |
| Dengan MySQL | `go run` memakai DSN port 3307 | Login koneksi berhasil, dashboard 200 |

Server dijalankan pada `http://127.0.0.1:8080`, lalu dihentikan setelah pemeriksaan. Tidak ada error 500 saat membuka halaman.

## 7. Asumsi

- Proyek kelas dan slide refactor tidak tersedia, jadi struktur default di prompt dipakai.
- Wireframe tidak tersedia sebagai file terpisah; elemen halaman diambil dari tabel 17 halaman dan deskripsi Pengumpulan 2.
- Zona waktu memakai `time.Local`.
- Pembulatan upah proporsional dilakukan ke bawah per segmen upah, lalu dijumlahkan.
- Periode pertama dengan tipe tertentu dianggap lolos prasyarat "periode sebelumnya ditutup".
- `hari_kerja_periode` hanya diisi untuk pekerja bulanan; pekerja harian memakai `NULL`.
- Bypass autentikasi hanya untuk mode kerangka. Aplikasi akhir wajib memakai sesi dan bcrypt.
- Nominal uang selalu `int64`; pengali lembur memakai `float64` karena bukan nilai uang.

## 8. Pertentangan atau kekurangan dokumen

- Dokumen belum menentukan perlakuan gaji bruto negatif, misalnya koreksi negatif yang lebih besar dari upah. Kerangka tidak menambahkan aturan; keputusan diserahkan ke Modul 4.
- Pembulatan upah proporsional per segmen disebut sebagai asumsi karena dokumen hanya menyebut pembulatan ke bawah.
- Cara membuat akun admin pertama belum ditentukan. README hanya menjelaskan bentuk hash bcrypt, tanpa password asli.
- Autentikasi kerangka sengaja meloloskan semua request. Ini bukan perilaku aplikasi akhir dan sudah ditandai TODO.
- Status turunan tidak memakai kolom baru: kasbon dari `sisa`, koreksi dari `periode_terapan_id`, pekerja dari `aktif`.

## 9. Yang belum diuji atau gagal

- `schema.sql` dan `seed.sql` diuji pada instance MySQL 9.5 sementara yang dibuat di folder temp, bukan pada MySQL 8.0 persis. Sintaks yang dipakai tersedia di MySQL 8.0.16 atau lebih baru.
- Login, logout, sesi, dan bcrypt belum berjalan; masih placeholder.
- Seluruh aturan bisnis, query SQL, dan perhitungan belum diimplementasikan.
- Tampilan belum diperiksa di browser sungguhan; pemeriksaan baru pada status HTTP dan isi HTML.
- Tidak ada remote git dan tidak ada push.

## 10. Langkah berikutnya

- Mahdi: isi CRUD pekerja dan riwayat upah, lalu uji `schema.sql` di MySQL 8. Jaga kontrak `CariUpahBerlaku` dan `DaftarPekerjaAktif` tetap stabil.
- Edgar: isi jadwal mingguan, salin jadwal tanpa menimpa, kasbon, dan riwayat potongan. Pastikan tanggal terkunci ditolak.
- Zahra: isi absensi, validasi lewat tengah malam, dan dua fungsi menit. Koordinasikan format datetime dengan Farah.
- Farah: isi seluruh fungsi hitung dan alur transaksi. Mulai dari nilai tes Contoh A/B/C, lalu kasus tepi.
- Anya: isi pembuatan periode, prasyarat, penutupan transaksi, dan koreksi. Pastikan koreksi tidak dihitung dua kali.
- Fadli: ganti bypass dengan sesi bcrypt, batasi peran, lalu sambungkan slip dan rekap ke data tersimpan. Periksa tampilan di browser.
