// Pemilik: Fadli | Modul 6 | Aturan: tidak ada aturan bisnis
package web

import (
	"net/http"
)

// DaftarRute adalah fungsi yang mendaftarkan rute satu modul.
type DaftarRute func(mux *http.ServeMux, renderer Renderer)

// Router menyusun seluruh rute aplikasi.
//
// Masukan: renderer, direktori aset statis, dan daftar pendaftar rute.
// Hasil: http.Handler siap dipakai.
// Aturan: setiap halaman di bagian 11.2 harus terdaftar.
// Tabel: tidak ada.
// Memanggil: http.NewServeMux, http.FileServer.
func Router(renderer Renderer, staticDir string, daftar ...DaftarRute) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	})

	for _, d := range daftar {
		d(mux, renderer)
	}

	return mux
}
