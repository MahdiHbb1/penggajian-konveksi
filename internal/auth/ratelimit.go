// Pemilik: Fadli | Modul 6 | Aturan: 12, 13
// ratelimit.go membatasi percobaan login gagal per username.
// Setelah maxPercobaanLogin kali gagal dalam jendelaRateLimit,
// username dikunci selama durasiKunci.
package auth

import (
	"sync"
	"time"
)

const (
	maxPercobaanLogin = 5
	jendelaRateLimit  = 5 * time.Minute
	durasiKunci       = 15 * time.Minute
)

type statusRateLimit struct {
	percobaan   int
	pertama     time.Time
	dikunciSamp time.Time
}

var (
	muRL    sync.Mutex
	storeRL = make(map[string]*statusRateLimit)
)

// cekRateLimit mengembalikan true jika username sedang dikunci.
func cekRateLimit(username string) bool {
	muRL.Lock()
	defer muRL.Unlock()

	s, ada := storeRL[username]
	if !ada {
		return false
	}
	// Cek apakah masih dalam masa kunci.
	if !s.dikunciSamp.IsZero() && time.Now().Before(s.dikunciSamp) {
		return true
	}
	// Reset jika jendela sudah lewat.
	if time.Since(s.pertama) > jendelaRateLimit {
		delete(storeRL, username)
		return false
	}
	return false
}

// catatGagalLogin mencatat satu percobaan login gagal.
// Jika sudah mencapai maxPercobaanLogin, kunci username.
func catatGagalLogin(username string) {
	muRL.Lock()
	defer muRL.Unlock()

	s, ada := storeRL[username]
	if !ada || time.Since(s.pertama) > jendelaRateLimit {
		storeRL[username] = &statusRateLimit{
			percobaan: 1,
			pertama:   time.Now(),
		}
		return
	}
	s.percobaan++
	if s.percobaan >= maxPercobaanLogin {
		s.dikunciSamp = time.Now().Add(durasiKunci)
	}
}

// resetRateLimit menghapus catatan percobaan login saat berhasil login.
func resetRateLimit(username string) {
	muRL.Lock()
	delete(storeRL, username)
	muRL.Unlock()
}

// bersihkanRateLimitKedaluwarsa menghapus entri rate limit yang sudah tidak relevan.
func bersihkanRateLimitKedaluwarsa() {
	muRL.Lock()
	sekarang := time.Now()
	for u, s := range storeRL {
		// Hapus jika jendela sudah lewat dan tidak sedang dikunci, atau masa kunci sudah lewat.
		jendela := time.Since(s.pertama) > jendelaRateLimit
		kunciLewat := !s.dikunciSamp.IsZero() && sekarang.After(s.dikunciSamp)
		if (jendela && s.dikunciSamp.IsZero()) || kunciLewat {
			delete(storeRL, u)
		}
	}
	muRL.Unlock()
}
