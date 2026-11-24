// Pemilik: Mahdi | Modul 1 | Aturan: 1, 8, 12, 13, 15
package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"

	"penggajian-konveksi/internal/absensi"
	"penggajian-konveksi/internal/auth"
	"penggajian-konveksi/internal/config"
	"penggajian-konveksi/internal/database"
	"penggajian-konveksi/internal/domain"
	"penggajian-konveksi/internal/jadwal"
	"penggajian-konveksi/internal/kasbon"
	"penggajian-konveksi/internal/laporan"
	"penggajian-konveksi/internal/pekerja"
	"penggajian-konveksi/internal/penggajian"
	"penggajian-konveksi/internal/periode"
	"penggajian-konveksi/internal/web"
)

// bangunHandler menyusun seluruh service, rute, pembatasan peran (RBAC), dan autentikasi.
// Rute Modul 1-5 dibungkus MiddlewarePeran(PeranAdmin) via sub-mux terpisah;
// rute laporan (dashboard, slip, rekap) mengizinkan Admin dan Pemilik;
// rute /logout tersedia untuk semua pengguna terautentikasi; /login dan /static/ publik.
func bangunHandler(db *sql.DB, renderer web.Renderer, staticDir string) http.Handler {
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

	// Sambungkan PeriodeProvider Modul 5 ke Modul 2 agar kontrak kunci periode aktif penuh.
	svcJadwal.AturPeriode(svcPeriode)
	svcKasbon.AturPeriode(svcPeriode)

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

	hanyaAdmin := auth.MiddlewarePeran(domain.PeranAdmin)

	router := web.Router(
		renderer,
		staticDir,
		func(mux *http.ServeMux, r web.Renderer) {
			adminMux := http.NewServeMux()
			pekerja.DaftarkanRute(adminMux, r, svcPekerja)
			jadwal.DaftarkanRute(adminMux, r, svcJadwal)
			kasbon.DaftarkanRute(adminMux, r, svcKasbon)
			absensi.DaftarkanRute(adminMux, r, svcAbsensi)
			penggajian.DaftarkanRute(adminMux, r, svcPenggajian)
			periode.DaftarkanRute(adminMux, r, svcPeriode)

			dilindungiAdmin := hanyaAdmin(adminMux)
			for _, prefix := range []string{
				"/pekerja", "/pekerja/",
				"/jadwal", "/jadwal/",
				"/kasbon", "/kasbon/",
				"/absensi", "/absensi/",
				"/penggajian", "/penggajian/",
				"/periode", "/periode/",
				"/koreksi", "/koreksi/",
			} {
				mux.Handle(prefix, dilindungiAdmin)
			}
		},
		// Laporan (Modul 6): dashboard, slip, dan rekap untuk Admin dan Pemilik.
		func(mux *http.ServeMux, r web.Renderer) {
			laporan.DaftarkanRute(mux, r, svcLaporan, auth.MiddlewarePeran)
		},
		func(mux *http.ServeMux, r web.Renderer) {
			auth.DaftarkanRute(mux, r, svcAuth)
		},
	)

	return auth.MiddlewareAuth(router)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Peringatan: file .env tidak ditemukan: %v", err)
	}
	cfg := config.Muat()
	auth.SetSecureCookie(cfg.SecureCookie)

	db, err := database.Buka(cfg.DSN)
	if err != nil {
		log.Fatalf("Gagal menyiapkan koneksi database: %v", err)
	}
	defer db.Close()

	renderer := &web.Render{Dir: cfg.TemplateDir}
	handler := bangunHandler(db, renderer, cfg.StaticDir)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	cleanupCtx, cancelCleanup := context.WithCancel(context.Background())
	defer cancelCleanup()
	auth.MulaiCleanupSesi(cleanupCtx)

	log.Printf("Sistem Penggajian berjalan di http://localhost:%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server berhenti: %v", err)
	}
}
