// Pemilik: Fadli | Modul 6 | Aturan: akses Admin dan Pemilik
package auth

import (
	"net/http"

	"penggajian-konveksi/internal/domain"
	"penggajian-konveksi/internal/web"
)

const namaCookie = "sesi_penggajian"

// halamanPublik adalah rute yang tidak memerlukan login.
var halamanPublik = map[string]bool{
	"/login": true,
}

// MiddlewareAuth memasang pengguna pada context request.
// Semua halaman kecuali /login memerlukan sesi sah.
// Bila tidak ada sesi, redirect ke /login.
func MiddlewareAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if halamanPublik[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		// Ambil cookie sesi.
		cookie, err := r.Cookie(namaCookie)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		pengguna := AmbilSesi(cookie.Value)
		if pengguna == nil {
			// Hapus cookie kedaluwarsa.
			http.SetCookie(w, &http.Cookie{
				Name:     namaCookie,
				Value:    "",
				MaxAge:   -1,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Path:     "/",
			})
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r.WithContext(web.DenganPengguna(r.Context(), pengguna)))
	})
}

// MiddlewarePeran membatasi akses berdasarkan peran.
// Bila pengguna tidak memiliki peran yang diizinkan, kembalikan 403.
func MiddlewarePeran(peran ...domain.Peran) func(http.Handler) http.Handler {
	diizinkan := make(map[domain.Peran]bool, len(peran))
	for _, p := range peran {
		diizinkan[p] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pengguna := web.PenggunaDariContext(r.Context())
			if pengguna == nil || !diizinkan[pengguna.Role] {
				http.Error(w, "Anda tidak memiliki akses ke halaman ini.", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
