// Pemilik: Fadli | Modul 6 | Aturan: 12, 13
package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"sync"
	"time"

	"penggajian-konveksi/internal/domain"
)

// durasisesi adalah masa berlaku sesi login.
const durasisesi = 8 * time.Hour

// intervalBersihSesi adalah seberapa sering goroutine pembersih berjalan.
const intervalBersihSesi = 30 * time.Minute

// maxSesiAktif membatasi jumlah sesi in-memory agar tidak terjadi kebocoran memori.
const maxSesiAktif = 10000

// entriSesi menyimpan data pengguna, token CSRF terikat sesi, dan waktu kedaluwarsa.
type entriSesi struct {
	pengguna    domain.Pengguna
	csrfToken   string
	kedaluwarsa time.Time
}

var (
	muSesi    sync.RWMutex
	storeSesi = make(map[string]entriSesi)
)

func buatTokenAcakHex() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// BuatSesi membuat ID sesi acak beserta token CSRF terikat sesi di store.
func BuatSesi(pengguna domain.Pengguna) (string, error) {
	id, err := buatTokenAcakHex()
	if err != nil {
		return "", err
	}
	csrfToken, err := buatTokenAcakHex()
	if err != nil {
		return "", err
	}

	sekarang := time.Now()
	muSesi.Lock()
	defer muSesi.Unlock()

	if len(storeSesi) >= maxSesiAktif {
		for sid, e := range storeSesi {
			if sekarang.After(e.kedaluwarsa) {
				delete(storeSesi, sid)
			}
		}
		if len(storeSesi) >= maxSesiAktif {
			return "", errors.New("kapasitas sesi aktif penuh")
		}
	}

	storeSesi[id] = entriSesi{
		pengguna:    pengguna,
		csrfToken:   csrfToken,
		kedaluwarsa: sekarang.Add(durasisesi),
	}
	return id, nil
}

// ambilEntriSesi mengembalikan pengguna dan token CSRF terikat sesi; nil bila tidak valid/kedaluwarsa.
func ambilEntriSesi(id string) (*domain.Pengguna, string) {
	if id == "" {
		return nil, ""
	}
	muSesi.RLock()
	entri, ada := storeSesi[id]
	muSesi.RUnlock()
	if !ada || time.Now().After(entri.kedaluwarsa) {
		if ada {
			HapusSesi(id)
		}
		return nil, ""
	}
	p := entri.pengguna
	return &p, entri.csrfToken
}

// AmbilSesi mengembalikan pengguna dari ID sesi; nil bila tidak ada atau kedaluwarsa.
func AmbilSesi(id string) *domain.Pengguna {
	p, _ := ambilEntriSesi(id)
	return p
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

// MulaiCleanupSesi memulai goroutine yang membersihkan sesi dan rate limit kedaluwarsa secara periodik.
// Goroutine berhenti ketika ctx dibatalkan (saat server shutdown).
func MulaiCleanupSesi(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(intervalBersihSesi)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				bersihkanSesiKedaluwarsa()
				bersihkanRateLimitKedaluwarsa()
				log.Printf("cleanup: sesi dan rate limit kedaluwarsa dibersihkan")
			case <-ctx.Done():
				log.Printf("cleanup sesi: goroutine berhenti")
				return
			}
		}
	}()
}
