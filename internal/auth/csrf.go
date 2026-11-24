// Pemilik: Fadli | Modul 6 | Aturan: 12, 13
// csrf.go menyediakan proteksi CSRF sederhana berbasis token sinkronisasi.
// Token dibuat saat login, disimpan di store in-memory, dan diverifikasi
// di setiap POST yang mengubah data.
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

const (
	namaCookieCSRF = "csrf_token"
	namaFieldCSRF  = "csrf_token"
	durasiCSRF     = 8 * time.Hour
)

var (
	muCSRF    sync.RWMutex
	storeCSRF = make(map[string]time.Time) // token → kedaluwarsa
)

// buatTokenCSRF menghasilkan token CSRF baru, menyimpannya di store,
// dan menyetel cookie CSRF (tidak HttpOnly agar JS bisa membacanya jika perlu).
func buatTokenCSRF(w http.ResponseWriter) string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	token := hex.EncodeToString(b)

	muCSRF.Lock()
	storeCSRF[token] = time.Now().Add(durasiCSRF)
	muCSRF.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     namaCookieCSRF,
		Value:    token,
		HttpOnly: false, // harus bisa dibaca form
		SameSite: http.SameSiteLaxMode,
		Secure:   secureCookie,
		Path:     "/",
		MaxAge:   int(durasiCSRF.Seconds()),
	})
	return token
}

// ambilTokenCSRF mengambil token dari cookie CSRF request (untuk embed ke form).
// Jika tidak ada, buat yang baru.
func ambilTokenCSRF(w http.ResponseWriter, r *http.Request) string {
	if cookie, err := r.Cookie(namaCookieCSRF); err == nil {
		muCSRF.RLock()
		exp, ada := storeCSRF[cookie.Value]
		muCSRF.RUnlock()
		if ada && time.Now().Before(exp) {
			return cookie.Value
		}
	}
	return buatTokenCSRF(w)
}

// periksaCSRF memvalidasi token CSRF dari form POST.
// Mengembalikan true jika token valid.
func periksaCSRF(r *http.Request) bool {
	formToken := r.FormValue(namaFieldCSRF)
	if formToken == "" {
		return false
	}
	muCSRF.RLock()
	exp, ada := storeCSRF[formToken]
	muCSRF.RUnlock()
	return ada && time.Now().Before(exp)
}

// hapusTokenCSRFKedaluwarsa menghapus token yang sudah kedaluwarsa.
// Dipanggil oleh MulaiCleanupSesi secara periodik.
func hapusTokenCSRFKedaluwarsa() {
	sekarang := time.Now()
	muCSRF.Lock()
	for token, exp := range storeCSRF {
		if sekarang.After(exp) {
			delete(storeCSRF, token)
		}
	}
	muCSRF.Unlock()
}
