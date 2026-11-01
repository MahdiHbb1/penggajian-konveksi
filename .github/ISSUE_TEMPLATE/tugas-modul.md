---
name: Tugas Modul
about: Satu tugas implementasi untuk satu modul penggajian
title: "[Modul X] "
labels: ["implementasi"]
assignees: ""
---

## Pemilik

- Nama:
- Username GitHub:
- Branch: `modul-X-...`

## Lingkup

- Halaman:
- File utama:
- Aturan bisnis:

## Checklist

- [ ] Baca aturan terkait di `PROMPT_Agent_Kerangka_Web_Penggajian.md`
- [ ] Implementasi repository, service, dan handler
- [ ] Hapus placeholder `ErrBelumDiimplementasi` di modul ini
- [ ] Isi template dengan data nyata
- [ ] Ganti `t.Skip` dengan tes yang benar-benar berjalan
- [ ] `gofmt -l .` bersih
- [ ] `go build ./...` lolos
- [ ] `go vet ./...` lolos
- [ ] `go test ./...` lolos
- [ ] Uji manual halaman utama modul
- [ ] Buka pull request ke `main`

## Definition of Done

- [ ] Tidak ada `ErrBelumDiimplementasi` yang tersisa di modul ini
- [ ] Semua tes modul lulus tanpa `t.Skip`
- [ ] Halaman mengembalikan 200 dengan data dari database
- [ ] Aturan validasi dan perhitungan sudah diuji
- [ ] File modul lain tidak diubah tanpa persetujuan pemiliknya
- [ ] Pull request sudah direview minimal satu anggota lain

## Catatan koordinasi

Tuliskan kontrak atau format data yang perlu disepakati dengan modul lain.
