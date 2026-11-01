# [Modul 1] Pekerja dan riwayat upah

- **PIC:** Mahdi
- **Branch:** `modul-1-pekerja`
- **Halaman:** 3 `/pekerja`, 4 `/pekerja/baru` dan `/pekerja/{id}/ubah`, 5 `/pekerja/{id}/upah`
- **Aturan:** 1, 2, 8, 9, 11, 16

## File utama

- `internal/pekerja/repository.go`
- `internal/pekerja/service.go`
- `internal/pekerja/handler.go`
- `internal/pekerja/pekerja_test.go`
- `web/templates/pekerja_daftar.html`
- `web/templates/pekerja_form.html`
- `web/templates/upah_riwayat.html`

## Pekerjaan

- [ ] Implementasi `TambahPekerja`, `UbahPekerja`, `NonaktifkanPekerja`
- [ ] Implementasi `DaftarPekerja`, `CariPekerja`, filter nama, jenis, dan status
- [ ] Implementasi `TambahRiwayatUpah`, `DaftarRiwayatUpah`
- [ ] Implementasi `CariUpahBerlaku(pekerjaID, tanggal)`
- [ ] Implementasi `DaftarPekerjaAktif(tipe, mulai, selesai)`
- [ ] Validasi kode unik maksimal 20 karakter dan nama maksimal 100 karakter
- [ ] Tolak perubahan jenis pekerja setelah disimpan
- [ ] Tolak tanggal berhenti sebelum tanggal mulai
- [ ] Tolak upah ganda pada tanggal berlaku yang sama
- [ ] Tolak upah yang berlaku sebelum tanggal mulai kerja
- [ ] Isi tabel dan form dengan data dari service
- [ ] Ganti semua `t.Skip` di `pekerja_test.go`

## Kontrak yang dijaga

- `kontrak.PekerjaProvider` dipakai Modul 4 dan Modul 2.
- `CariUpahBerlaku` harus membaca upah terbaru yang tidak melewati tanggal kerja.
- `DaftarPekerjaAktif` harus menghormati tanggal mulai, tanggal berhenti, dan jenis.

## Definition of Done

- [ ] Tidak ada `ErrBelumDiimplementasi` di `internal/pekerja`
- [ ] Tes `pekerja_test.go` lulus tanpa skip
- [ ] Ketiga halaman mengembalikan 200 dengan data nyata
- [ ] Data pekerja tidak pernah dihapus permanen
- [ ] Pull request direview minimal satu anggota lain
