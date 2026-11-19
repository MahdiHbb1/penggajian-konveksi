// Pemilik: Fadli | Modul 6 | Aturan: 12
// Tool satu kali untuk generate hash seed akun. Tidak disertakan di build produksi.
//go:build ignore

package main

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
)

func hashPassword(password, salt string) string {
	h := sha512.New()
	h.Write([]byte(salt + password))
	return hex.EncodeToString(h.Sum(nil))
}

func hashPasswordBaru(password string) string {
	h := sha512.New()
	h.Write([]byte(fmt.Sprintf("salt-%s", password)))
	salt := hex.EncodeToString(h.Sum(nil))[:16]
	return fmt.Sprintf("sha512:%s:%s", salt, hashPassword(password, salt))
}

func main() {
	fmt.Printf("admin:  %s\n", hashPasswordBaru("admin123"))
	fmt.Printf("pemilik: %s\n", hashPasswordBaru("pemilik123"))
}
