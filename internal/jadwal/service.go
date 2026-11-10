// Pemilik: Edgar | Modul 2 | Aturan: 3, 14, 19
package jadwal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"penggajian-konveksi/internal/domain"
	"penggajian-konveksi/internal/kontrak"
)

// Pesan galat sesuai bagian 5.2 dokumen rancangan.
var (
	ErrJadwalGanda       = errors.New("Pekerja sudah punya jadwal pada tanggal ini.")
	ErrSudahAdaAbsensi   = errors.New("Jadwal ini sudah memiliki absensi dan tidak dapat diubah.")
	ErrHariMinggu        = errors.New("Hari Minggu tidak dapat dijadwalkan.")
	ErrPeriodeTertutup   = errors.New("Tanggal berada pada periode yang sudah ditutup.")
	ErrPekerjaTidakAktif = errors.New("Pekerja wajib dipilih dan harus berstatus aktif.")
	ErrShiftTidakValid   = errors.New("Shift tidak valid.")
	ErrTanggalKosong     = errors.New("Tanggal jadwal wajib diisi.")
	ErrJadwalTidakAda    = errors.New("Jadwal tidak ditemukan.")
)

var daftarGalatInput = []error{
	ErrJadwalGanda, ErrSudahAdaAbsensi, ErrHariMinggu, ErrPeriodeTertutup,
	ErrPekerjaTidakAktif, ErrShiftTidakValid, ErrTanggalKosong,
}

// PesanGalatInput mengembalikan pesan bila err adalah galat input pengguna.
// Hasil kosong berarti galat sistem yang tidak boleh dibuka ke pengguna.
func PesanGalatInput(err error) string {
	for _, e := range daftarGalatInput {
		if errors.Is(err, e) {
			return e.Error()
		}
	}
	return ""
}

// PenguncianPeriode adalah bagian PeriodeProvider (Modul 5) yang kita pakai.
type PenguncianPeriode interface {
	PeriksaPeriodeTerkunci(ctx context.Context, tanggal time.Time) (bool, error)
}

// Service mengatur jadwal shift mingguan.
type Service struct {
	repo    *Repository
	periode PenguncianPeriode
}

// NewService membuat service jadwal.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// AturPeriode memasang PeriodeProvider dari Modul 5 (opsional).
// Tanpa ini, atau selama fungsi Modul 5 belum diimplementasi, service
// memeriksa tabel periode_gaji langsung.
func (s *Service) AturPeriode(p PenguncianPeriode) {
	s.periode = p
}

// Terkunci memeriksa apakah tanggal berada di periode yang sudah ditutup.
func (s *Service) Terkunci(ctx context.Context, tanggal time.Time) (bool, error) {
	if s.periode != nil {
		kunci, err := s.periode.PeriksaPeriodeTerkunci(ctx, tanggal)
		if err == nil {
			return kunci, nil
		}
		if !errors.Is(err, domain.ErrBelumDiimplementasi) {
			return false, err
		}
	}
	return s.repo.TanggalTerkunci(ctx, tanggal)
}

// SeninPekan mengembalikan Senin (pukul 00.00) pada pekan yang memuat t.
// Hari Minggu dianggap penutup pekan sebelumnya.
func SeninPekan(t time.Time) time.Time {
	selisih := (int(t.Weekday()) + 6) % 7
	return time.Date(t.Year(), t.Month(), t.Day()-selisih, 0, 0, 0, 0, t.Location())
}

// WaktuShift mengubah jam shift ("07:00" atau "07:00:00") menjadi datetime
// pada tanggal tertentu. Dipakai Modul 3 dan 4 untuk menghitung terlambat
// dan lembur. Shift tidak melewati tengah malam (jam selesai > jam mulai).
func WaktuShift(tanggal time.Time, jam string) (time.Time, error) {
	t, err := time.Parse("15:04", jam)
	if err != nil {
		t, err = time.Parse("15:04:05", jam)
		if err != nil {
			return time.Time{}, fmt.Errorf("jam shift %q tidak valid: %w", jam, err)
		}
	}
	return time.Date(tanggal.Year(), tanggal.Month(), tanggal.Day(), t.Hour(), t.Minute(), 0, 0, tanggal.Location()), nil
}

// DaftarShift mengambil daftar shift.
//
// Masukan: context.
// Hasil: daftar shift.
// Aturan: 3.
// Tabel: shift.
// Memanggil: Repository.DaftarShift.
func (s *Service) DaftarShift(ctx context.Context) ([]domain.Shift, error) {
	return s.repo.DaftarShift(ctx)
}

// PekerjaPekan mengambil pekerja aktif untuk baris grid pekan yang dimulai pada mulai.
func (s *Service) PekerjaPekan(ctx context.Context, mulai time.Time) ([]domain.Pekerja, error) {
	return s.repo.PekerjaAktifSampai(ctx, mulai.AddDate(0, 0, 5))
}

// JadwalPekan mengambil semua jadwal Senin sampai Sabtu pada pekan itu.
func (s *Service) JadwalPekan(ctx context.Context, mulai time.Time) ([]SelJadwal, error) {
	return s.repo.JadwalRentang(ctx, mulai, mulai.AddDate(0, 0, 5))
}

// BuatJadwal menambahkan jadwal pekerja.
//
// Masukan: context dan data jadwal.
// Hasil: ID jadwal baru.
// Aturan: 3, 14, 19.
// Tabel: jadwal_shift, pekerja, periode_gaji.
// Memanggil: Repository.SimpanJadwal, PeriodeProvider.PeriksaPeriodeTerkunci.
func (s *Service) BuatJadwal(ctx context.Context, jadwal domain.JadwalShift) (int64, error) {
	if jadwal.PekerjaID <= 0 {
		return 0, ErrPekerjaTidakAktif
	}
	if jadwal.ShiftID <= 0 {
		return 0, ErrShiftTidakValid
	}
	if jadwal.Tanggal.IsZero() {
		return 0, ErrTanggalKosong
	}
	if jadwal.Tanggal.Weekday() == time.Sunday {
		return 0, ErrHariMinggu
	}

	aktif, err := s.repo.PekerjaAktif(ctx, jadwal.PekerjaID)
	if err != nil {
		return 0, err
	}
	if !aktif {
		return 0, ErrPekerjaTidakAktif
	}

	ada, err := s.repo.ShiftAda(ctx, jadwal.ShiftID)
	if err != nil {
		return 0, err
	}
	if !ada {
		return 0, ErrShiftTidakValid
	}

	kunci, err := s.Terkunci(ctx, jadwal.Tanggal)
	if err != nil {
		return 0, err
	}
	if kunci {
		return 0, ErrPeriodeTertutup
	}

	_, sudah, err := s.repo.CariJadwalID(ctx, jadwal.PekerjaID, jadwal.Tanggal)
	if err != nil {
		return 0, err
	}
	if sudah {
		return 0, ErrJadwalGanda
	}

	if err := s.repo.SimpanJadwal(ctx, &jadwal); err != nil {
		return 0, err
	}
	return jadwal.ID, nil
}

// periksaBisaDiubah memastikan jadwal tidak terkunci oleh absensi atau periode.
func (s *Service) periksaBisaDiubah(ctx context.Context, id int64) (SelJadwal, error) {
	sel, err := s.repo.JadwalByID(ctx, id)
	if err != nil {
		return SelJadwal{}, err
	}
	if sel.PunyaAbsensi {
		return SelJadwal{}, ErrSudahAdaAbsensi
	}
	kunci, err := s.Terkunci(ctx, sel.Tanggal)
	if err != nil {
		return SelJadwal{}, err
	}
	if kunci {
		return SelJadwal{}, ErrPeriodeTertutup
	}
	return sel, nil
}

// UbahJadwal mengganti shift pada jadwal. Hanya ShiftID yang berubah.
//
// Masukan: context dan data jadwal (ID dan ShiftID terisi).
// Hasil: error.
// Aturan: 14, 19.
// Tabel: jadwal_shift, absensi.
// Memanggil: Repository.UbahJadwal.
func (s *Service) UbahJadwal(ctx context.Context, jadwal domain.JadwalShift) error {
	if jadwal.ShiftID <= 0 {
		return ErrShiftTidakValid
	}
	if _, err := s.periksaBisaDiubah(ctx, jadwal.ID); err != nil {
		return err
	}
	ada, err := s.repo.ShiftAda(ctx, jadwal.ShiftID)
	if err != nil {
		return err
	}
	if !ada {
		return ErrShiftTidakValid
	}
	return s.repo.UbahJadwal(ctx, jadwal)
}

// BatalkanJadwal membatalkan jadwal yang belum punya absensi.
//
// Masukan: context dan ID jadwal.
// Hasil: error.
// Aturan: 14, 19.
// Tabel: jadwal_shift, absensi.
// Memanggil: Repository.BatalkanJadwal.
func (s *Service) BatalkanJadwal(ctx context.Context, id int64) error {
	if _, err := s.periksaBisaDiubah(ctx, id); err != nil {
		return err
	}
	return s.repo.BatalkanJadwal(ctx, id)
}

func kunciSel(pekerjaID int64, tanggal time.Time) string {
	return fmt.Sprintf("%d_%s", pekerjaID, tanggal.Format(formatTanggalSQL))
}

// SalinJadwalMingguLalu menyalin jadwal pekan sebelumnya ke pekan tujuan.
//
// Masukan: context dan tanggal pekan tujuan (dinormalkan ke Senin).
// Hasil: jumlah disalin, jumlah dilewati, error.
// Aturan: 3, 19.
// Tabel: jadwal_shift, pekerja, periode_gaji.
// Memanggil: Repository.JadwalRentang, Repository.SimpanJadwal.
//
// Sel dilewati bila sudah terisi, tanggalnya di periode yang ditutup,
// atau pekerjanya sudah tidak aktif. Sel yang terisi tidak ditimpa.
func (s *Service) SalinJadwalMingguLalu(ctx context.Context, mingguMulai time.Time) (int, int, error) {
	tujuan := SeninPekan(mingguMulai)
	sumber := tujuan.AddDate(0, 0, -7)

	asal, err := s.repo.JadwalRentang(ctx, sumber, sumber.AddDate(0, 0, 5))
	if err != nil {
		return 0, 0, err
	}
	sudahAda, err := s.repo.JadwalRentang(ctx, tujuan, tujuan.AddDate(0, 0, 5))
	if err != nil {
		return 0, 0, err
	}
	terisi := make(map[string]bool, len(sudahAda))
	for _, c := range sudahAda {
		terisi[kunciSel(c.PekerjaID, c.Tanggal)] = true
	}

	aktifList, err := s.repo.PekerjaAktifSampai(ctx, tujuan.AddDate(0, 0, 5))
	if err != nil {
		return 0, 0, err
	}
	aktif := make(map[int64]bool, len(aktifList))
	for _, p := range aktifList {
		aktif[p.ID] = true
	}

	var terkunci [6]bool
	for i := range terkunci {
		terkunci[i], err = s.Terkunci(ctx, tujuan.AddDate(0, 0, i))
		if err != nil {
			return 0, 0, err
		}
	}

	disalin, dilewati := 0, 0
	for _, c := range asal {
		idx := (int(c.Tanggal.Weekday()) + 6) % 7
		if idx > 5 { // Minggu tidak pernah dijadwalkan
			continue
		}
		tanggal := tujuan.AddDate(0, 0, idx)
		if !aktif[c.PekerjaID] || terisi[kunciSel(c.PekerjaID, tanggal)] || terkunci[idx] {
			dilewati++
			continue
		}
		j := domain.JadwalShift{PekerjaID: c.PekerjaID, ShiftID: c.ShiftID, Tanggal: tanggal}
		if err := s.repo.SimpanJadwal(ctx, &j); err != nil {
			return disalin, dilewati, err
		}
		disalin++
	}
	return disalin, dilewati, nil
}

// AmbilJadwal mengambil jadwal pekerja pada tanggal.
//
// Masukan: context, ID pekerja, tanggal.
// Hasil: JadwalShift beserta shift (jam "07:00"); ErrJadwalTidakAda bila
// pekerja tidak dijadwalkan. Gunakan WaktuShift untuk mendapat datetime.
// Aturan: 3.
// Tabel: jadwal_shift, shift.
// Memanggil: Repository.AmbilJadwal.
func (s *Service) AmbilJadwal(ctx context.Context, pekerjaID int64, tanggal time.Time) (domain.JadwalShift, error) {
	return s.repo.AmbilJadwal(ctx, pekerjaID, tanggal)
}

// Pastikan Service memenuhi kontrak lintas modul.
var _ kontrak.JadwalProvider = (*Service)(nil)