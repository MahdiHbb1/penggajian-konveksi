# Modul 1 — Pekerja dan Riwayat Upah Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement Modul 1 (pekerja, riwayat upah, pencarian upah berlaku, daftar pekerja aktif) sesuai issue #1 dan aturan bisnis 1, 2, 8, 9, 11, 16.

**Architecture:** Handler menerima request dan menyiapkan view model; Service memegang validasi dan aturan bisnis; Repository mengakses MySQL. Service bergantung pada interface `Repositori` agar aturan bisnis dapat diuji tanpa MySQL. `TambahPekerja` menyimpan pekerja dan upah pertama dalam satu transaksi repository.

**Tech Stack:** Go 1.24, `net/http` (Go 1.22 pattern routing), `html/template`, MySQL 8 via `database/sql`, `github.com/go-sql-driver/mysql`.

**Spec:** `../PROMPT_Agent_Kerangka_Web_Penggajian.md` (bagian 7 aturan 1–19, bagian 8.1, bagian 10, bagian 11.2 halaman 3–5, bagian 13.5–13.6) dan `docs/issues/modul-1.md` (issue #1).

## Global Constraints

- Uang selalu `domain.Rupiah`/`int64`; jangan memakai `float64` untuk uang.
- Data pekerja tidak pernah dihapus permanen; nonaktifkan dengan `tanggal_berhenti` + `aktif = false`.
- Kontrak `kontrak.PekerjaProvider` tidak berubah: `CariUpahBerlaku(ctx, pekerjaID, tanggal)` dan `DaftarPekerjaAktif(ctx, jenis, mulai, selesai)`.
- `CariUpahBerlaku` memakai upah dengan `berlaku_mulai` terbaru yang tidak melewati tanggal kerja.
- `DaftarPekerjaAktif` menghormati `tanggal_mulai <= selesai`, `tanggal_berhenti IS NULL OR tanggal_berhenti >= mulai`, dan jenis.
- Format tampilan: uang `Rp 100.000`, tanggal `02 Jan 2006`; nilai input `YYYY-MM-DD`.
- Tes harus lolos tanpa MySQL. Repository SQL diverifikasi manual terhadap MySQL pada task verifikasi.
- Jangan mengubah file milik modul lain (`internal/web`, `layout.html`, template modul lain) tanpa persetujuan pemiliknya.
- Pesan galat yang dipakai persis: `Kode pekerja sudah dipakai.`, `Tanggal berhenti tidak boleh sebelum tanggal mulai.`, `Sudah ada upah dengan tanggal berlaku yang sama.`, `Tanggal berlaku upah tidak boleh sebelum pekerja mulai bekerja.`
- Branch kerja: `modul-1-pekerja`. Pesan commit bahasa Indonesia.

## Review Focus

- Pekerja dengan `tanggal_berhenti` tepat di awal/akhir periode: harus ikut aktif pada periode yang bersinggungan, tidak pada periode setelahnya.
- Perbandingan tanggal: input `DATE` tidak boleh dipengaruhi jam/timezone; bandingkan tanggal saja.
- Upah pada tanggal yang sama dengan tanggal mulai: harus diterima sebagai upah pertama, tetapi upah kedua pada tanggal itu ditolak.
- Pekerja tanpa riwayat upah: `CariUpahBerlaku` harus mengembalikan error yang bisa ditangani Modul 4, bukan `Rp 0` diam-diam.
- Filter kosong/`jenis` tidak valid: handler tidak boleh membuat query error atau halaman 500.

---

### Task 1: Aturan service pekerja dan kontrak repository

**Files:**
- Modify: `internal/domain/tipe.go`
- Modify: `internal/pekerja/service.go`
- Modify: `internal/pekerja/repository.go` (tambah signature + stub yang masih placeholder)
- Test: `internal/pekerja/pekerja_test.go`

**Interfaces:**
- Consumes: `domain.Pekerja`, `domain.Upah`, `domain.Rupiah`, `domain.JenisPekerja`.
- Produces:
  - `domain.ErrTidakDitemukan`, `domain.ErrDuplikat`.
  - `type Repositori interface { SimpanPekerja; SimpanPekerjaDanUpah; UbahPekerja; NonaktifkanPekerja; DaftarPekerja; CariPekerja; KodeSudahDipakai; SimpanUpah; DaftarUpah; UpahBerlaku; UpahPadaTanggal; PekerjaAktif }`.
  - `func NewService(repo Repositori) *Service`.
  - `func (s *Service) TambahPekerja(ctx, pekerja, upahAwal) (int64, error)`, `UbahPekerja`, `NonaktifkanPekerja`, `DaftarPekerja`, `CariPekerja`.

- [ ] **Step 1: Write the failing service tests**

Add a `fakeRepository` in `pekerja_test.go` implementing every `Repositori` method with in-memory maps. Add tests:

```go
func TestTambahPekerjaValidasi(t *testing.T) // kode kosong, kode >20, nama kosong, nama >100, jenis invalid, tanggal mulai zero, upah <=0
func TestTambahPekerjaKodeSudahDipakai(t *testing.T) // fake punya kode "P-001"; error harus "Kode pekerja sudah dipakai."
func TestTambahPekerjaMenyimpanUpahAwal(t *testing.T) // pekerja.Aktif == true; upah pertama bertanggal sama dengan tanggal mulai; nominal == upahAwal
func TestUbahPekerjaTolakPerubahanJenis(t *testing.T) // existing HARIAN, input BULANAN -> error
func TestUbahTanggalBerhentiSebelumMulai(t *testing.T) // error harus "Tanggal berhenti tidak boleh sebelum tanggal mulai."
func TestUbahPekerjaMempertahankanJenis(t *testing.T) // nama berubah, jenis tetap; tanggal berhenti mengisi aktif=false
func TestNonaktifkanPekerja(t *testing.T) // aktif=false dan tanggal berhenti tersimpan; tanggal sebelum mulai ditolak
func TestDaftarPekerjaMeneruskanFilter(t *testing.T) // fake merekam Filter{Cari, Jenis, Aktif}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/pekerja -run "TestTambahPekerja|TestUbah|TestNonaktifkan|TestDaftarPekerjaMeneruskanFilter" -v`
Expected: FAIL karena `NewService` masih memakai `*Repository` dan method masih mengembalikan `domain.ErrBelumDiimplementasi` (atau gagal kompilasi karena fake belum memenuhi interface).

- [ ] **Step 3: Implement domain errors, `Repositori`, validasi, dan CRUD pekerja**

In `service.go`:
- `normalisasiPekerja` memangkas spasi kode/nama/alasan, menormalkan tanggal ke tanggal saja.
- Validasi: kode wajib/maks 20 karakter, nama wajib/maks 100, jenis HARIAN/BULANAN, tanggal mulai wajib, tanggal berhenti >= tanggal mulai, upah > 0, alasan maks 100.
- `TambahPekerja`: validasi, cek `KodeSudahDipakai(ctx, kode, 0)`, set `Aktif=true`, buat `Upah{PekerjaID, Nominal: upahAwal, BerlakuMulai: tanggalMulai, Alasan: "Upah awal"}`, panggil `SimpanPekerjaDanUpah`.
- `UbahPekerja`: `CariPekerja`, tolak perubahan jenis, validasi, cek kode unik dengan `kecualiID = pekerja.ID`, pertahankan `Aktif` lama kecuali `TanggalBerhenti != nil` (maka false), panggil `UbahPekerja`.
- `NonaktifkanPekerja`: `CariPekerja`, validasi tanggal >= mulai, panggil `NonaktifkanPekerja` repository.
- `DaftarPekerja` dan `CariPekerja` meneruskan ke repository; `CariPekerja` menolak ID <= 0.
- Tambahkan stub `Repository.KodeSudahDipakai`, `SimpanPekerjaDanUpah`, `UpahPadaTanggal` yang mengembalikan `domain.ErrBelumDiimplementasi` supaya kompilasi tetap jalan sampai Task 2.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/pekerja -v`
Expected: PASS untuk test Task 1; test upah yang belum diisi tetap ada dan akan ditangani Task 2.

- [ ] **Step 5: Commit**

```bash
git add internal/domain/tipe.go internal/pekerja/service.go internal/pekerja/repository.go internal/pekerja/pekerja_test.go
git commit -m "feat(pekerja): implementasi aturan service data pekerja"
```

---

### Task 2: Riwayat upah, upah berlaku, pekerja aktif, dan query repository

**Files:**
- Modify: `internal/pekerja/service.go`
- Modify: `internal/pekerja/repository.go`
- Test: `internal/pekerja/pekerja_test.go`

**Interfaces:**
- Consumes: `Repositori` dan sentinel error dari Task 1.
- Produces: `TambahRiwayatUpah`, `DaftarRiwayatUpah`, `CariUpahBerlaku`, `DaftarPekerjaAktif` yang berfungsi; seluruh method `Repository` memakai SQL nyata.

- [ ] **Step 1: Write the failing wage/active-worker tests**

```go
func TestRiwayatUpahTanggalGanda(t *testing.T) // fake sudah punya upah 2026-11-02; input tanggal sama -> "Sudah ada upah dengan tanggal berlaku yang sama."
func TestTambahRiwayatUpahSebelumMulai(t *testing.T) // mulai 2026-11-10, upah 2026-11-01 -> "Tanggal berlaku upah tidak boleh sebelum pekerja mulai bekerja."
func TestTambahRiwayatUpahValid(t *testing.T) // tersimpan dengan pekerja_id dan alasan yang dipangkas
func TestCariUpahBerlakuMemilihTerbaru(t *testing.T) // upah 100.000 @2026-11-01 dan 120.000 @2026-11-05; tanggal 2026-11-06 -> 120.000
func TestCariUpahBerlakuSebelumRiwayatPertama(t *testing.T) // tanggal 2026-10-31 -> domain.ErrTidakDitemukan
func TestDaftarPekerjaAktifMenghormatiRentangDanJenis(t *testing.T) // pekerja berhenti tepat di awal periode ikut; pekerja berhenti sehari sebelum periode tidak ikut
func TestDaftarRiwayatUpahTerurut(t *testing.T) // fake mengembalikan urutan tanggal terbaru lebih dulu
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/pekerja -run "TestRiwayat|TestTambahRiwayat|TestCariUpah|TestDaftarPekerjaAktif|TestDaftarRiwayat" -v`
Expected: FAIL karena method masih `ErrBelumDiimplementasi`.

- [ ] **Step 3: Implement service wage methods**

- `TambahRiwayatUpah`: validasi `PekerjaID > 0`, nominal > 0, tanggal wajib, alasan maks 100; `CariPekerja`; tolak tanggal sebelum `TanggalMulai`; cek `UpahPadaTanggal`; panggil `SimpanUpah`.
- `DaftarRiwayatUpah`: validasi ID; panggil `DaftarUpah`.
- `CariUpahBerlaku`: validasi ID dan tanggal; panggil `UpahBerlaku`; teruskan `domain.ErrTidakDitemukan` bila tidak ada.
- `DaftarPekerjaAktif`: validasi rentang (`mulai <= selesai`), jenis opsional harus valid; panggil `PekerjaAktif`.

- [ ] **Step 4: Implement repository SQL**

Implementasi query berikut di `repository.go` (gunakan `?`, `sql.NullTime` untuk `tanggal_berhenti`, `sql.NullString` untuk `alasan`):
- `SimpanPekerja`: `INSERT INTO pekerja (kode, nama, jenis, tanggal_mulai, tanggal_berhenti, aktif) VALUES (?, ?, ?, ?, ?, ?)`.
- `SimpanPekerjaDanUpah`: `BeginTx`, insert pekerja, set `upah.PekerjaID = pekerja.ID`, insert upah, `Commit`; rollback pada error.
- `UbahPekerja`: `UPDATE pekerja SET kode=?, nama=?, tanggal_mulai=?, tanggal_berhenti=?, aktif=? WHERE id=?` (jenis tidak diubah).
- `NonaktifkanPekerja`: `UPDATE pekerja SET tanggal_berhenti=?, aktif=FALSE WHERE id=?`.
- `DaftarPekerja`: `WHERE` dinamis untuk `kode/nama LIKE`, `jenis`, `aktif`; `ORDER BY kode`.
- `CariPekerja`: `SELECT ... WHERE id=?`; `sql.ErrNoRows` -> `domain.ErrTidakDitemukan`.
- `KodeSudahDipakai`: `SELECT EXISTS(SELECT 1 FROM pekerja WHERE kode=? AND id<>?)`.
- `SimpanUpah`: `INSERT INTO riwayat_upah (pekerja_id, nominal, berlaku_mulai, alasan) VALUES (?, ?, ?, ?)`.
- `DaftarUpah`: `ORDER BY berlaku_mulai DESC, id DESC`.
- `UpahBerlaku`: `WHERE pekerja_id=? AND berlaku_mulai<=? ORDER BY berlaku_mulai DESC, id DESC LIMIT 1`; tidak ada baris -> `domain.ErrTidakDitemukan`.
- `UpahPadaTanggal`: `SELECT EXISTS(SELECT 1 FROM riwayat_upah WHERE pekerja_id=? AND berlaku_mulai=?)`.
- `PekerjaAktif`: `WHERE tanggal_mulai<=? AND (tanggal_berhenti IS NULL OR tanggal_berhenti>=?)`, filter jenis bila diisi, `ORDER BY nama, kode`.
- Semua method mengembalikan error jelas bila `r.db == nil`, dan menerjemahkan duplikat MySQL 1062 menjadi pesan spesifikasi.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./internal/pekerja -v`
Expected: PASS, tidak ada `t.Skip`.

- [ ] **Step 6: Commit**

```bash
git add internal/pekerja/service.go internal/pekerja/repository.go internal/pekerja/pekerja_test.go
git commit -m "feat(pekerja): implementasi riwayat upah dan query repository"
```

---

### Task 3: Handler dan tiga halaman Modul 1

**Files:**
- Modify: `internal/pekerja/handler.go`
- Modify: `web/templates/pekerja_daftar.html`
- Modify: `web/templates/pekerja_form.html`
- Modify: `web/templates/upah_riwayat.html`
- Test: `internal/pekerja/handler_test.go` (baru)

**Interfaces:**
- Consumes: `*Service` dari Task 1–2, `web.Renderer`, `web.Data`.
- Produces: view model `DaftarView`, `FormView`, `UpahView`; halaman `/pekerja`, `/pekerja/baru`, `/pekerja/{id}/ubah`, `/pekerja/{id}/upah`, aksi POST-nya.

- [ ] **Step 1: Write the failing handler tests**

Use `web.Render{Dir: filepath.Join("..", "..", "web", "templates")}`, `NewService(fakeRepository)`, `http.NewServeMux`, dan `DaftarkanRute`. Tests:

```go
func TestHalamanDaftarPekerjaMenampilkanData(t *testing.T) // GET /pekerja -> 200, memuat kode dan nama
func TestHalamanDaftarPekerjaMenerapkanFilter(t *testing.T) // GET /pekerja?cari=Budi&jenis=HARIAN&aktif=1 -> 200, filter diteruskan
func TestHalamanFormBaruDanSimpan(t *testing.T) // GET 200; POST valid -> 303 ke /pekerja; data tersimpan
func TestHalamanFormUbahDanSimpan(t *testing.T) // GET memuat nilai lama; POST -> 303; data berubah
func TestHalamanRiwayatUpahDanSimpan(t *testing.T) // GET 200 memuat nominal; POST -> 303; riwayat bertambah
func TestFormMenampilkanGalatValidasi(t *testing.T) // POST upah 0 -> 200 dengan pesan "Nominal upah harus lebih dari 0."
func TestNonaktifkanPekerjaDariDaftar(t *testing.T) // POST /pekerja/1/nonaktif -> 303; pekerja nonaktif
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/pekerja -run "TestHalaman|TestForm|TestNonaktifkanPekerjaDariDaftar" -v`
Expected: FAIL karena handler masih merender data kosong dan aksi POST belum memproses form.

- [ ] **Step 3: Implement handlers**

- `Daftar`: parse `cari`, `jenis`, `aktif` (`1`/`0`), panggil `DaftarPekerja`, bangun `[]BarisPekerja` dengan format `02 Jan 2006`, label jenis, dan status.
- `FormBaru`/`FormUbah`: siapkan `FormView{Aksi, Baru, Kode, Nama, Jenis, TanggalMulai, TanggalBerhenti, UpahAwal}`.
- `SimpanBaru`: parse form, panggil `TambahPekerja`, redirect 303 ke `/pekerja?info=...`; validasi gagal -> render form 200 dengan `Galat`.
- `SimpanUbah`: ambil data lama, gunakan jenis lama bila field jenis tidak dikirim, panggil `UbahPekerja`, redirect 303.
- `Nonaktifkan`: parse `tanggal_berhenti` (default hari ini), panggil service, redirect 303 atau render daftar dengan galat.
- `RiwayatUpah`: panggil `CariPekerja` + `DaftarRiwayatUpah`, bangun `UpahView`.
- `SimpanUpah`: parse nominal/tanggal/alasan, panggil `TambahRiwayatUpah`, redirect 303 atau render ulang 200 dengan galat.
- Gunakan `r.PathValue("id")`, `strconv.ParseInt`, `time.ParseInLocation("2006-01-02", ..., time.Local)`.

- [ ] **Step 4: Implement templates**

- `pekerja_daftar.html`: pertahankan nilai filter, tampilkan tabel dari `.Isi.Pekerja`, empty state, tombol Ubah/Upah/Nonaktifkan; tombol Nonaktifkan hanya untuk pekerja aktif.
- `pekerja_form.html`: isi nilai form, radio jenis `checked` sesuai data, nonaktifkan radio jenis saat ubah, tampilkan upah awal hanya saat baru.
- `upah_riwayat.html`: tampilkan nama/kode pekerja, tabel riwayat, form tambah upah dengan action `/pekerja/{{.Isi.Pekerja.ID}}/upah`, dan catatan periode tertutup.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./internal/pekerja -v`
Expected: PASS, ketiga halaman merender 200 dengan data dari fake.

- [ ] **Step 6: Commit**

```bash
git add internal/pekerja/handler.go internal/pekerja/handler_test.go web/templates/pekerja_daftar.html web/templates/pekerja_form.html web/templates/upah_riwayat.html
git commit -m "feat(pekerja): sambungkan handler dan halaman ke service"
```

---

### Task 4: Verifikasi penuh dan bukti halaman

**Files:**
- Modify: `docs/PROGRES.md` (baris Modul 1)
- Verify: seluruh repo

**Interfaces:**
- Consumes: seluruh implementasi Task 1–3.

- [ ] **Step 1: Run formatter, build, vet, dan seluruh tes**

Run:
```bash
gofmt -l .
go build ./...
go vet ./...
go test ./...
```
Expected: `gofmt` tidak menghasilkan path; build/vet/test lolos; jumlah `t.Skip` di `internal/pekerja` = 0.

- [ ] **Step 2: Pastikan tidak ada placeholder Modul 1**

Run: `Select-String -Path internal/pekerja/*.go -Pattern "ErrBelumDiimplementasi"`
Expected: tidak ada hasil.

- [ ] **Step 3: Uji dengan MySQL bila tersedia**

Muat `database/schema.sql` dan `database/seed.sql` ke MySQL 8/9 lokal, jalankan server, lalu:
- `GET /pekerja` -> 200 dan memuat pekerja yang dibuat.
- `POST /pekerja/baru` -> 303, lalu `GET /pekerja/{id}/ubah` -> 200.
- `POST /pekerja/{id}/upah` -> 303, lalu `GET /pekerja/{id}/upah` -> 200 dan memuat nominal.
- `POST /pekerja/{id}/nonaktif` -> 303, lalu `GET /pekerja` -> status Nonaktif.
Jika MySQL tidak tersedia, catat hasilnya di laporan akhir; jangan mengklaim sudah diuji.

- [ ] **Step 4: Perbarui docs/PROGRES.md**

Ubah baris Modul 1: status `Done` setelah verifikasi, isi kolom Tes dengan hasil `go test ./...`, dan catat branch `modul-1-pekerja`.

- [ ] **Step 5: Commit**

```bash
git add docs/PROGRES.md
git commit -m "docs: perbarui progres modul 1"
```
