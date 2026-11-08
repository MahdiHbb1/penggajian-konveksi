// Pemilik: Mahdi | Modul 1 | Aturan: 1, 2, 8, 9, 11, 16
package pekerja

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"penggajian-konveksi/internal/domain"
)

// Repositori adalah kontrak penyimpanan yang dibutuhkan Service.
// Implementasi produksi: *Repository. Implementasi tes: fake in-memory.
type Repositori interface {
	SimpanPekerja(ctx context.Context, pekerja *domain.Pekerja) error
	SimpanPekerjaDanUpah(ctx context.Context, pekerja *domain.Pekerja, upah *domain.Upah) error
	UbahPekerja(ctx context.Context, pekerja domain.Pekerja) error
	NonaktifkanPekerja(ctx context.Context, id int64, tanggal time.Time) error
	DaftarPekerja(ctx context.Context, filter Filter) ([]domain.Pekerja, error)
	CariPekerja(ctx context.Context, id int64) (domain.Pekerja, error)
	KodeSudahDipakai(ctx context.Context, kode string, kecualiID int64) (bool, error)
	SimpanUpah(ctx context.Context, upah *domain.Upah) error
	DaftarUpah(ctx context.Context, pekerjaID int64) ([]domain.Upah, error)
	UpahBerlaku(ctx context.Context, pekerjaID int64, tanggal time.Time) (domain.Upah, error)
	UpahPadaTanggal(ctx context.Context, pekerjaID int64, tanggal time.Time) (bool, error)
	PekerjaAktif(ctx context.Context, jenis domain.JenisPekerja, mulai, selesai time.Time) ([]domain.Pekerja, error)
}

// Service mengatur aturan data pekerja dan riwayat upah.
type Service struct {
	repo Repositori
}

// GalatValidasi menandai kesalahan masukan pengguna yang aman ditampilkan.
type GalatValidasi struct {
	Pesan string
}

// Error mengembalikan pesan validasi.
func (g GalatValidasi) Error() string {
	return g.Pesan
}

func salahInput(pesan string) error {
	return GalatValidasi{Pesan: pesan}
}

func pesanGalatValidasi(err error) (string, bool) {
	var galat GalatValidasi
	if errors.As(err, &galat) {
		return galat.Pesan, true
	}
	return "", false
}

// NewService membuat service pekerja.
func NewService(repo Repositori) *Service {
	return &Service{repo: repo}
}

// TambahPekerja menyimpan pekerja baru beserta upah pertama.
//
// Masukan: context, data pekerja, nominal upah awal.
// Hasil: ID pekerja baru.
// Aturan: 1, 2, 8, 11, 16.
// Tabel: pekerja, riwayat_upah.
// Memanggil: Repositori.KodeSudahDipakai, Repositori.SimpanPekerjaDanUpah.
func (s *Service) TambahPekerja(ctx context.Context, pekerja domain.Pekerja, upahAwal domain.Rupiah) (int64, error) {
	normalisasiPekerja(&pekerja)
	if err := validasiPekerja(pekerja); err != nil {
		return 0, err
	}
	if pekerja.TanggalBerhenti != nil {
		return 0, salahInput("Pekerja baru tidak boleh memiliki tanggal berhenti.")
	}
	if upahAwal <= 0 {
		return 0, salahInput("Nominal upah harus lebih dari 0.")
	}

	dipakai, err := s.repo.KodeSudahDipakai(ctx, pekerja.Kode, 0)
	if err != nil {
		return 0, err
	}
	if dipakai {
		return 0, salahInput("Kode pekerja sudah dipakai.")
	}

	pekerja.Aktif = true
	upah := domain.Upah{
		Nominal:      upahAwal,
		BerlakuMulai: pekerja.TanggalMulai,
		Alasan:       "Upah awal",
	}
	if err := s.repo.SimpanPekerjaDanUpah(ctx, &pekerja, &upah); err != nil {
		return 0, err
	}
	return pekerja.ID, nil
}

// UbahPekerja memperbarui data pekerja.
//
// Masukan: context dan data pekerja.
// Hasil: error.
// Aturan: 11, 16.
// Tabel: pekerja.
// Memanggil: Repositori.CariPekerja, Repositori.KodeSudahDipakai, Repositori.UbahPekerja.
func (s *Service) UbahPekerja(ctx context.Context, pekerja domain.Pekerja) error {
	if pekerja.ID <= 0 {
		return salahInput("ID pekerja tidak valid.")
	}
	lama, err := s.repo.CariPekerja(ctx, pekerja.ID)
	if err != nil {
		return err
	}

	normalisasiPekerja(&pekerja)
	if err := validasiPekerja(pekerja); err != nil {
		return err
	}
	if pekerja.Jenis != lama.Jenis {
		return salahInput("Jenis pekerja tidak dapat diubah setelah disimpan.")
	}
	if !samaTanggal(pekerja.TanggalMulai, lama.TanggalMulai) {
		riwayat, err := s.repo.DaftarUpah(ctx, pekerja.ID)
		if err != nil {
			return err
		}
		if len(riwayat) > 0 {
			return salahInput("Tanggal mulai tidak dapat diubah karena sudah ada riwayat upah.")
		}
	}

	dipakai, err := s.repo.KodeSudahDipakai(ctx, pekerja.Kode, pekerja.ID)
	if err != nil {
		return err
	}
	if dipakai {
		return salahInput("Kode pekerja sudah dipakai.")
	}

	if pekerja.TanggalBerhenti != nil {
		pekerja.Aktif = false
	} else {
		pekerja.Aktif = lama.Aktif
	}
	return s.repo.UbahPekerja(ctx, pekerja)
}

// NonaktifkanPekerja menonaktifkan pekerja tanpa menghapus data.
//
// Masukan: context, ID pekerja, tanggal berhenti.
// Hasil: error.
// Aturan: 16.
// Tabel: pekerja.
// Memanggil: Repositori.CariPekerja, Repositori.NonaktifkanPekerja.
func (s *Service) NonaktifkanPekerja(ctx context.Context, id int64, tanggalBerhenti time.Time) error {
	if id <= 0 {
		return salahInput("ID pekerja tidak valid.")
	}
	lama, err := s.repo.CariPekerja(ctx, id)
	if err != nil {
		return err
	}
	if tanggalBerhenti.IsZero() {
		return salahInput("Tanggal berhenti wajib diisi.")
	}
	tanggalBerhenti = tanggalSaja(tanggalBerhenti)
	if tanggalBerhenti.Before(tanggalSaja(lama.TanggalMulai)) {
		return salahInput("Tanggal berhenti tidak boleh sebelum tanggal mulai.")
	}
	return s.repo.NonaktifkanPekerja(ctx, id, tanggalBerhenti)
}

// DaftarPekerja mengambil daftar pekerja.
//
// Masukan: context dan Filter.
// Hasil: daftar pekerja.
// Aturan: 11, 16.
// Tabel: pekerja.
// Memanggil: Repositori.DaftarPekerja.
func (s *Service) DaftarPekerja(ctx context.Context, filter Filter) ([]domain.Pekerja, error) {
	return s.repo.DaftarPekerja(ctx, filter)
}

// CariPekerja mengambil satu pekerja.
//
// Masukan: context dan ID pekerja.
// Hasil: Pekerja.
// Aturan: 16.
// Tabel: pekerja.
// Memanggil: Repositori.CariPekerja.
func (s *Service) CariPekerja(ctx context.Context, id int64) (domain.Pekerja, error) {
	if id <= 0 {
		return domain.Pekerja{}, salahInput("ID pekerja tidak valid.")
	}
	return s.repo.CariPekerja(ctx, id)
}

// TambahRiwayatUpah menambah upah baru pada tanggal berlaku tertentu.
//
// Masukan: context dan data upah.
// Hasil: error.
// Aturan: 8.
// Tabel: riwayat_upah, periode_gaji.
// Memanggil: Repositori.CariPekerja, Repositori.UpahPadaTanggal, Repositori.SimpanUpah.
func (s *Service) TambahRiwayatUpah(ctx context.Context, upah domain.Upah) error {
	if upah.PekerjaID <= 0 {
		return salahInput("Pekerja wajib dipilih.")
	}
	if upah.Nominal <= 0 {
		return salahInput("Nominal upah harus lebih dari 0.")
	}
	if upah.BerlakuMulai.IsZero() {
		return salahInput("Tanggal berlaku upah wajib diisi.")
	}
	upah.BerlakuMulai = tanggalSaja(upah.BerlakuMulai)
	upah.Alasan = strings.TrimSpace(upah.Alasan)
	if utf8.RuneCountInString(upah.Alasan) > 100 {
		return salahInput("Alasan perubahan upah maksimal 100 karakter.")
	}

	pekerja, err := s.repo.CariPekerja(ctx, upah.PekerjaID)
	if err != nil {
		return err
	}
	if upah.BerlakuMulai.Before(tanggalSaja(pekerja.TanggalMulai)) {
		return salahInput("Tanggal berlaku upah tidak boleh sebelum pekerja mulai bekerja.")
	}

	sudahAda, err := s.repo.UpahPadaTanggal(ctx, upah.PekerjaID, upah.BerlakuMulai)
	if err != nil {
		return err
	}
	if sudahAda {
		return salahInput("Sudah ada upah dengan tanggal berlaku yang sama.")
	}
	return s.repo.SimpanUpah(ctx, &upah)
}

// DaftarRiwayatUpah mengambil riwayat upah pekerja.
//
// Masukan: context dan ID pekerja.
// Hasil: daftar upah.
// Aturan: 8.
// Tabel: riwayat_upah.
// Memanggil: Repositori.DaftarUpah.
func (s *Service) DaftarRiwayatUpah(ctx context.Context, pekerjaID int64) ([]domain.Upah, error) {
	if pekerjaID <= 0 {
		return nil, salahInput("ID pekerja tidak valid.")
	}
	return s.repo.DaftarUpah(ctx, pekerjaID)
}

// CariUpahBerlaku menghasilkan nominal upah pada tanggal kerja.
//
// Masukan: context, ID pekerja, tanggal kerja.
// Hasil: nominal upah berlaku.
// Aturan: 7, 8, 9.
// Tabel: riwayat_upah.
// Memanggil: Repositori.UpahBerlaku.
func (s *Service) CariUpahBerlaku(ctx context.Context, pekerjaID int64, tanggal time.Time) (domain.Rupiah, error) {
	if pekerjaID <= 0 {
		return 0, salahInput("ID pekerja tidak valid.")
	}
	if tanggal.IsZero() {
		return 0, salahInput("Tanggal kerja wajib diisi.")
	}
	upah, err := s.repo.UpahBerlaku(ctx, pekerjaID, tanggalSaja(tanggal))
	if err != nil {
		return 0, err
	}
	return upah.Nominal, nil
}

// DaftarPekerjaAktif menghasilkan pekerja aktif pada rentang periode.
//
// Masukan: context, jenis pekerja, awal dan akhir periode.
// Hasil: daftar pekerja.
// Aturan: 9, 16, 17.
// Tabel: pekerja, riwayat_upah.
// Memanggil: Repositori.PekerjaAktif.
func (s *Service) DaftarPekerjaAktif(ctx context.Context, jenis domain.JenisPekerja, mulai, selesai time.Time) ([]domain.Pekerja, error) {
	if mulai.IsZero() || selesai.IsZero() {
		return nil, salahInput("Rentang tanggal periode wajib diisi.")
	}
	mulai = tanggalSaja(mulai)
	selesai = tanggalSaja(selesai)
	if mulai.After(selesai) {
		return nil, salahInput("Rentang tanggal periode tidak valid.")
	}
	if jenis != "" && jenis != domain.PekerjaHarian && jenis != domain.PekerjaBulanan {
		return nil, salahInput("Jenis pekerja harus Harian atau Bulanan.")
	}
	return s.repo.PekerjaAktif(ctx, jenis, mulai, selesai)
}

func normalisasiPekerja(pekerja *domain.Pekerja) {
	pekerja.Kode = strings.TrimSpace(pekerja.Kode)
	pekerja.Nama = strings.TrimSpace(pekerja.Nama)
	pekerja.TanggalMulai = tanggalSaja(pekerja.TanggalMulai)
	if pekerja.TanggalBerhenti != nil {
		berhenti := tanggalSaja(*pekerja.TanggalBerhenti)
		pekerja.TanggalBerhenti = &berhenti
	}
}

func validasiPekerja(pekerja domain.Pekerja) error {
	if pekerja.Kode == "" {
		return salahInput("Kode pekerja wajib diisi.")
	}
	if utf8.RuneCountInString(pekerja.Kode) > 20 {
		return salahInput("Kode pekerja maksimal 20 karakter.")
	}
	if pekerja.Nama == "" {
		return salahInput("Nama pekerja wajib diisi.")
	}
	if utf8.RuneCountInString(pekerja.Nama) > 100 {
		return salahInput("Nama pekerja maksimal 100 karakter.")
	}
	if pekerja.Jenis != domain.PekerjaHarian && pekerja.Jenis != domain.PekerjaBulanan {
		return salahInput("Jenis pekerja harus Harian atau Bulanan.")
	}
	if pekerja.TanggalMulai.IsZero() {
		return salahInput("Tanggal mulai kerja wajib diisi.")
	}
	if pekerja.TanggalBerhenti != nil &&
		tanggalSaja(*pekerja.TanggalBerhenti).Before(tanggalSaja(pekerja.TanggalMulai)) {
		return salahInput("Tanggal berhenti tidak boleh sebelum tanggal mulai.")
	}
	return nil
}

func tanggalSaja(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	tahun, bulan, hari := t.Date()
	return time.Date(tahun, bulan, hari, 0, 0, 0, 0, t.Location())
}

func samaTanggal(a, b time.Time) bool {
	return tanggalSaja(a).Equal(tanggalSaja(b))
}

// Pastikan Service memenuhi kontrak lintas modul.
var _ interface {
	CariUpahBerlaku(context.Context, int64, time.Time) (domain.Rupiah, error)
	DaftarPekerjaAktif(context.Context, domain.JenisPekerja, time.Time, time.Time) ([]domain.Pekerja, error)
} = (*Service)(nil)
