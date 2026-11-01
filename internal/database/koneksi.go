// Pemilik: Mahdi | Modul 1 | Aturan: tidak ada aturan bisnis
package database

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Buka membuka koneksi MySQL tanpa menggagalkan aplikasi bila server mati.
//
// Masukan: DSN MySQL.
// Hasil: *sql.DB dan error hanya bila driver gagal dibuka.
// Aturan: proyek harus bisa dijalankan tanpa MySQL.
// Tabel: tidak ada.
// Memanggil: sql.Open, db.PingContext.
func Buka(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, batal := context.WithTimeout(context.Background(), 2*time.Second)
	defer batal()

	if err := db.PingContext(ctx); err != nil {
		log.Printf("PERINGATAN: MySQL belum tersedia (%v). Aplikasi tetap berjalan dalam mode kerangka.", err)
		return db, nil
	}

	log.Println("Koneksi MySQL tersedia.")
	return db, nil
}
