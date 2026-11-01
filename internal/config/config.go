// Pemilik: Mahdi | Modul 1 | Aturan: 5, 6, 7, 12
package config

import "os"

// Konstanta aturan bisnis. Angka-angka ini tidak boleh tersebar di kode.
const (
	// MenitTerlambatDipotong adalah batas keterlambatan yang masih dimaafkan.
	MenitTerlambatDipotong = 15
	// JamKerjaPerHari dipakai untuk upah per jam pekerja harian.
	JamKerjaPerHari = 8
	// HariKerjaPerBulan adalah pembagi upah per jam pekerja bulanan.
	HariKerjaPerBulan = 173
	// PengaliLemburPertama berlaku untuk 60 menit pertama.
	PengaliLemburPertama = 1.5
	// PengaliLemburBerikutnya berlaku untuk menit setelah 60 menit pertama.
	PengaliLemburBerikutnya = 2.0
	// PersenBatasKasbon adalah batas maksimum potongan kasbon dari gaji bruto.
	PersenBatasKasbon = 30
)

// Config adalah konfigurasi aplikasi yang dibaca dari environment.
type Config struct {
	Port          string
	DSN           string
	SessionSecret string
	TemplateDir   string
	StaticDir     string
}

// Muat membaca konfigurasi dari environment dengan nilai default lokal.
//
// Masukan: environment proses.
// Hasil: Config siap dipakai.
// Aturan: tidak ada aturan bisnis, hanya konfigurasi.
// Tabel: tidak ada.
// Memanggil: os.Getenv.
func Muat() Config {
	return Config{
		Port:          env("PORT", "8080"),
		DSN:           env("DB_DSN", "root:@tcp(127.0.0.1:3306)/penggajian_konveksi?parseTime=true&loc=Local"),
		SessionSecret: os.Getenv("SESSION_SECRET"),
		TemplateDir:   env("TEMPLATE_DIR", "web/templates"),
		StaticDir:     env("STATIC_DIR", "web/static"),
	}
}

// env mengembalikan nilai environment atau nilai default.
func env(kunci, bawaan string) string {
	if nilai := os.Getenv(kunci); nilai != "" {
		return nilai
	}
	return bawaan
}
