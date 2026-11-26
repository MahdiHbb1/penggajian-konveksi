// Pemilik: Fadli | Modul 6 | Aturan: akses Admin dan Pemilik
package auth

import (
	"log"
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

func adalahMetodeMutasi(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func pasangHeaderKeamanan(w http.ResponseWriter) {
	h := w.Header()
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("X-Frame-Options", "DENY")
	h.Set("Referrer-Policy", "same-origin")
}

// MiddlewareAuth memasang pengguna & CSRF token terikat sesi pada context request,
// menetapkan header keamanan dasar, serta memvalidasi CSRF untuk semua permintaan mutasi.
// Semua halaman kecuali /login dan /static/ memerlukan sesi sah.
func MiddlewareAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pasangHeaderKeamanan(w)

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

		pengguna, csrfToken := ambilEntriSesi(cookie.Value)
		if pengguna == nil {
			// Hapus cookie kedaluwarsa.
			http.SetCookie(w, buatCookieSesi("", -1))
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Validasi CSRF terpusat untuk semua request mutasi terautentikasi (fail-closed).
		if adalahMetodeMutasi(r.Method) {
			if err := r.ParseForm(); err != nil {
				log.Printf("auth middleware: gagal membaca form %s %s dari %s: %v", r.Method, r.URL.Path, r.RemoteAddr, err)
				http.Error(w, "Permintaan formulir tidak valid.", http.StatusForbidden)
				return
			}
			if !cocokTokenCSRF(csrfToken, r.PostFormValue(namaFieldCSRF)) {
				log.Printf("auth middleware: CSRF token tidak valid pada %s %s dari %s", r.Method, r.URL.Path, r.RemoteAddr)
				http.Error(w, "Sesi formulir tidak valid. Silakan muat ulang halaman dan coba lagi.", http.StatusForbidden)
				return
			}
		}

		ctx := web.DenganPengguna(r.Context(), pengguna)
		ctx = web.DenganCSRFToken(ctx, csrfToken)

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
