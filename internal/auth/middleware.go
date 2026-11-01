// Pemilik: Fadli | Modul 6 | Aturan: akses Admin dan Pemilik
package auth

import (
	"net/http"

	"penggajian-konveksi/internal/domain"
	"penggajian-konveksi/internal/web"
)

// MiddlewareAuth memasang pengguna pada context request.
//
// Masukan: handler berikutnya.
// Hasil: handler yang sudah membawa pengguna.
// Aturan: semua halaman selain /login memerlukan sesi sah.
// Tabel: pengguna (nanti).
// Memanggil: web.DenganPengguna.
//
// TODO(Fadli): ganti dengan sesi bcrypt dan pengecekan peran.
//
// PERINGATAN: versi kerangka ini meloloskan semua request dengan pengguna
// palsu berperan Admin supaya 17 halaman bisa diperiksa. Bypass ini hanya
// untuk mode kerangka dan tidak boleh dipakai di aplikasi akhir.
func MiddlewareAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pengguna := &domain.Pengguna{
			ID:       0,
			Username: "kerangka",
			Role:     domain.PeranAdmin,
		}
		next.ServeHTTP(w, r.WithContext(web.DenganPengguna(r.Context(), pengguna)))
	})
}

// MiddlewarePeran membatasi akses berdasarkan peran.
//
// Masukan: daftar peran yang diizinkan.
// Hasil: middleware.
// Aturan: Admin mengelola data, Pemilik hanya melihat.
// Tabel: tidak ada.
// Memanggil: web.PenggunaDariContext.
//
// TODO(Fadli): tolak request dengan pesan "Anda tidak memiliki akses ke halaman ini."
// Versi kerangka saat ini meloloskan semua peran.
func MiddlewarePeran(peran ...domain.Peran) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}
