// Pemilik: Anya | Modul 5 | Aturan: 1, 12, 13, 14, 15, 18
package periode

import (
	"context"
	"errors"
	"penggajian-konveksi/internal/domain"
	"strings"
	"time"
)

type Service struct{ repo *Repository }

func NewService(r *Repository) *Service { return &Service{repo: r} }
func (s *Service) BuatPeriode(ctx context.Context, tipe domain.TipePeriode, mulai time.Time) (int64, error) {
	mulai = time.Date(mulai.Year(), mulai.Month(), mulai.Day(), 0, 0, 0, 0, mulai.Location())
	var selesai time.Time
	switch tipe {
	case domain.PeriodeMingguan:
		if mulai.Weekday() != time.Monday {
			return 0, errors.New("periode mingguan wajib dimulai hari Senin")
		}
		selesai = mulai.AddDate(0, 0, 5)
	case domain.PeriodeBulanan:
		if mulai.Day() != 1 {
			return 0, errors.New("periode bulanan wajib dimulai tanggal 1")
		}
		selesai = mulai.AddDate(0, 1, 0).AddDate(0, 0, -1)
	default:
		return 0, errors.New("tipe periode tidak valid")
	}
	p := domain.PeriodeGaji{Tipe: tipe, TanggalMulai: mulai, TanggalSelesai: selesai, Status: domain.PeriodeTerbuka}
	ada, e := s.repo.AdaTumpangTindih(ctx, p)
	if e != nil {
		return 0, e
	}
	if ada {
		return 0, errors.New("periode bertumpang tindih dengan periode yang sudah ada")
	}
	if e = s.repo.SimpanPeriode(ctx, &p); e != nil {
		return 0, e
	}
	return p.ID, nil
}
func (s *Service) DaftarPeriode(c context.Context) ([]domain.PeriodeGaji, error) {
	return s.repo.DaftarPeriode(c)
}
func (s *Service) AmbilPeriode(c context.Context, id int64) (domain.PeriodeGaji, error) {
	return s.repo.AmbilPeriode(c, id)
}
func (s *Service) PeriksaPeriodeTerkunci(c context.Context, t time.Time) (bool, error) {
	return s.repo.PeriodeTerkunci(c, t)
}
func (s *Service) PeriodeSebelumnyaDitutup(c context.Context, p domain.PeriodeGaji) (bool, error) {
	prev, ada, e := s.repo.PeriodeSebelumnya(c, p)
	if e != nil {
		return false, e
	}
	return !ada || prev.Status == domain.PeriodeDitutup, nil
}
func (s *Service) UbahStatusPeriode(c context.Context, id int64, st domain.StatusPeriode) error {
	if st == domain.PeriodeDitutup {
		return errors.New("gunakan TutupPeriode untuk menutup periode")
	}
	if st != domain.PeriodeTerbuka && st != domain.PeriodeDihitung {
		return errors.New("status periode tidak valid")
	}
	return s.repo.UbahStatus(c, id, st)
}
func (s *Service) TutupPeriode(c context.Context, id int64) error {
	if s.repo == nil || s.repo.db == nil {
		return errors.New("database periode tidak tersedia")
	}
	p, e := s.repo.AmbilPeriode(c, id)
	if e != nil {
		return e
	}
	if p.Status != domain.PeriodeDihitung {
		return errors.New("periode hanya dapat ditutup setelah dihitung")
	}
	return s.repo.Tutup(c, id)
}
func (s *Service) RingkasanTutupPeriode(c context.Context, id int64) (RingkasanTutup, error) {
	return s.repo.Ringkasan(c, id)
}
func (s *Service) CatatKoreksi(c context.Context, k domain.KoreksiGaji) (int64, error) {
	if s.repo == nil || s.repo.db == nil {
		return 0, errors.New("database koreksi tidak tersedia")
	}
	p, e := s.repo.AmbilPeriode(c, k.PeriodeAsalID)
	if e != nil {
		return 0, e
	}
	if p.Status != domain.PeriodeDitutup {
		return 0, errors.New("koreksi hanya dapat dicatat untuk periode yang sudah ditutup")
	}
	if k.PekerjaID <= 0 || k.Jumlah == 0 || strings.TrimSpace(k.Keterangan) == "" {
		return 0, errors.New("pekerja, nominal nonzero, dan alasan wajib diisi")
	}
	e = s.repo.SimpanKoreksi(c, &k)
	return k.ID, e
}
func (s *Service) DaftarKoreksi(c context.Context, p *int64) ([]domain.KoreksiGaji, error) {
	return s.repo.DaftarKoreksi(c, p)
}
func (s *Service) DaftarPekerjaKoreksi(c context.Context) ([]domain.Pekerja, error) {
	return s.repo.DaftarPekerjaKoreksi(c)
}
func (s *Service) DaftarKoreksiMenunggu(c context.Context, p int64) ([]domain.KoreksiGaji, error) {
	return s.repo.DaftarKoreksiMenunggu(c, p)
}
func (s *Service) TandaiKoreksiDiterapkan(c context.Context, id int64) error {
	if s.repo == nil || s.repo.db == nil {
		return errors.New("database koreksi tidak tersedia")
	}
	return s.repo.TandaiKoreksiDiterapkan(c, id)
}

var _ interface {
	AmbilPeriode(context.Context, int64) (domain.PeriodeGaji, error)
	PeriodeSebelumnyaDitutup(context.Context, domain.PeriodeGaji) (bool, error)
	UbahStatusPeriode(context.Context, int64, domain.StatusPeriode) error
	PeriksaPeriodeTerkunci(context.Context, time.Time) (bool, error)
	DaftarKoreksiMenunggu(context.Context, int64) ([]domain.KoreksiGaji, error)
	TandaiKoreksiDiterapkan(context.Context, int64) error
} = (*Service)(nil)
