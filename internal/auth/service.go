// Pemilik: Fadli | Modul 6 | Aturan: 12, 13
package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"penggajian-konveksi/internal/domain"
)

// ErrCredentialSalah dikembalikan bila username atau password tidak cocok.
var ErrCredentialSalah = errors.New("username atau password salah")

// biayaBcrypt adalah cost factor bcrypt yang digunakan untuk hashing password.
const biayaBcrypt = 12

// Service menangani login, logout, dan sesi.
type Service struct {
	db *sql.DB
}

// NewService membuat service autentikasi.
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// HashPasswordBaru menghasilkan hash bcrypt dari password teks biasa.
// Gunakan fungsi ini untuk menyimpan password baru ke database.
// Jangan pernah menyimpan atau mencatat password teks biasa.
func HashPasswordBaru(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), biayaBcrypt)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// Login memeriksa username dan password.
//
// Masukan: context, username, password teks biasa.
// Hasil: Pengguna yang cocok.
// Aturan: hanya pengguna berperan Admin atau Pemilik.
// Tabel: pengguna.
// Memanggil: database/sql QueryRowContext, bcrypt.CompareHashAndPassword.
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
	if err := bcrypt.CompareHashAndPassword([]byte(p.PasswordHash), []byte(password)); err != nil {
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
