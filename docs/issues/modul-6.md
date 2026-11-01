# [Modul 6] Slip, rekap, autentikasi, dan antarmuka

- **PIC:** Fadli
- **Branch:** `modul-6-laporan-antarmuka`
- **Halaman:** 1 `/login`, 2 `/dashboard`, 16 `/slip/{periodeID}/{pekerjaID}`, 17 `/rekap/{periodeID}`
- **Aturan:** 12, 13, 17, akses Admin dan Pemilik

## File utama

- `internal/laporan/repository.go`, `service.go`, `handler.go`
- `internal/auth/service.go`, `middleware.go`, `handler.go`
- `internal/web/render.go`, `router.go`
- `internal/laporan/laporan_test.go`, `internal/auth/auth_test.go`
- `web/templates/login.html`, `dashboard.html`, `slip.html`, `rekap.html`
- `web/static/css/style.css`

## Pekerjaan

- [ ] Implementasi `Login` dengan bcrypt, tanpa password teks biasa
- [ ] Implementasi sesi HttpOnly dan `Logout`
- [ ] Ganti bypass `MiddlewareAuth` dengan pengecekan sesi sungguhan
- [ ] Implementasi `MiddlewarePeran` untuk Admin dan Pemilik
- [ ] Tampilkan pesan "Anda tidak memiliki akses ke halaman ini."
- [ ] Implementasi `AmbilSlipGaji`, `DaftarSlipPeriode`, `RekapGajiPerPeriode`
- [ ] Slip dibaca dari data tersimpan, bukan dihitung ulang
- [ ] Hanya periode Dihitung atau Ditutup yang menampilkan slip dan rekap
- [ ] Isi dashboard dengan ringkasan nyata
- [ ] Pastikan rekap tidak memotong keterlambatan dua kali
- [ ] Periksa tampilan di browser dan cetak slip
- [ ] Ganti semua `t.Skip` di `auth_test.go` dan `laporan_test.go`

## Keamanan

- [ ] Tidak ada password, token, atau `.env` di repo
- [ ] Cookie sesi HttpOnly, SameSite, dan Secure saat produksi
- [ ] Jangan menampilkan hash password di halaman atau log
- [ ] Jangan membuat akun contoh yang terlihat nyata

## Definition of Done

- [ ] Tidak ada `ErrBelumDiimplementasi` di `internal/laporan` dan `internal/auth`
- [ ] Login, logout, dan pembatasan peran terbukti di tes
- [ ] Keempat halaman mengembalikan 200 sesuai peran
- [ ] Slip dan rekap hanya menampilkan data tersimpan
- [ ] Pull request direview minimal satu anggota lain
