// Pemilik: Fadli | Modul 6 | Aturan: 12, 13
// ratelimit.go membatasi percobaan login gagal per username (dinormalisasi case-insensitive).
// Reservasi slot dilakukan secara atomik sebelum bcrypt untuk mencegah race condition TOCTOU.
package auth

import (
	"strings"
	"sync"
	"time"
)

const (
	maxPercobaanLogin = 5
	jendelaRateLimit  = 5 * time.Minute
	durasiKunci       = 15 * time.Minute
	maxEntriRL        = 10000
)

type statusRateLimit struct {
	percobaan    int
	sedangProses int
	pertama      time.Time
	dikunciSamp  time.Time
}

var (
	muRL    sync.Mutex
	storeRL = make(map[string]*statusRateLimit)
)

// kunciRateLimit menormalisasi username agar cocok dengan kolasi case-insensitive MySQL.
func kunciRateLimit(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func evictRateLimitLocked(sekarang time.Time) {
	for u, s := range storeRL {
		if s.sedangProses > 0 {
			continue
		}
		kunciLewat := !s.dikunciSamp.IsZero() && !sekarang.Before(s.dikunciSamp)
		jendelaLewat := s.dikunciSamp.IsZero() && sekarang.Sub(s.pertama) > jendelaRateLimit
		if kunciLewat || jendelaLewat {
			delete(storeRL, u)
		}
	}
}

// mulaiPercobaanLogin memeriksa dan mereservasi satu slot percobaan login secara atomik
// sebelum menjalankan query DB dan bcrypt. Mengembalikan false bila username sedang dikunci
// atau total percobaan gagal + in-flight sudah mencapai maxPercobaanLogin.
func mulaiPercobaanLogin(username string) bool {
	kunci := kunciRateLimit(username)
	sekarang := time.Now()

	muRL.Lock()
	defer muRL.Unlock()

	s, ada := storeRL[kunci]
	if ada {
		// Masih dalam masa kunci.
		if !s.dikunciSamp.IsZero() && sekarang.Before(s.dikunciSamp) {
			return false
		}
		// Masa kunci sudah habis → reset entri jika tidak ada request in-flight.
		if !s.dikunciSamp.IsZero() && !sekarang.Before(s.dikunciSamp) && s.sedangProses == 0 {
			delete(storeRL, kunci)
			ada = false
		} else if s.dikunciSamp.IsZero() && sekarang.Sub(s.pertama) > jendelaRateLimit && s.sedangProses == 0 {
			delete(storeRL, kunci)
			ada = false
		}
	}

	if !ada {
		if len(storeRL) >= maxEntriRL {
			evictRateLimitLocked(sekarang)
			if len(storeRL) >= maxEntriRL {
				return false
			}
		}
		s = &statusRateLimit{pertama: sekarang}
		storeRL[kunci] = s
	}

	if s.percobaan+s.sedangProses >= maxPercobaanLogin {
		return false
	}
	s.sedangProses++
	return true
}

// selesaiPercobaanLogin melepaskan slot in-flight dan mencatat hasil login.
// Jika gagal dan mencapai batas, username dikunci selama durasiKunci.
// Jika gagal karena galat sistem (bukan kredensial salah), panggil dengan catatGagal=false.
func selesaiPercobaanLogin(username string, berhasil bool, catatGagal bool) {
	kunci := kunciRateLimit(username)
	sekarang := time.Now()

	muRL.Lock()
	defer muRL.Unlock()

	if berhasil {
		delete(storeRL, kunci)
		return
	}

	s, ada := storeRL[kunci]
	if !ada {
		if catatGagal {
			storeRL[kunci] = &statusRateLimit{percobaan: 1, pertama: sekarang}
		}
		return
	}

	if s.sedangProses > 0 {
		s.sedangProses--
	}
	if !catatGagal {
		if s.percobaan == 0 && s.sedangProses == 0 && s.dikunciSamp.IsZero() {
			delete(storeRL, kunci)
		}
		return
	}

	// Jangan pernah mereset kunci aktif meski jendelaRateLimit sudah lewat.
	if !s.dikunciSamp.IsZero() && sekarang.Before(s.dikunciSamp) {
		return
	}
	s.percobaan++
	if s.percobaan >= maxPercobaanLogin {
		s.dikunciSamp = sekarang.Add(durasiKunci)
	}
}

// cekRateLimit mengembalikan true jika username sedang dikunci.
func cekRateLimit(username string) bool {
	kunci := kunciRateLimit(username)
	sekarang := time.Now()

	muRL.Lock()
	defer muRL.Unlock()

	s, ada := storeRL[kunci]
	if !ada {
		return false
	}
	if !s.dikunciSamp.IsZero() && sekarang.Before(s.dikunciSamp) {
		return true
	}
	if s.sedangProses == 0 && ((!s.dikunciSamp.IsZero() && !sekarang.Before(s.dikunciSamp)) || sekarang.Sub(s.pertama) > jendelaRateLimit) {
		delete(storeRL, kunci)
	}
	return false
}

// catatGagalLogin mencatat satu percobaan login gagal (kompatibilitas tes unit).
func catatGagalLogin(username string) {
	kunci := kunciRateLimit(username)
	sekarang := time.Now()

	muRL.Lock()
	defer muRL.Unlock()

	s, ada := storeRL[kunci]
	if ada && !s.dikunciSamp.IsZero() && sekarang.Before(s.dikunciSamp) {
		// Tetap terkunci; jangan buka kunci lebih awal.
		return
	}
	if !ada || (!s.dikunciSamp.IsZero() && !sekarang.Before(s.dikunciSamp)) || (s.dikunciSamp.IsZero() && sekarang.Sub(s.pertama) > jendelaRateLimit) {
		if len(storeRL) >= maxEntriRL {
			evictRateLimitLocked(sekarang)
		}
		storeRL[kunci] = &statusRateLimit{
			percobaan: 1,
			pertama:   sekarang,
		}
		return
	}
	s.percobaan++
	if s.percobaan >= maxPercobaanLogin {
		s.dikunciSamp = sekarang.Add(durasiKunci)
	}
}

// resetRateLimit menghapus catatan percobaan login saat berhasil login.
func resetRateLimit(username string) {
	kunci := kunciRateLimit(username)
	muRL.Lock()
	delete(storeRL, kunci)
	muRL.Unlock()
}

// bersihkanRateLimitKedaluwarsa menghapus entri rate limit yang sudah tidak relevan.
func bersihkanRateLimitKedaluwarsa() {
	muRL.Lock()
	evictRateLimitLocked(time.Now())
	muRL.Unlock()
}
