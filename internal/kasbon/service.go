// Pemilik: Edgar | Modul 2 | Aturan: 12, 14, 16
package kasbon

import (
	"context"
	"errors"
	"time"
	"unicode/utf8"

	"penggajian-konveksi/internal/domain"
	"penggajian-konveksi/internal/kontrak"
)

// Pesan galat sesuai bagian 5.2 dokumen rancangan.
var (
	ErrJumlahTidakValid  = errors.New("Jumlah kasbon harus lebih dari 0.")
	ErrPeriodeTertutup   = errors.New("Tanggal berada pada periode yang sudah ditutup.")
	ErrTanggalKosong     = errors.New("Tanggal kasbon wajib diisi.")
	ErrKeteranganPanjang = errors.New("Keterangan kasbon maksimal 150 karakter.")
	ErrPekerjaTidakAktif = errors.New("Pekerja wajib dipilih dan harus berstatus aktif.")
	ErrKasbonTidakAda    = errors.New("Kasbon tidak ditemukan.")
)

// daftarGalatInput adalah galat yang aman ditampilkan ke pengguna.
var daftarGalatInput = []error{
	ErrJumlahTidakValid, ErrPeriodeTertutup, ErrTanggalKosong,
	ErrKeteranganPanjang, ErrPekerjaTidakAktif,
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

const maksKeterangan = 150

// PenguncianPeriode adalah bagian PeriodeProvider (Modul 5) yang kita pakai.
type PenguncianPeriode interface {
	PeriksaPeriodeTerkunci(ctx context.Context, tanggal time.Time) (bool, error)
}

// Service mengatur kasbon dan sisa potongannya.
type Service struct {
	repo    *Repository
	periode PenguncianPeriode
}

// NewService membuat service kasbon.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// AturPeriode memasang PeriodeProvider dari Modul 5 (opsional).
// Tanpa ini, atau selama fungsi Modul 5 belum diimplementasi, service
// memeriksa tabel periode_gaji langsung.
func (s *Service) AturPeriode(p PenguncianPeriode) {
	s.periode = p
}

func (s *Service) terkunci(ctx context.Context, tanggal time.Time) (bool, error) {
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

// validasiKasbon memeriksa aturan yang tidak butuh database.
func validasiKasbon(k domain.Kasbon) error {
	if k.PekerjaID <= 0 {
		return ErrPekerjaTidakAktif
	}
	if k.Tanggal.IsZero() {
		return ErrTanggalKosong
	}
	if k.Jumlah <= 0 {
		return ErrJumlahTidakValid
	}
	if utf8.RuneCountInString(k.Keterangan) > maksKeterangan {
		return ErrKeteranganPanjang
	}
	return nil
}

// CatatKasbon mencatat kasbon baru.
//
// Masukan: context dan data kasbon.
// Hasil: ID kasbon baru.
// Aturan: 12, 14.
// Tabel: kasbon, pekerja, periode_gaji.
// Memanggil: Repository.SimpanKasbon, PeriodeProvider.PeriksaPeriodeTerkunci.
func (s *Service) CatatKasbon(ctx context.Context, kasbon domain.Kasbon) (int64, error) {
	if err := validasiKasbon(kasbon); err != nil {
		return 0, err
	}

	aktif, err := s.repo.PekerjaAktif(ctx, kasbon.PekerjaID)
	if err != nil {
		return 0, err
	}
	if !aktif {
		return 0, ErrPekerjaTidakAktif
	}

	kunci, err := s.terkunci(ctx, kasbon.Tanggal)
	if err != nil {
		return 0, err
	}
	if kunci {
		return 0, ErrPeriodeTertutup
	}

	kasbon.Sisa = kasbon.Jumlah // sisa awal sama dengan jumlah
	if err := s.repo.SimpanKasbon(ctx, &kasbon); err != nil {
		return 0, err
	}
	return kasbon.ID, nil
}

// DaftarKasbon mengambil daftar kasbon.
//
// Masukan: context dan Filter.
// Hasil: daftar kasbon beserta nama pekerja.
// Aturan: 12, 16.
// Tabel: kasbon, pekerja.
// Memanggil: Repository.DaftarKasbon.
func (s *Service) DaftarKasbon(ctx context.Context, filter Filter) ([]KasbonTampil, error) {
	return s.repo.DaftarKasbon(ctx, filter)
}

// AmbilKasbon mengambil satu kasbon.
//
// Masukan: context dan ID kasbon.
// Hasil: kasbon beserta nama pekerja; ErrKasbonTidakAda bila tidak ada.
// Aturan: 12.
// Tabel: kasbon, pekerja.
// Memanggil: Repository.AmbilKasbon.
func (s *Service) AmbilKasbon(ctx context.Context, id int64) (KasbonTampil, error) {
	return s.repo.AmbilKasbon(ctx, id)
}

// SisaKasbon menghasilkan total sisa kasbon pekerja.
//
// Masukan: context dan ID pekerja.
// Hasil: total sisa.
// Aturan: 12, 16.
// Tabel: kasbon.
// Memanggil: Repository.SisaKasbon.
func (s *Service) SisaKasbon(ctx context.Context, pekerjaID int64) (domain.Rupiah, error) {
	return s.repo.SisaKasbon(ctx, pekerjaID)
}

// KasbonDiprosesSampai menghasilkan kasbon yang ikut dipotong pada periode.
//
// Masukan: context, ID pekerja, batas tanggal (akhir periode).
// Hasil: kasbon bersisa, terurut dari yang tertua.
// Aturan: 12.
// Tabel: kasbon.
// Memanggil: Repository.KasbonDiprosesSampai.
func (s *Service) KasbonDiprosesSampai(ctx context.Context, pekerjaID int64, tanggal time.Time) ([]domain.Kasbon, error) {
	return s.repo.KasbonDiprosesSampai(ctx, pekerjaID, tanggal)
}

// RiwayatPotongan mengambil potongan kasbon per periode.
//
// Masukan: context dan ID kasbon.
// Hasil: daftar potongan.
// Aturan: 12, 13.
// Tabel: potongan_kasbon, slip_gaji, periode_gaji.
// Memanggil: Repository.RiwayatPotongan.
func (s *Service) RiwayatPotongan(ctx context.Context, kasbonID int64) ([]PotonganTampil, error) {
	return s.repo.RiwayatPotongan(ctx, kasbonID)
}

// DaftarPekerja mengambil pekerja untuk dropdown.
func (s *Service) DaftarPekerja(ctx context.Context, hanyaAktif bool) ([]domain.Pekerja, error) {
	return s.repo.DaftarPekerja(ctx, hanyaAktif)
}

// Pastikan Service memenuhi kontrak lintas modul.
var _ kontrak.KasbonProvider = (*Service)(nil)