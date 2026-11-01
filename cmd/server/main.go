// Pemilik: Mahdi | Modul 1 | Aturan: 1, 8, 12, 13, 15
package main

import (
	"log"
	"net/http"
	"time"

	"penggajian-konveksi/internal/absensi"
	"penggajian-konveksi/internal/auth"
	"penggajian-konveksi/internal/config"
	"penggajian-konveksi/internal/database"
	"penggajian-konveksi/internal/jadwal"
	"penggajian-konveksi/internal/kasbon"
	"penggajian-konveksi/internal/laporan"
	"penggajian-konveksi/internal/pekerja"
	"penggajian-konveksi/internal/penggajian"
	"penggajian-konveksi/internal/periode"
	"penggajian-konveksi/internal/web"
)

func main() {
	cfg := config.Muat()

	db, err := database.Buka(cfg.DSN)
	if err != nil {
		log.Fatalf("Gagal menyiapkan koneksi database: %v", err)
	}
	defer db.Close()

	repoPekerja := pekerja.NewRepository(db)
	repoJadwal := jadwal.NewRepository(db)
	repoKasbon := kasbon.NewRepository(db)
	repoAbsensi := absensi.NewRepository(db)
	repoPeriode := periode.NewRepository(db)
	repoPenggajian := penggajian.NewRepository(db)
	repoLaporan := laporan.NewRepository(db)

	svcPekerja := pekerja.NewService(repoPekerja)
	svcJadwal := jadwal.NewService(repoJadwal)
	svcKasbon := kasbon.NewService(repoKasbon)
	svcAbsensi := absensi.NewService(repoAbsensi)
	svcPeriode := periode.NewService(repoPeriode)
	svcPenggajian := penggajian.NewService(
		repoPenggajian,
		svcPekerja,
		svcJadwal,
		svcKasbon,
		svcAbsensi,
		svcPeriode,
		svcPeriode,
	)
	svcLaporan := laporan.NewService(repoLaporan)
	svcAuth := auth.NewService(db)

	renderer := &web.Render{Dir: cfg.TemplateDir}
	handler := web.Router(
		renderer,
		cfg.StaticDir,
		func(mux *http.ServeMux, r web.Renderer) { pekerja.DaftarkanRute(mux, r, svcPekerja) },
		func(mux *http.ServeMux, r web.Renderer) { jadwal.DaftarkanRute(mux, r, svcJadwal) },
		func(mux *http.ServeMux, r web.Renderer) { kasbon.DaftarkanRute(mux, r, svcKasbon) },
		func(mux *http.ServeMux, r web.Renderer) { absensi.DaftarkanRute(mux, r, svcAbsensi) },
		func(mux *http.ServeMux, r web.Renderer) { penggajian.DaftarkanRute(mux, r, svcPenggajian) },
		func(mux *http.ServeMux, r web.Renderer) { periode.DaftarkanRute(mux, r, svcPeriode) },
		func(mux *http.ServeMux, r web.Renderer) { laporan.DaftarkanRute(mux, r, svcLaporan) },
		func(mux *http.ServeMux, r web.Renderer) { auth.DaftarkanRute(mux, r, svcAuth) },
	)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           auth.MiddlewareAuth(handler),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Kerangka Sistem Penggajian berjalan di http://localhost:%s", cfg.Port)
	log.Printf("Mode kerangka: autentikasi masih placeholder, MySQL tidak wajib menyala.")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server berhenti: %v", err)
	}
}
