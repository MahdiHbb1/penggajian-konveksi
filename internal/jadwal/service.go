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

// PenguncianPeriodeJenis memeriksa kunci periode sesuai jenis pekerja (HARIAN->MINGGUAN, BULANAN->BULANAN).
type PenguncianPeriodeJenis interface {
	PeriksaPeriodeTerkunciJenis(ctx context.Context, jenis domain.JenisPekerja, tanggal time.Time) (bool, error)
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

// AturPeriode memasang PeriodeProvider dari Modul 5.
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

// TerkunciJenis memeriksa apakah tanggal berada di periode DITUTUP yang sesuai jenis pekerja.
func (s *Service) TerkunciJenis(ctx context.Context, jenis domain.JenisPekerja, tanggal time.Time) (bool, error) {
	if pj, ok := s.periode.(PenguncianPeriodeJenis); ok {
		kunci, err := pj.PeriksaPeriodeTerkunciJenis(ctx, jenis, tanggal)
		if err == nil {
			return kunci, nil
		}
		if !errors.Is(err, domain.ErrBelumDiimplementasi) {
			return false, err
		}
	}
	return s.repo.TanggalTerkunciJenis(ctx, jenis, tanggal)
}

// SeninPekan mengembalikan Senin (pukul 00.00) pada pekan yang memuat t.
// Hari Minggu dianggap penutup pekan sebelumnya.
func SeninPekan(t time.Time) time.Time {
	selisih := (int(t.Weekday()) + 6) % 7
	return time.Date(t.Year(), t.Month(), t.Day()-selisih, 0, 0, 0, 0, t.Location())
}

// WaktuShift mengubah jam shift ("07:00" atau "07:00:00") menjadi datetime pada tanggal tertentu.
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

	aktif, jenis, err := s.repo.PekerjaAktifDanJenis(ctx, jadwal.PekerjaID)
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

	kunci, err := s.TerkunciJenis(ctx, jenis, jadwal.Tanggal)
	if err != nil {
		return 0, err
	}
	if kunci {
		return 0, ErrPeriodeTertutup
	}

	if err := s.repo.SimpanJadwal(ctx, &jadwal); err != nil {
		return 0, err
	}
	return jadwal.ID, nil
}

// BuatJadwalTervalidasi menyimpan jadwal ketika status pekerja aktif, shift, dan kunci periode
// sudah divalidasi secara batch di pemanggil (menghindari N+1 saat simpan grid pekan).
func (s *Service) BuatJadwalTervalidasi(ctx context.Context, jadwal domain.JadwalShift) (int64, error) {
	if jadwal.Tanggal.Weekday() == time.Sunday {
		return 0, ErrHariMinggu
	}
	if err := s.repo.SimpanJadwal(ctx, &jadwal); err != nil {
		return 0, err
	}
	return jadwal.ID, nil
}

// UbahJadwal mengganti shift pada jadwal secara atomik (terlindungi lock transaksi).
func (s *Service) UbahJadwal(ctx context.Context, jadwal domain.JadwalShift) error {
	if jadwal.ShiftID <= 0 {
		return ErrShiftTidakValid
	}
	return s.repo.UbahJadwalAtomik(ctx, jadwal.ID, jadwal.ShiftID)
}

// BatalkanJadwal membatalkan jadwal yang belum punya absensi secara atomik.
func (s *Service) BatalkanJadwal(ctx context.Context, id int64) error {
	return s.repo.BatalkanJadwalAtomik(ctx, id)
}

func kunciSel(pekerjaID int64, tanggal time.Time) string {
	return fmt.Sprintf("%d_%s", pekerjaID, tanggal.Format(formatTanggalSQL))
}

// SalinJadwalMingguLalu menyalin jadwal pekan sebelumnya ke pekan tujuan secara transaksional.
// Sel dilewati bila sudah terisi, tanggalnya terkunci sesuai jenis pekerja, atau pekerja nonaktif.
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
	jenisAktif := make(map[int64]domain.JenisPekerja, len(aktifList))
	for _, p := range aktifList {
		jenisAktif[p.ID] = p.Jenis
	}

	var kunciHarian, kunciBulanan [6]bool
	for i := 0; i < 6; i++ {
		tgl := tujuan.AddDate(0, 0, i)
		kunciHarian[i], err = s.TerkunciJenis(ctx, domain.PekerjaHarian, tgl)
		if err != nil {
			return 0, 0, err
		}
		kunciBulanan[i], err = s.TerkunciJenis(ctx, domain.PekerjaBulanan, tgl)
		if err != nil {
			return 0, 0, err
		}
	}

	dilewatiAwal := 0
	var akanDisalin []domain.JadwalShift
	for _, c := range asal {
		idx := (int(c.Tanggal.Weekday()) + 6) % 7
		if idx > 5 {
			continue
		}
		tanggal := tujuan.AddDate(0, 0, idx)
		jenis, ada := jenisAktif[c.PekerjaID]
		terkunci := kunciHarian[idx]
		if jenis == domain.PekerjaBulanan {
			terkunci = kunciBulanan[idx]
		}
		if !ada || terisi[kunciSel(c.PekerjaID, tanggal)] || terkunci {
			dilewatiAwal++
			continue
		}
		akanDisalin = append(akanDisalin, domain.JadwalShift{
			PekerjaID: c.PekerjaID,
			ShiftID:   c.ShiftID,
			Tanggal:   tanggal,
		})
	}

	disalin, dilewatiKonflik, err := s.repo.SalinJadwalBatch(ctx, akanDisalin)
	if err != nil {
		return 0, 0, err
	}
	return disalin, dilewatiAwal + dilewatiKonflik, nil
}

// AmbilJadwal mengambil jadwal pekerja pada tanggal.
func (s *Service) AmbilJadwal(ctx context.Context, pekerjaID int64, tanggal time.Time) (domain.JadwalShift, error) {
	return s.repo.AmbilJadwal(ctx, pekerjaID, tanggal)
}

// Pastikan Service memenuhi kontrak lintas modul.
var _ kontrak.JadwalProvider = (*Service)(nil)
