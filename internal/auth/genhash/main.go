// Pemilik: Fadli | Modul 6 | Aturan: 12
// Tool satu kali untuk generate hash bcrypt akun. Tidak disertakan di build produksi.
//go:build ignore

package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

const biayaBcrypt = 12

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Penggunaan: go run main.go <password>")
		fmt.Fprintln(os.Stderr, "Contoh:     go run main.go rahasia123")
		os.Exit(1)
	}
	password := os.Args[1]
	hash, err := bcrypt.GenerateFromPassword([]byte(password), biayaBcrypt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Gagal membuat hash: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(hash))
}
