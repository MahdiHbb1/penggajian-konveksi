// Pemilik: Fadli | Modul 6 | Aturan: 12, 13
// csrf.go menyediakan proteksi CSRF tanpa map global tak terbatas:
//   - Halaman pra-login (/login): Double-Submit Cookie (cookie HttpOnly dicocokkan dengan form).
//   - Halaman terautentikasi (semua POST): token CSRF terikat sesi (Synchronizer Token di entriSesi).
package auth

import (
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"time"
)

const (
	namaCookieCSRF = "csrf_token"
	namaFieldCSRF  = "csrf_token"
	durasiCSRF     = 2 * time.Hour
	panjangHexCSRF = 64 // 32 byte acak dalam hex
)

func tokenHexValid(token string) bool {
	if len(token) != panjangHexCSRF {
		return false
	}
	_, err := hex.DecodeString(token)
	return err == nil
}

// cocokTokenCSRF membandingkan dua token secara constant-time dan menolak token kosong.
func cocokTokenCSRF(diharapkan, diterima string) bool {
	if diharapkan == "" || diterima == "" || len(diharapkan) != len(diterima) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(diharapkan), []byte(diterima)) == 1
}

// buatTokenCSRF menghasilkan token CSRF baru dan menyetel cookie Double-Submit (HttpOnly).
func buatTokenCSRF(w http.ResponseWriter) string {
	token, err := buatTokenAcakHex()
	if err != nil {
		return ""
	}
	http.SetCookie(w, &http.Cookie{
		Name:     namaCookieCSRF,
		Value:    token,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isSecureCookie(),
		Path:     "/",
		MaxAge:   int(durasiCSRF.Seconds()),
	})
	return token
}

// ambilTokenCSRF mengambil token dari cookie CSRF pra-login yang valid; membuat baru bila belum ada.
func ambilTokenCSRF(w http.ResponseWriter, r *http.Request) string {
	if cookie, err := r.Cookie(namaCookieCSRF); err == nil && tokenHexValid(cookie.Value) {
		return cookie.Value
	}
	return buatTokenCSRF(w)
}

// hapusCookieCSRF menghapus cookie CSRF pra-login setelah login berhasil.
func hapusCookieCSRF(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     namaCookieCSRF,
		Value:    "",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isSecureCookie(),
		Path:     "/",
		MaxAge:   -1,
	})
}

// periksaCSRF memvalidasi token CSRF pada form pra-login dengan pola Double-Submit Cookie
// (cookie csrf_token pada request wajib cocok constant-time dengan field form csrf_token).
func periksaCSRF(r *http.Request) bool {
	cookie, err := r.Cookie(namaCookieCSRF)
	if err != nil || !tokenHexValid(cookie.Value) {
		return false
	}
	formToken := r.PostFormValue(namaFieldCSRF)
	return cocokTokenCSRF(cookie.Value, formToken)
}
