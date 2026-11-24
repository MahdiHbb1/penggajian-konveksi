// Pemilik: Fadli | Modul 6 | Aturan: akses Admin dan Pemilik
package auth

import (
	"net/http"
	"strings"

	"penggajian-konveksi/internal/domain"
	"penggajian-konveksi/internal/web"
)

const namaCookie = "sesi_penggajian"

// adalahRutePublik menentukan apakah path tidak memerlukan login.
// /login dan semua path di bawah /static/ bisa diakses tanpa sesi.
func adalahRutePublik(path string) bool {
	return path == "/login" || strings.HasPrefix(path, "/static/")
}

// MiddlewareAuth memasang pengguna pada context request.
// Semua halaman kecuali /login dan /static/ memerlukan sesi sah.
// Bila tidak ada sesi, redirect ke /login.
// Halaman yang memerlukan login akan mendapatkan CSRF token untuk form POST.
func MiddlewareAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if adalahRutePublik(r.URL.Path) {
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
			http.SetCookie(w, buatCookieSesi("", -1))
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Siapkan CSRF token untuk halaman yang sudah login.
		token := ambilTokenCSRF(w, r)
		ctx := web.DenganPengguna(r.Context(), pengguna)
		ctx = web.DenganCSRFToken(ctx, token)

		next.ServeHTTP(w, r.WithContext(ctx))
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
