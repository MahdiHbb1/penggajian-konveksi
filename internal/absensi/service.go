// Pemilik: Zahra | Modul 3 | Aturan: 2, 3, 4, 5, 6, 14
package absensi

import (
	"context"
	"errors"
	"time"

	"penggajian-konveksi/internal/domain"
)

var (
	ErrJadwalTidakDitemukan  = errors.New("jadwal shift tidak ditemukan")
	ErrAbsensiGanda          = errors.New("absensi untuk jadwal ini sudah tercatat")
	ErrAbsensiTidakDitemukan = errors.New("data absensi tidak ditemukan")
	ErrPeriodeTerkunci       = errors.New("absensi tidak dapat diubah karena periodenya sudah ditutup")
	ErrJamMasukTidakValid    = errors.New("jam masuk wajib diisi dengan waktu yang valid")
	ErrJamKeluarTidakValid   = errors.New("jam keluar harus setelah jam masuk")
	ErrTanggalTidakSesuai    = errors.New("tanggal absensi harus sesuai dengan tanggal jadwal")
	ErrDataTidakValid        = errors.New("data absensi tidak valid")
)

// Service mengatur absensi dan hitungan menit.
type Service struct {
	repo *Repository
}

// NewService membuat service absensi.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CatatAbsensi mencatat absensi untuk jadwal yang ada secara atomik (mengunci jadwal_shift).
func (s *Service) CatatAbsensi(ctx context.Context, data domain.Absensi) (int64, error) {
	if s.repo == nil || data.JadwalShiftID <= 0 {
		return 0, ErrDataTidakValid
	}
	if data.JamMasuk.IsZero() {
		return 0, ErrJamMasukTidakValid
	}

	var err error
	data.JamKeluar, err = normalisasiJamKeluar(data.JamMasuk, data.JamKeluar)
	if err != nil {
		return 0, err
	}
	if err := s.repo.SimpanAbsensiAtomik(ctx, &data); err != nil {
		return 0, err
	}
	return data.ID, nil
}

// UbahAbsensi memperbarui jam masuk dan jam keluar secara atomik.
func (s *Service) UbahAbsensi(ctx context.Context, data domain.Absensi) error {
	if s.repo == nil || data.ID <= 0 {
		return ErrDataTidakValid
	}
	if data.JamMasuk.IsZero() {
		return ErrJamMasukTidakValid
	}

	var err error
	data.JamKeluar, err = normalisasiJamKeluar(data.JamMasuk, data.JamKeluar)
	if err != nil {
		return err
	}
	return s.repo.UbahAbsensiAtomik(ctx, data)
}

// DaftarAbsensi mengambil absensi yang sudah tercatat pada tanggal tertentu.
func (s *Service) DaftarAbsensi(ctx context.Context, tanggal time.Time) ([]domain.AbsensiLengkap, error) {
	return s.repo.DaftarAbsensi(ctx, tanggal)
}

// DaftarJadwalAbsensi mengambil seluruh jadwal tanggal tertentu, termasuk yang belum diisi absensinya.
// Fungsi ini dipakai halaman input agar semua pekerja yang terjadwal tampil.
func (s *Service) DaftarJadwalAbsensi(ctx context.Context, tanggal time.Time) ([]domain.AbsensiLengkap, error) {
	return s.repo.DaftarJadwalAbsensi(ctx, tanggal)
}

// DaftarAbsensiBelumLengkap mengambil absensi tanpa jam keluar pada periode.
func (s *Service) DaftarAbsensiBelumLengkap(ctx context.Context, periodeID int64) ([]domain.AbsensiLengkap, error) {
	return s.repo.DaftarBelumLengkap(ctx, periodeID)
}

// AbsensiDalamPeriode mengambil hanya absensi lengkap pada periode.
func (s *Service) AbsensiDalamPeriode(ctx context.Context, periodeID int64) ([]domain.AbsensiLengkap, error) {
	return s.repo.AbsensiDalamPeriode(ctx, periodeID)
}

// AdaAbsensiBelumLengkap mengecek jumlah absensi tanpa jam keluar pada periode.
func (s *Service) AdaAbsensiBelumLengkap(ctx context.Context, periodeID int64) (bool, int, error) {
	return s.repo.AdaBelumLengkap(ctx, periodeID)
}

// DaftarPeriode mengambil periode untuk pilihan pada halaman absensi belum lengkap.
func (s *Service) DaftarPeriode(ctx context.Context) ([]domain.PeriodeGaji, error) {
	return s.repo.DaftarPeriode(ctx)
}

// HitungMenitTerlambat menghitung selisih jam masuk terhadap jam mulai shift.
// Hasil dibulatkan ke bawah ke menit penuh dan tidak pernah negatif.
//
// Masukan: jam masuk dan jam mulai shift.
// Hasil: menit terlambat, minimal 0.
// Aturan: 4, 5.
// Tabel: tidak ada.
// Memanggil: tidak ada.
func (s *Service) HitungMenitTerlambat(jamMasuk time.Time, jamMulaiShift time.Time) int {
	return menitPositif(jamMasuk.Sub(jamMulaiShift))
}

// HitungMenitLembur menghitung selisih jam keluar terhadap jam selesai shift.
// Jam keluar pada hari berikutnya tetap dihitung benar selama time.Time memuat tanggalnya.
func (s *Service) HitungMenitLembur(jamKeluar time.Time, jamSelesaiShift time.Time) int {
	return menitPositif(jamKeluar.Sub(jamSelesaiShift))
}

func menitPositif(durasi time.Duration) int {
	if durasi <= 0 {
		return 0
	}
	return int(durasi / time.Minute)
}

// normalisasiJamKeluar menganggap jam keluar yang lebih kecil dari jam masuk sebagai
// waktu pada hari berikutnya. Ini diperlukan untuk shift yang melintasi tengah malam.
func normalisasiJamKeluar(jamMasuk time.Time, jamKeluar *time.Time) (*time.Time, error) {
	if jamKeluar == nil {
		return nil, nil
	}
	if jamKeluar.IsZero() {
		return nil, ErrJamKeluarTidakValid
	}
	keluar := *jamKeluar
	if keluar.Before(jamMasuk) {
		keluar = keluar.Add(24 * time.Hour)
	}
	if !keluar.After(jamMasuk) {
		return nil, ErrJamKeluarTidakValid
	}
	return &keluar, nil
}

func tanggalSama(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// Pastikan Service memenuhi kontrak lintas modul.
var _ interface {
	AbsensiDalamPeriode(context.Context, int64) ([]domain.AbsensiLengkap, error)
	AdaAbsensiBelumLengkap(context.Context, int64) (bool, int, error)
	HitungMenitTerlambat(time.Time, time.Time) int
	HitungMenitLembur(time.Time, time.Time) int
} = (*Service)(nil)
