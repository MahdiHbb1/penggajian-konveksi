// Pemilik: Fadli | Modul 6 | Aturan: akses Admin dan Pemilik
package auth

import (
	"context"
	"testing"

	"penggajian-konveksi/internal/domain"
)

func TestLoginMenolakPasswordSalah(t *testing.T) {
	t.Skip("belum diimplementasi: Login")
	svc := NewService(nil)
	if _, err := svc.Login(context.Background(), "admin", "salah"); err == nil {
		t.Fatal("password salah seharusnya ditolak")
	}
}

func TestLoginMenolakPenggunaTidakDikenal(t *testing.T) {
	t.Skip("belum diimplementasi: Login")
	svc := NewService(nil)
	if _, err := svc.Login(context.Background(), "tidak-ada", "apa-saja"); err == nil {
		t.Fatal("pengguna tidak dikenal seharusnya ditolak")
	}
}

func TestPasswordDisimpanSebagaiHash(t *testing.T) {
	t.Skip("belum diimplementasi: Login")
	pengguna := domain.Pengguna{Username: "admin", PasswordHash: "$2a$10$contoh-hash-bcrypt"}
	if pengguna.PasswordHash == "admin" {
		t.Fatal("password tidak boleh disimpan sebagai teks biasa")
	}
}

func TestMiddlewarePeranMenolakPemilikDiHalamanAdmin(t *testing.T) {
	t.Skip("belum diimplementasi: MiddlewarePeran")
	if _, err := NewService(nil).Login(context.Background(), "pemilik", "salah"); err == nil {
		t.Fatal("pemilik tidak boleh mendapat akses admin")
	}
}
