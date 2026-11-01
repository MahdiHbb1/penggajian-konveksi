// Pemilik: Mahdi | Modul 1 | Aturan: 1 sampai 19 (kontrak lintas modul)
package kontrak

import (
	"context"
	"time"

	"penggajian-konveksi/internal/domain"
)

// PekerjaProvider disediakan Modul 1 untuk modul lain.
// Implementasi: *pekerja.Service. Pemakai: penggajian dan jadwal.
type PekerjaProvider interface {
	CariUpahBerlaku(ctx context.Context, pekerjaID int64, tanggal time.Time) (domain.Rupiah, error)
	DaftarPekerjaAktif(ctx context.Context, jenis domain.JenisPekerja, mulai, selesai time.Time) ([]domain.Pekerja, error)
}

// JadwalProvider disediakan Modul 2 untuk modul lain.
// Implementasi: *jadwal.Service. Pemakai: absensi dan penggajian.
type JadwalProvider interface {
	AmbilJadwal(ctx context.Context, pekerjaID int64, tanggal time.Time) (domain.JadwalShift, error)
}

// KasbonProvider disediakan Modul 2 untuk modul lain.
// Implementasi: *kasbon.Service. Pemakai: penggajian dan periode.
type KasbonProvider interface {
	KasbonDiprosesSampai(ctx context.Context, pekerjaID int64, tanggal time.Time) ([]domain.Kasbon, error)
	SisaKasbon(ctx context.Context, pekerjaID int64) (domain.Rupiah, error)
}

// AbsensiProvider disediakan Modul 3 untuk modul lain.
// Implementasi: *absensi.Service. Pemakai: penggajian.
type AbsensiProvider interface {
	AbsensiDalamPeriode(ctx context.Context, periodeID int64) ([]domain.AbsensiLengkap, error)
	AdaAbsensiBelumLengkap(ctx context.Context, periodeID int64) (bool, int, error)
	HitungMenitTerlambat(jamMasuk time.Time, jamMulaiShift time.Time) int
	HitungMenitLembur(jamKeluar time.Time, jamSelesaiShift time.Time) int
}

// PeriodeProvider disediakan Modul 5 untuk modul lain.
// Implementasi: *periode.Service. Pemakai: penggajian, absensi, jadwal, kasbon.
type PeriodeProvider interface {
	AmbilPeriode(ctx context.Context, periodeID int64) (domain.PeriodeGaji, error)
	PeriodeSebelumnyaDitutup(ctx context.Context, periode domain.PeriodeGaji) (bool, error)
	UbahStatusPeriode(ctx context.Context, periodeID int64, status domain.StatusPeriode) error
	PeriksaPeriodeTerkunci(ctx context.Context, tanggal time.Time) (bool, error)
}

// KoreksiProvider disediakan Modul 5 untuk modul lain.
// Implementasi: *periode.Service. Pemakai: penggajian.
type KoreksiProvider interface {
	DaftarKoreksiMenunggu(ctx context.Context, pekerjaID int64) ([]domain.KoreksiGaji, error)
	TandaiKoreksiDiterapkan(ctx context.Context, periodeID int64) error
}

// SlipReader disediakan Modul 6 untuk modul lain.
// Implementasi: *laporan.Service. Pemakai: halaman slip dan rekap.
type SlipReader interface {
	AmbilSlipGaji(ctx context.Context, periodeID, pekerjaID int64) (domain.SlipGaji, error)
	DaftarSlipPeriode(ctx context.Context, periodeID int64) ([]domain.SlipGaji, error)
	RekapGajiPerPeriode(ctx context.Context, periodeID int64) (domain.RekapGaji, error)
}
