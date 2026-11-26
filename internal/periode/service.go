// Pemilik: Anya | Modul 5 | Aturan: 1, 12, 13, 14, 15, 18
package periode

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"penggajian-konveksi/internal/domain"
	"penggajian-konveksi/internal/kontrak"
)

// Service mengatur pembuatan periode, penutupan, dan koreksi gaji.
type Service struct{ repo *Repository }

// NewService membuat service periode.
func NewService(r *Repository) *Service { return &Service{repo: r} }

func (s *Service) siap() error {
	if s == nil || s.repo == nil || s.repo.db == nil {
		return ErrDatabasePeriodeTidakTersedia
	}
	return nil
}

func normalisasiTanggal(t time.Time) time.Time {
	tahun, bulan, hari := t.Date()
	return time.Date(tahun, bulan, hari, 0, 0, 0, 0, t.Location())
}

// hitungRentang menurunkan tanggal selesai dari tanggal mulai.
// Mingguan selalu Senin sampai Sabtu, bulanan selalu tanggal 1 sampai akhir bulan.
func hitungRentang(tipe domain.TipePeriode, mulai time.Time) (time.Time, error) {
	switch tipe {
	case domain.PeriodeMingguan:
		if mulai.Weekday() != time.Monday {
			return time.Time{}, errors.New("periode mingguan wajib dimulai hari Senin")
		}
		selesai := mulai.AddDate(0, 0, 5)
		if selesai.Weekday() != time.Saturday {
			return time.Time{}, errors.New("periode mingguan wajib selesai hari Sabtu")
		}
		return selesai, nil
	case domain.PeriodeBulanan:
		if mulai.Day() != 1 {
			return time.Time{}, errors.New("periode bulanan wajib dimulai tanggal 1")
		}
		return mulai.AddDate(0, 1, -1), nil
	default:
		return time.Time{}, errors.New("tipe periode tidak valid")
	}
}

// BuatPeriode membuat periode mingguan atau bulanan yang tidak tumpang tindih.
func (s *Service) BuatPeriode(ctx context.Context, tipe domain.TipePeriode, mulai time.Time) (int64, error) {
	if mulai.IsZero() {
		return 0, errors.New("tanggal mulai periode wajib diisi")
	}
	mulai = normalisasiTanggal(mulai)
	selesai, e := hitungRentang(tipe, mulai)
	if e != nil {
		return 0, e
	}
	if e = s.siap(); e != nil {
		return 0, e
	}
	p := domain.PeriodeGaji{Tipe: tipe, TanggalMulai: mulai, TanggalSelesai: selesai, Status: domain.PeriodeTerbuka}
	ada, e := s.repo.AdaTumpangTindih(ctx, p)
	if e != nil {
		return 0, e
	}
	if ada {
		return 0, ErrPeriodeTumpangTindih
	}
	if e = s.repo.SimpanPeriode(ctx, &p); e != nil {
		return 0, e
	}
	return p.ID, nil
}

// DaftarPeriode mengambil seluruh periode.
func (s *Service) DaftarPeriode(c context.Context) ([]domain.PeriodeGaji, error) {
	if e := s.siap(); e != nil {
		return nil, e
	}
	return s.repo.DaftarPeriode(c)
}

// AmbilPeriode mengambil satu periode.
func (s *Service) AmbilPeriode(c context.Context, id int64) (domain.PeriodeGaji, error) {
	if e := s.siap(); e != nil {
		return domain.PeriodeGaji{}, e
	}
	return s.repo.AmbilPeriode(c, id)
}

// PeriksaPeriodeTerkunci menunjukkan apakah tanggal berada di periode yang sudah ditutup.
func (s *Service) PeriksaPeriodeTerkunci(c context.Context, t time.Time) (bool, error) {
	if e := s.siap(); e != nil {
		return false, e
	}
	return s.repo.PeriodeTerkunci(c, t)
}

// PeriksaPeriodeTerkunciJenis menunjukkan apakah tanggal berada di periode DITUTUP
// yang sesuai jenis pekerja (HARIAN -> MINGGUAN, BULANAN -> BULANAN).
func (s *Service) PeriksaPeriodeTerkunciJenis(c context.Context, jenis domain.JenisPekerja, t time.Time) (bool, error) {
	if e := s.siap(); e != nil {
		return false, e
	}
	return s.repo.PeriodeTerkunciJenis(c, jenis, t)
}

// PeriodeSebelumnyaDitutup bernilai benar bila periode sebelumnya sudah ditutup.
// Periode pertama dengan tipe tertentu selalu lolos.
func (s *Service) PeriodeSebelumnyaDitutup(c context.Context, p domain.PeriodeGaji) (bool, error) {
	if e := s.siap(); e != nil {
		return false, e
	}
	prev, ada, e := s.repo.PeriodeSebelumnya(c, p)
	if e != nil {
		return false, e
	}
	return !ada || prev.Status == domain.PeriodeDitutup, nil
}

// UbahStatusPeriode mengubah status TERBUKA atau DIHITUNG.
// Status DITUTUP hanya boleh lewat TutupPeriode dan tidak dapat dibatalkan.
func (s *Service) UbahStatusPeriode(c context.Context, id int64, st domain.StatusPeriode) error {
	if st == domain.PeriodeDitutup {
		return errors.New("gunakan TutupPeriode untuk menutup periode")
	}
	if st != domain.PeriodeTerbuka && st != domain.PeriodeDihitung {
		return errors.New("status periode tidak valid")
	}
	if e := s.siap(); e != nil {
		return e
	}
	return s.repo.UbahStatus(c, id, st)
}

// TutupPeriode menutup periode setelah dihitung dalam satu transaksi repository.
func (s *Service) TutupPeriode(c context.Context, id int64) error {
	if e := s.siap(); e != nil {
		return e
	}
	p, e := s.repo.AmbilPeriode(c, id)
	if e != nil {
		return e
	}
	if p.Status == domain.PeriodeDitutup {
		return ErrPeriodeSudahDitutup
	}
	if p.Status != domain.PeriodeDihitung {
		return errors.New("periode hanya dapat ditutup setelah dihitung")
	}
	sebelumnyaDitutup, e := s.PeriodeSebelumnyaDitutup(c, p)
	if e != nil {
		return e
	}
	if !sebelumnyaDitutup {
		return errors.New("periode sebelumnya harus ditutup terlebih dahulu")
	}
	return s.repo.Tutup(c, id)
}

// RingkasanTutupPeriode mengambil isi halaman konfirmasi penutupan.
func (s *Service) RingkasanTutupPeriode(c context.Context, id int64) (RingkasanTutup, error) {
	if e := s.siap(); e != nil {
		return RingkasanTutup{}, e
	}
	return s.repo.Ringkasan(c, id)
}

// CatatKoreksi mencatat koreksi hanya untuk periode yang sudah ditutup.
func (s *Service) CatatKoreksi(c context.Context, k domain.KoreksiGaji) (int64, error) {
	k.Keterangan = strings.TrimSpace(k.Keterangan)
	if k.PekerjaID <= 0 {
		return 0, errors.New("pekerja wajib dipilih")
	}
	if k.Jumlah == 0 {
		return 0, errors.New("nominal koreksi wajib diisi dan tidak boleh nol")
	}
	if k.Keterangan == "" {
		return 0, errors.New("alasan koreksi wajib diisi")
	}
	if utf8.RuneCountInString(k.Keterangan) > 200 {
		return 0, errors.New("alasan koreksi maksimal 200 karakter")
	}
	if e := s.siap(); e != nil {
		return 0, e
	}
	p, e := s.repo.AmbilPeriode(c, k.PeriodeAsalID)
	if e != nil {
		return 0, e
	}
	if p.Status != domain.PeriodeDitutup {
		return 0, errors.New("koreksi hanya dapat dicatat untuk periode yang sudah ditutup")
	}
	ada, e := s.repo.PekerjaAda(c, k.PekerjaID)
	if e != nil {
		return 0, e
	}
	if !ada {
		return 0, errors.New("pekerja tidak ditemukan")
	}
	e = s.repo.SimpanKoreksi(c, &k)
	return k.ID, e
}

// DaftarKoreksi mengambil koreksi, seluruhnya atau milik satu pekerja.
func (s *Service) DaftarKoreksi(c context.Context, p *int64) ([]domain.KoreksiGaji, error) {
	if e := s.siap(); e != nil {
		return nil, e
	}
	return s.repo.DaftarKoreksi(c, p)
}

// DaftarPekerjaKoreksi mengambil pilihan pekerja pada formulir koreksi.
func (s *Service) DaftarPekerjaKoreksi(c context.Context) ([]domain.Pekerja, error) {
	if e := s.siap(); e != nil {
		return nil, e
	}
	return s.repo.DaftarPekerjaKoreksi(c)
}

// DaftarKoreksiMenunggu mengambil koreksi yang belum masuk slip pekerja.
func (s *Service) DaftarKoreksiMenunggu(c context.Context, p int64) ([]domain.KoreksiGaji, error) {
	if e := s.siap(); e != nil {
		return nil, e
	}
	return s.repo.DaftarKoreksiMenunggu(c, p)
}

// TandaiKoreksiDiterapkan menandai koreksi selesai setelah masuk slip.
func (s *Service) TandaiKoreksiDiterapkan(c context.Context, id int64) error {
	if e := s.siap(); e != nil {
		return e
	}
	return s.repo.TandaiKoreksiDiterapkan(c, id)
}

// Pastikan Service memenuhi kontrak lintas modul.
var _ kontrak.PeriodeProvider = (*Service)(nil)
var _ kontrak.KoreksiProvider = (*Service)(nil)
