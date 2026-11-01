## Ringkasan

Jelaskan perubahan dan alasan singkatnya.

## Modul dan pemilik

- Modul:
- PIC:
- Issue: `Closes #...`

## Yang berubah

- [ ] Repository
- [ ] Service
- [ ] Handler
- [ ] Template
- [ ] Tes
- [ ] Dokumentasi

## Aturan yang diterapkan

Sebutkan nomor aturan dari dokumen tugas, misalnya 5, 6, 7.

## Cara menguji

Tuliskan perintah dan langkah manual yang dipakai untuk membuktikan perubahan.

```powershell
gofmt -l .
go build ./...
go vet ./...
go test ./...
```

## Checklist

- [ ] `gofmt -l .` bersih
- [ ] `go build ./...` lolos
- [ ] `go vet ./...` lolos
- [ ] `go test ./...` lolos
- [ ] Tes baru berjalan, bukan `t.Skip`
- [ ] Tidak ada `.env`, password, token, atau data pribadi di diff
- [ ] Tidak mengubah file modul lain tanpa persetujuan
- [ ] Halaman sudah dicek di browser
- [ ] Screenshot dilampirkan untuk perubahan tampilan

## Catatan reviewer

- [ ] Logika sesuai aturan bisnis
- [ ] Tidak ada query atau perhitungan di handler
- [ ] Error ditangani, bukan diabaikan
- [ ] Nama variabel dan pesan memakai bahasa Indonesia yang konsisten
