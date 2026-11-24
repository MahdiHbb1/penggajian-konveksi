// Pemilik: Fadli | Modul 6 | Aturan: 12, 13
package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"sync"
	"time"

	"penggajian-konveksi/internal/domain"
)

// durasisesi adalah masa berlaku sesi login.
const durasisesi = 8 * time.Hour

// intervalBersihSesi adalah seberapa sering goroutine pembersih berjalan.
const intervalBersihSesi = 30 * time.Minute

// entriSesi menyimpan data pengguna dan waktu kedaluwarsa.
type entriSesi struct {
	pengguna    domain.Pengguna
	kedaluwarsa time.Time
}

var (
	muSesi    sync.RWMutex
	storeSesi = make(map[string]entriSesi)
)

// BuatSesi membuat ID sesi acak dan menyimpan pengguna di store.
func BuatSesi(pengguna domain.Pengguna) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	id := hex.EncodeToString(b)

	muSesi.Lock()
	storeSesi[id] = entriSesi{
		pengguna:    pengguna,
		kedaluwarsa: time.Now().Add(durasisesi),
	}
	muSesi.Unlock()
	return id, nil
}

// AmbilSesi mengembalikan pengguna dari ID sesi; nil bila tidak ada atau kedaluwarsa.
func AmbilSesi(id string) *domain.Pengguna {
	muSesi.RLock()
	entri, ada := storeSesi[id]
	muSesi.RUnlock()
	if !ada || time.Now().After(entri.kedaluwarsa) {
		if ada {
			HapusSesi(id)
		}
		return nil
	}
	p := entri.pengguna
	return &p
}

// HapusSesi menghapus sesi dari store.
func HapusSesi(id string) {
	muSesi.Lock()
	delete(storeSesi, id)
	muSesi.Unlock()
}

// bersihkanSesiKedaluwarsa menghapus semua entri sesi yang sudah kedaluwarsa.
// Dipanggil secara periodik oleh MulaiCleanupSesi.
func bersihkanSesiKedaluwarsa() {
	sekarang := time.Now()
	muSesi.Lock()
	for id, entri := range storeSesi {
		if sekarang.After(entri.kedaluwarsa) {
			delete(storeSesi, id)
		}
	}
	muSesi.Unlock()
}

// MulaiCleanupSesi memulai goroutine yang membersihkan sesi kedaluwarsa secara periodik.
// Goroutine berhenti ketika ctx dibatalkan (saat server shutdown).
// Panggil dari main() setelah server siap.
func MulaiCleanupSesi(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(intervalBersihSesi)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				bersihkanSesiKedaluwarsa()
				hapusTokenCSRFKedaluwarsa()
				bersihkanRateLimitKedaluwarsa()
				log.Printf("cleanup: sesi, csrf, dan rate limit dibersihkan")
			case <-ctx.Done():
				log.Printf("cleanup sesi: goroutine berhenti")
				return
			}
		}
	}()
}
