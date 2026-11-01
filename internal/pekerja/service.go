// Pemilik: Mahdi | Modul 1 | Aturan: 1, 2, 8, 9, 11, 16
package pekerja

import (
	"context"
	"time"

	"penggajian-konveksi/internal/domain"
)

// Service mengatur aturan data pekerja dan riwayat upah.
type Service struct {
	repo *Repository
}

// NewService membuat service pekerja.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// TambahPekerja menyimpan pekerja baru beserta upah pertama.
//
// Masukan: context, data pekerja, nominal upah awal.
// Hasil: ID pekerja baru.
// Aturan: 1, 2, 8, 11, 16.
// Tabel: pekerja, riwayat_upah.
// Memanggil: Repository.SimpanPekerja, Repository.SimpanUpah.
//
// TODO(Mahdi):
//  1. Validasi kode, nama, jenis, tanggal mulai, dan nominal upah.
//  2. Cek kode unik.
//  3. Buka transaksi, simpan pekerja, simpan upah tanggal mulai, commit.
func (s *Service) TambahPekerja(ctx context.Context, pekerja domain.Pekerja, upahAwal domain.Rupiah) (int64, error) {
	return 0, domain.ErrBelumDiimplementasi
}

// UbahPekerja memperbarui data pekerja.
//
// Masukan: context dan data pekerja.
// Hasil: error.
// Aturan: 11, 16.
// Tabel: pekerja.
// Memanggil: Repository.CariPekerja, Repository.UbahPekerja.
//
// TODO(Mahdi):
//  1. Pastikan pekerja ada.
//  2. Tolak perubahan jenis pekerja.
//  3. Validasi tanggal berhenti tidak sebelum tanggal mulai.
func (s *Service) UbahPekerja(ctx context.Context, pekerja domain.Pekerja) error {
	return domain.ErrBelumDiimplementasi
}

// NonaktifkanPekerja menonaktifkan pekerja tanpa menghapus data.
//
// Masukan: context, ID pekerja, tanggal berhenti.
// Hasil: error.
// Aturan: 16.
// Tabel: pekerja.
// Memanggil: Repository.NonaktifkanPekerja.
//
// TODO(Mahdi):
//  1. Pastikan tanggal berhenti tidak sebelum tanggal mulai.
//  2. Isi tanggal_berhenti dan aktif=false.
//  3. Jangan menghapus baris.
func (s *Service) NonaktifkanPekerja(ctx context.Context, id int64, tanggal time.Time) error {
	return domain.ErrBelumDiimplementasi
}

// DaftarPekerja mengambil daftar pekerja.
//
// Masukan: context dan Filter.
// Hasil: daftar pekerja.
// Aturan: 11, 16.
// Tabel: pekerja.
// Memanggil: Repository.DaftarPekerja.
func (s *Service) DaftarPekerja(ctx context.Context, filter Filter) ([]domain.Pekerja, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// CariPekerja mengambil satu pekerja.
//
// Masukan: context dan ID pekerja.
// Hasil: Pekerja.
// Aturan: 16.
// Tabel: pekerja.
// Memanggil: Repository.CariPekerja.
func (s *Service) CariPekerja(ctx context.Context, id int64) (domain.Pekerja, error) {
	return domain.Pekerja{}, domain.ErrBelumDiimplementasi
}

// TambahRiwayatUpah menambah upah baru pada tanggal berlaku tertentu.
//
// Masukan: context dan data upah.
// Hasil: error.
// Aturan: 8.
// Tabel: riwayat_upah, periode_gaji.
// Memanggil: Repository.CariPekerja, Repository.SimpanUpah, PeriodeProvider.PeriksaPeriodeTerkunci.
//
// TODO(Mahdi):
//  1. Validasi nominal lebih dari 0 dan tanggal berlaku.
//  2. Tolak tanggal ganda untuk pekerja yang sama.
//  3. Tolak tanggal sebelum pekerja mulai bekerja.
//  4. Beri peringatan bila tanggal jatuh di periode tertutup.
func (s *Service) TambahRiwayatUpah(ctx context.Context, upah domain.Upah) error {
	return domain.ErrBelumDiimplementasi
}

// DaftarRiwayatUpah mengambil riwayat upah pekerja.
//
// Masukan: context dan ID pekerja.
// Hasil: daftar upah.
// Aturan: 8.
// Tabel: riwayat_upah.
// Memanggil: Repository.DaftarUpah.
func (s *Service) DaftarRiwayatUpah(ctx context.Context, pekerjaID int64) ([]domain.Upah, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// CariUpahBerlaku menghasilkan nominal upah pada tanggal kerja.
//
// Masukan: context, ID pekerja, tanggal kerja.
// Hasil: nominal upah berlaku.
// Aturan: 7, 8, 9.
// Tabel: riwayat_upah.
// Memanggil: Repository.UpahBerlaku.
func (s *Service) CariUpahBerlaku(ctx context.Context, pekerjaID int64, tanggal time.Time) (domain.Rupiah, error) {
	return 0, domain.ErrBelumDiimplementasi
}

// DaftarPekerjaAktif menghasilkan pekerja aktif pada rentang periode.
//
// Masukan: context, jenis pekerja, awal dan akhir periode.
// Hasil: daftar pekerja.
// Aturan: 9, 16, 17.
// Tabel: pekerja, riwayat_upah.
// Memanggil: Repository.PekerjaAktif.
func (s *Service) DaftarPekerjaAktif(ctx context.Context, jenis domain.JenisPekerja, mulai, selesai time.Time) ([]domain.Pekerja, error) {
	return nil, domain.ErrBelumDiimplementasi
}

// Pastikan Service memenuhi kontrak lintas modul.
var _ interface {
	CariUpahBerlaku(context.Context, int64, time.Time) (domain.Rupiah, error)
	DaftarPekerjaAktif(context.Context, domain.JenisPekerja, time.Time, time.Time) ([]domain.Pekerja, error)
} = (*Service)(nil)
