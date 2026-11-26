// Pemilik: Fadli | Modul 6 | Aturan: 13, 17
package laporan

import (
	"context"
	"fmt"
	"log"

	"penggajian-konveksi/internal/domain"
	"penggajian-konveksi/internal/kontrak"
)

// ErrPeriodeBelumDihitung dikembalikan bila periode masih berstatus Terbuka.
var ErrPeriodeBelumDihitung = fmt.Errorf("laporan hanya tersedia untuk periode Dihitung atau Ditutup")

// repoAkses mendefinisikan operasi repository yang dibutuhkan Service.
// Interface ini memungkinkan pengujian unit tanpa database.
type repoAkses interface {
	AmbilSlip(ctx context.Context, periodeID, pekerjaID int64) (domain.SlipGaji, error)
	DaftarSlip(ctx context.Context, periodeID int64) ([]domain.SlipGaji, error)
	Rekap(ctx context.Context, periodeID int64) (domain.RekapGaji, error)
	AmbilStatusPeriode(ctx context.Context, periodeID int64) (domain.StatusPeriode, error)
	DaftarRincianSlip(ctx context.Context, slipID int64) ([]domain.SlipGajiDetail, error)
	DaftarPeriodeTerbaru(ctx context.Context, limit int) ([]domain.PeriodeGaji, error)
	DaftarPeriodeSiapLaporan(ctx context.Context, limit int) ([]domain.PeriodeGaji, error)
	HitungAbsensiBelumLengkap(ctx context.Context, periodeID int64) (int, error)
	TotalGajiBersihPeriode(ctx context.Context, periodeID int64) (domain.Rupiah, error)
	TotalSisaKasbon(ctx context.Context) (domain.Rupiah, error)
	SisaKasbonPekerja(ctx context.Context, pekerjaID int64) (domain.Rupiah, error)
}

// Pastikan Repository memenuhi repoAkses.
var _ repoAkses = (*Repository)(nil)

// Service menyediakan slip dan rekap penggajian.
type Service struct {
	repo repoAkses
}

// NewService membuat service laporan.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func statusBolehDilaporkan(status domain.StatusPeriode) bool {
	return status == domain.PeriodeDihitung || status == domain.PeriodeDitutup
}

// periksaStatusPeriode memastikan periode berstatus Dihitung atau Ditutup.
func (s *Service) periksaStatusPeriode(ctx context.Context, periodeID int64) error {
	status, err := s.repo.AmbilStatusPeriode(ctx, periodeID)
	if err != nil {
		return err
	}
	if !statusBolehDilaporkan(status) {
		return ErrPeriodeBelumDihitung
	}
	return nil
}

// AmbilSlipGaji mengambil satu slip dari data tersimpan.
// Hanya periode Dihitung atau Ditutup yang boleh ditampilkan (fail-closed bila Periode nil).
func (s *Service) AmbilSlipGaji(ctx context.Context, periodeID, pekerjaID int64) (domain.SlipGaji, error) {
	slip, err := s.repo.AmbilSlip(ctx, periodeID, pekerjaID)
	if err != nil {
		return domain.SlipGaji{}, err
	}
	if slip.Periode == nil || !statusBolehDilaporkan(slip.Periode.Status) {
		return domain.SlipGaji{}, ErrPeriodeBelumDihitung
	}
	return slip, nil
}

// DaftarSlipPeriode mengambil semua slip dalam satu periode.
// Periode Terbuka ditolak; hanya Dihitung atau Ditutup yang diizinkan.
func (s *Service) DaftarSlipPeriode(ctx context.Context, periodeID int64) ([]domain.SlipGaji, error) {
	if err := s.periksaStatusPeriode(ctx, periodeID); err != nil {
		return nil, err
	}
	return s.repo.DaftarSlip(ctx, periodeID)
}

// DaftarRincianSlip mengambil rincian upah harian dari repository.
func (s *Service) DaftarRincianSlip(ctx context.Context, slipID int64) ([]domain.SlipGajiDetail, error) {
	return s.repo.DaftarRincianSlip(ctx, slipID)
}

// SisaKasbonPekerja mengambil sisa kasbon berjalan milik pekerja untuk tampilan slip.
func (s *Service) SisaKasbonPekerja(ctx context.Context, pekerjaID int64) (domain.Rupiah, error) {
	return s.repo.SisaKasbonPekerja(ctx, pekerjaID)
}

// RekapGajiPerPeriode mengambil rekap penggajian satu periode.
// Periode Terbuka ditolak; hanya Dihitung atau Ditutup yang diizinkan.
// Potongan keterlambatan tidak dipotong ulang di rekap karena sudah terkandung dalam gaji_bruto.
func (s *Service) RekapGajiPerPeriode(ctx context.Context, periodeID int64) (domain.RekapGaji, error) {
	if err := s.periksaStatusPeriode(ctx, periodeID); err != nil {
		return domain.RekapGaji{}, err
	}
	return s.repo.Rekap(ctx, periodeID)
}

// RekapDanDaftarSlip mengambil daftar slip beserta total rekap yang dihitung konsisten
// dari snapshot slip yang sama (menghindari race & query ganda AmbilStatusPeriode).
func (s *Service) RekapDanDaftarSlip(ctx context.Context, periodeID int64) (domain.RekapGaji, []domain.SlipGaji, error) {
	slips, err := s.DaftarSlipPeriode(ctx, periodeID)
	if err != nil {
		return domain.RekapGaji{}, nil, err
	}
	rekap := domain.RekapGaji{
		PeriodeID:  periodeID,
		JumlahSlip: len(slips),
	}
	for _, slip := range slips {
		rekap.TotalBruto += slip.GajiBruto
		rekap.TotalLembur += slip.TotalLembur
		rekap.TotalPotonganKasbon += slip.PotonganKasbon
		rekap.TotalBersih += slip.GajiBersih
	}
	return rekap, slips, nil
}

// RingkasanDashboard berisi data ringkasan untuk halaman dashboard.
type RingkasanDashboard struct {
	PeriodeBerjalan        *domain.PeriodeGaji
	AbsensiBelumLengkap    int
	GajiBersihTerakhir     domain.Rupiah
	TotalSisaKasbon        domain.Rupiah
	DaftarPeriode          []domain.PeriodeGaji
	PeriodeTerakhirSelesai *domain.PeriodeGaji
}

// AmbilRingkasanDashboard mengumpulkan data ringkasan dari database.
// Kesalahan pada setiap bagian dicatat ke log dan ditangani secara graceful.
func (s *Service) AmbilRingkasanDashboard(ctx context.Context) RingkasanDashboard {
	var ringkasan RingkasanDashboard

	periodes, err := s.repo.DaftarPeriodeTerbaru(ctx, 10)
	if err != nil {
		log.Printf("dashboard: gagal mengambil periode terbaru: %v", err)
	} else {
		ringkasan.DaftarPeriode = periodes
		var terbuka []domain.PeriodeGaji
		for i := range periodes {
			if periodes[i].Status == domain.PeriodeTerbuka {
				if ringkasan.PeriodeBerjalan == nil {
					ringkasan.PeriodeBerjalan = &periodes[i]
				}
				terbuka = append(terbuka, periodes[i])
			}
			if periodes[i].Status == domain.PeriodeDitutup && ringkasan.PeriodeTerakhirSelesai == nil {
				ringkasan.PeriodeTerakhirSelesai = &periodes[i]
			}
		}
		// Jumlahkan absensi belum lengkap pada semua periode terbuka (mingguan + bulanan yang berjalan paralel).
		for _, p := range terbuka {
			n, err := s.repo.HitungAbsensiBelumLengkap(ctx, p.ID)
			if err != nil {
				log.Printf("dashboard: gagal hitung absensi belum lengkap periode=%d: %v", p.ID, err)
				continue
			}
			ringkasan.AbsensiBelumLengkap += n
		}
	}

	if ringkasan.PeriodeTerakhirSelesai != nil {
		total, err := s.repo.TotalGajiBersihPeriode(ctx, ringkasan.PeriodeTerakhirSelesai.ID)
		if err != nil {
			log.Printf("dashboard: gagal ambil total gaji bersih periode=%d: %v", ringkasan.PeriodeTerakhirSelesai.ID, err)
		} else {
			ringkasan.GajiBersihTerakhir = total
		}
	}

	totalKasbon, err := s.repo.TotalSisaKasbon(ctx)
	if err != nil {
		log.Printf("dashboard: gagal ambil total sisa kasbon: %v", err)
	} else {
		ringkasan.TotalSisaKasbon = totalKasbon
	}

	return ringkasan
}

// DaftarPeriodeTersedia mengambil daftar periode dengan status Dihitung atau Ditutup
// langsung dari query SQL untuk pemilih periode di halaman rekap.
func (s *Service) DaftarPeriodeTersedia(ctx context.Context) ([]domain.PeriodeGaji, error) {
	return s.repo.DaftarPeriodeSiapLaporan(ctx, 100)
}

// Memastikan Service memenuhi kontrak SlipReader.
var _ kontrak.SlipReader = (*Service)(nil)
