# Panduan GitHub dan Tracking Progres

Dokumen ini menjelaskan cara memindahkan repo ke GitHub, mengatur GitHub Projects, dan alur kerja enam anggota. Semua perintah ditulis untuk PowerShell di Windows.

## 1. Ringkasan keputusan

- `main` adalah cabang utama yang selalu bisa di-build.
- Setiap anggota bekerja di cabang modulnya sendiri.
- Perubahan masuk ke `main` lewat pull request, bukan push langsung.
- Satu issue per modul, dilacak di GitHub Projects.
- Status issue: `Backlog`, `Ready`, `In progress`, `In review`, `Done`.
- Branch modul tidak dihapus sampai proyek selesai.

## 2. Membuat repo dan push pertama

Pastikan `gh` sudah terpasang dan login.

```powershell
gh --version
gh auth status
```

Jika belum login:

```powershell
gh auth login
```

Dari folder `penggajian-konveksi`:

```powershell
gh repo create penggajian-konveksi --private --source=. --remote=origin --push
git push origin --all
```

Ganti `--private` menjadi `--public` hanya jika repo memang boleh dilihat publik. Jangan pernah membuat repo publik untuk data atau kredensial asli.

Jika repo sudah dibuat manual di GitHub:

```powershell
git remote add origin https://github.com/USERNAME/penggajian-konveksi.git
git push -u origin main
git push origin --all
```

Periksa hasilnya:

```powershell
git remote -v
git branch -a
gh repo view --web
```

## 3. Menambahkan kolaborator

Undang lewat web: **Settings > Collaborators > Add people**, lalu kirim username GitHub masing-masing anggota.

Atau lewat CLI:

```powershell
gh api --method PUT repos/OWNER/REPO/collaborators/USERNAME -f permission=push
```

Setiap anggota harus menerima undangan sebelum bisa push.

## 4. Membuat GitHub Project

Project adalah papan tingkat akun atau organisasi, bukan milik satu repo. Buat satu project untuk seluruh mata kuliah ini.

```powershell
gh auth refresh -s project
gh project create --owner @me --title "Penggajian Konveksi"
gh project list --owner @me
```

Catat nomor project dari output. Contoh nomor `1`.

Tambahkan field agar progres per anggota terlihat:

```powershell
gh project field-create 1 --owner @me --name "Modul" --data-type SINGLE_SELECT --single-select-options "1 Pekerja,2 Jadwal dan Kasbon,3 Absensi,4 Penggajian,5 Periode dan Koreksi,6 Laporan dan Antarmuka"
gh project field-create 1 --owner @me --name "PIC" --data-type TEXT
gh project field-create 1 --owner @me --name "Branch" --data-type TEXT
```

Buat view berikut lewat tombol **New view**:

| Nama view | Tipe | Isi |
|---|---|---|
| Semua modul | Table | Title, Status, PIC, Modul, Branch, linked pull requests |
| Papan status | Board | Group by Status |
| Per anggota | Board | Group by PIC |
| Butuh review | Table | Filter `is:open is:pr review-requested:@me` |

Aktifkan automasi di **Project > Workflows**:

- **Auto-add to project**: filter `is:issue is:open` untuk repo ini.
- **Item closed**: set Status ke `Done`.
- **Pull request merged**: set Status ke `Done`.

## 5. Membuat label dan issue

Isi username GitHub anggota terlebih dahulu:

```powershell
Copy-Item docs/github-anggota.example.json docs/github-anggota.json
notepad docs/github-anggota.json
```

Jalankan script dari root repo:

```powershell
.\scripts\siapkan-github.ps1 -ProjectNumber 1
```

Script akan:

- membuat label `modul-1` sampai `modul-6`, `implementasi`, `tes`, dan `dokumen`;
- membuat enam issue dari `docs/issues/`;
- menambahkan setiap issue ke project;
- meng-assign issue jika username di `docs/github-anggota.json` sudah diisi.

Jika belum ingin memakai script, buat issue manual dari template **Tugas Modul**, lalu tempel isi file `docs/issues/modul-X.md`.

## 6. Alur kerja anggota

### 6.1 Persiapan sekali di komputer masing-masing

```powershell
git clone https://github.com/OWNER/penggajian-konveksi.git
cd penggajian-konveksi
git config user.name "Nama Lengkap"
git config user.email "email@contoh.com"
```

Gunakan email yang sama dengan akun GitHub, atau email noreply GitHub, supaya kontribusi terhubung ke profil.

### 6.2 Mulai bekerja

```powershell
git fetch origin
git switch modul-1-pekerja
git merge origin/main
```

Kerjakan satu potongan kecil, lalu commit:

```powershell
git add .
git commit -m "feat(modul-1): tambah validasi kode pekerja unik"
git push -u origin modul-1-pekerja
```

Format pesan commit:

```text
feat(modul-1): tambah form pekerja
fix(modul-2): tolak jadwal hari Minggu
test(modul-3): uji lembur lewat tengah malam
docs: perbarui panduan tracking
```

### 6.3 Membuka pull request

Sebelum PR, pastikan branch sudah menggabungkan `main` terbaru:

```powershell
git fetch origin
git merge origin/main
gofmt -l .
go build ./...
go vet ./...
go test ./...
git push
```

Buat PR:

```powershell
gh pr create --base main --head modul-1-pekerja --title "feat(modul-1): pekerja dan riwayat upah" --body-file .github/pull_request_template.md
```

Atau lewat web: **Pull requests > New pull request**, pilih `main` sebagai base dan branch modul sebagai compare.

Aturan review:

- Minimal satu anggota lain menyetujui.
- Jika mengubah file bersama, minta review pemilik file tersebut.
- `cmd/server`, `internal/domain`, `internal/kontrak`, `internal/config`, `internal/database`, dan `database/` direview Mahdi.
- `internal/web` dan `web/` direview Fadli.
- Jangan merge PR sendiri tanpa review, kecuali keadaan darurat yang dicatat di issue.

### 6.4 Setelah PR di-merge

```powershell
git switch main
git pull origin main
git switch modul-1-pekerja
git merge main
git push
```

Jangan menghapus branch modul sampai proyek selesai. Branch itu dipakai lagi untuk potongan pekerjaan berikutnya.

### 6.5 Jika dua orang menyentuh modul yang sama

Buat sub-branch dari branch modul, lalu PR ke branch modul, bukan langsung ke `main`.

```powershell
git switch modul-1-pekerja
git switch -c modul-1-pekerja/perbaikan-upah
git push -u origin modul-1-pekerja/perbaikan-upah
```

Setelah PR sub-branch di-merge, hapus sub-branch dan lanjutkan dari `modul-1-pekerja`.

## 7. Melindungi `main`

Lewat web: **Settings > Branches > Add branch protection rule**.

- Branch name pattern: `main`
- Centang **Require a pull request before merging**
- Centang **Require approvals**, minimal 1
- Centang **Dismiss stale pull request approvals when new commits are pushed**
- Centang **Require status checks to pass before merging**
- Centang **Require branches to be up to date before merging**
- Centang **Do not allow bypassing the above settings**

Jika nanti ada GitHub Actions, tambahkan nama check-nya, misalnya `build` dan `test`.

## 8. Memperbarui progres

Setiap anggota memperbarui issue, bukan hanya mengirim chat.

- Pindahkan kartu ke `In progress` saat mulai.
- Pindahkan ke `In review` saat PR dibuka.
- Pindahkan ke `Done` saat PR di-merge dan tes lulus.
- Tulis komentar singkat di issue setiap akhir pekan: apa yang selesai, tes apa yang lulus, apa yang menghambat.

Contoh komentar:

```text
Status: In progress
Selesai: CRUD pekerja dan validasi kode unik
Tes: go test ./internal/pekerja lulus
Berikutnya: form riwayat upah
Hambatan: belum ada
```

`docs/PROGRES.md` adalah rekap manual untuk laporan kampus. Sumber kebenaran harian tetap GitHub Projects.

## 9. Rahasia dan data asli

- Jangan commit `.env`, password, token, atau data pekerja asli.
- `seed.sql` hanya berisi shift, bukan akun.
- Jika sebuah rahasia sudah terlanjur di-commit, hapus dari riwayat dan ganti nilainya. Menghapus file di commit baru tidak cukup.

## 10. Masalah yang sering muncul

| Gejala | Perbaikan |
|---|---|
| `gh: command not found` | Install GitHub CLI, lalu buka terminal baru |
| `missing required scope 'project'` | `gh auth refresh -s project` |
| `remote origin already exists` | `git remote set-url origin URL_BARU` |
| Push ditolak karena branch tertinggal | `git fetch origin`, `git merge origin/main`, selesaikan konflik, push lagi |
| Kolaborator tidak melihat Project | Undang lewat **Project > Settings > Manage access**, atau tautkan project ke repo |
| Anggota tidak bisa push ke `main` | Ini disengaja. Buat branch dan pull request |
| Author commit bukan akun sendiri | Set `user.name` dan `user.email`, lalu perbaiki commit sebelum push |
