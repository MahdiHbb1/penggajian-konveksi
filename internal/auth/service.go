// Pemilik: Fadli | Modul 6 | Aturan: 12, 13
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"penggajian-konveksi/internal/domain"
)

// ErrCredentialSalah dikembalikan bila username atau password tidak cocok.
var ErrCredentialSalah = errors.New("username atau password salah")

// Service menangani login, logout, dan sesi.
type Service struct {
	db *sql.DB
}

// NewService membuat service autentikasi.
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// hashPassword menghasilkan hash SHA-512 dengan salt untuk password.
// Format simpanan: "sha512:{salt}:{hash}".
// Fungsi ini dipakai saat menyimpan password baru ke database.
func hashPassword(password, salt string) string {
	h := sha512.New()
	h.Write([]byte(salt + password))
	return hex.EncodeToString(h.Sum(nil))
}

// periksaPassword memverifikasi password teks biasa terhadap hash tersimpan.
// Format hash: "sha512:{salt}:{hash}".
func periksaPassword(passwordTertulis, hashTersimpan string) bool {
	bagian := strings.SplitN(hashTersimpan, ":", 3)
	if len(bagian) != 3 || bagian[0] != "sha512" {
		return false
	}
	salt := bagian[1]
	hashExpected := bagian[2]
	hashActual := hashPassword(passwordTertulis, salt)
	// subtle.ConstantTimeCompare untuk menghindari timing attack.
	return subtle.ConstantTimeCompare([]byte(hashActual), []byte(hashExpected)) == 1
}

// HashPasswordBaru menghasilkan hash password dengan salt acak.
func HashPasswordBaru(password string) string {
	saltBytes := make([]byte, 8)

	if _, err := rand.Read(saltBytes); err != nil {
		panic(fmt.Errorf("gagal membuat salt: %w", err))
	}

	salt := hex.EncodeToString(saltBytes)

	return fmt.Sprintf("sha512:%s:%s", salt, hashPassword(password, salt))
}

// Login memeriksa username dan password.
//
// Masukan: context, username, password teks biasa.
// Hasil: Pengguna yang cocok.
// Aturan: hanya pengguna berperan Admin atau Pemilik.
// Tabel: pengguna.
// Memanggil: database/sql QueryRowContext.
func (s *Service) Login(ctx context.Context, username, password string) (domain.Pengguna, error) {
	if s.db == nil {
		return domain.Pengguna{}, ErrCredentialSalah
	}
	var p domain.Pengguna
	err := s.db.QueryRowContext(ctx,
		`SELECT id, username, password_hash, role FROM pengguna WHERE username = ? LIMIT 1`,
		username,
	).Scan(&p.ID, &p.Username, &p.PasswordHash, &p.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Pengguna{}, ErrCredentialSalah
	}
	if err != nil {
		return domain.Pengguna{}, fmt.Errorf("query pengguna: %w", err)
	}
	if !periksaPassword(password, p.PasswordHash) {
		return domain.Pengguna{}, ErrCredentialSalah
	}
	// Jangan kembalikan hash ke pemanggil.
	p.PasswordHash = ""
	return p, nil
}

// Logout menghapus sesi pengguna dari store sesi global.
//
// Masukan: context dan ID sesi.
// Hasil: error.
// Aturan: sesi memakai cookie HttpOnly.
func (s *Service) Logout(ctx context.Context, sessionID string) error {
	HapusSesi(sessionID)
	return nil
}
