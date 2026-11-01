// Pemilik: Fadli | Modul 6 | Aturan: 12, 13
package auth

import (
	"context"
	"database/sql"

	"penggajian-konveksi/internal/domain"
)

// Service menangani login, logout, dan sesi.
type Service struct {
	db *sql.DB
}

// NewService membuat service autentikasi.
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// Login memeriksa username dan password.
//
// Masukan: context, username, password teks biasa.
// Hasil: Pengguna yang cocok.
// Aturan: hanya pengguna berperan Admin atau Pemilik.
// Tabel: pengguna.
// Memanggil: database/sql QueryRowContext, bcrypt.CompareHashAndPassword.
//
// TODO(Fadli):
//  1. Ambil pengguna berdasarkan username.
//  2. Bandingkan password dengan hash bcrypt.
//  3. Jangan pernah menyimpan atau mencatat password teks biasa.
func (s *Service) Login(ctx context.Context, username, password string) (domain.Pengguna, error) {
	return domain.Pengguna{}, domain.ErrBelumDiimplementasi
}

// Logout menghapus sesi pengguna.
//
// Masukan: context dan ID sesi.
// Hasil: error.
// Aturan: sesi memakai cookie HttpOnly.
// Tabel: tidak ada (sesi disimpan di cookie atau store sesi).
// Memanggil: penyimpanan sesi yang dipilih.
//
// TODO(Fadli):
//  1. Hapus sesi aktif.
//  2. Kosongkan cookie HttpOnly.
func (s *Service) Logout(ctx context.Context, sessionID string) error {
	return domain.ErrBelumDiimplementasi
}
