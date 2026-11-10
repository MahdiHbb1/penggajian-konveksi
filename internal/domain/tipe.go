// Pemilik: Mahdi | Modul 1 | Aturan: 1, 7, 11, 13, 16
package domain

import (
	"errors"
	"time"
)

var (
	// ErrBelumDiimplementasi dipakai seluruh badan fungsi kerangka.
	// Isi fungsi anggota diganti oleh pemilik modul masing-masing.
	ErrBelumDiimplementasi = errors.New("belum diimplementasi")
	// ErrTidakDitemukan dipakai bila data yang diminta tidak ada.
	ErrTidakDitemukan = errors.New("data tidak ditemukan")
	// ErrDuplikat dipakai bila data unik sudah dipakai.
	ErrDuplikat = errors.New("data sudah ada")
)

// Rupiah adalah nominal uang dalam bilangan bulat rupiah.
// Jangan memakai float64 untuk uang.
type Rupiah int64

// JenisPekerja membedakan pekerja harian dan bulanan.
type JenisPekerja string

const (
	PekerjaHarian  JenisPekerja = "HARIAN"
	PekerjaBulanan JenisPekerja = "BULANAN"
)

// TipePeriode menentukan rentang periode gaji.
type TipePeriode string

const (
	PeriodeMingguan TipePeriode = "MINGGUAN"
	PeriodeBulanan  TipePeriode = "BULANAN"
)

// StatusPeriode adalah status hidup periode gaji.
type StatusPeriode string

const (
	PeriodeTerbuka  StatusPeriode = "TERBUKA"
	PeriodeDihitung StatusPeriode = "DIHITUNG"
	PeriodeDitutup  StatusPeriode = "DITUTUP"
)

// Peran menentukan hak akses pengguna.
type Peran string

const (
	PeranAdmin   Peran = "ADMIN"
	PeranPemilik Peran = "PEMILIK"
)

// Pekerja adalah data induk pekerja konveksi.
type Pekerja struct {
	ID              int64
	Kode            string
	Nama            string
	Jenis           JenisPekerja
	TanggalMulai    time.Time
	TanggalBerhenti *time.Time
	Aktif           bool
}

// Status menurunkan label aktif atau nonaktif dari kolom aktif.
func (p Pekerja) Status() string {
	if p.Aktif {
		return "Aktif"
	}
	return "Nonaktif"
}

// Upah adalah satu baris riwayat upah pekerja.
type Upah struct {
	ID           int64
	PekerjaID    int64
	Nominal      Rupiah
	BerlakuMulai time.Time
	Alasan       string
}

// Shift adalah definisi jam kerja, misalnya Pagi 07.00-15.00.
type Shift struct {
	ID         int64
	Nama       string
	JamMulai   string
	JamSelesai string
}

// JadwalShift menghubungkan pekerja, shift, dan tanggal kerja.
type JadwalShift struct {
	ID        int64
	PekerjaID int64
	ShiftID   int64
	Tanggal   time.Time
	Pekerja   Pekerja
	Shift     Shift
}

// Absensi menyimpan jam masuk dan jam keluar sebagai datetime.
type Absensi struct {
	ID            int64
	JadwalShiftID int64
	JamMasuk      time.Time
	JamKeluar     *time.Time
}

// AbsensiLengkap adalah absensi dengan jadwal dan pekerja untuk tampilan.
type AbsensiLengkap struct {
	Absensi
	Jadwal  JadwalShift
	Pekerja Pekerja
}

// PeriodeGaji adalah rentang perhitungan gaji.
type PeriodeGaji struct {
	ID             int64
	Tipe           TipePeriode
	TanggalMulai   time.Time
	TanggalSelesai time.Time
	Status         StatusPeriode
	DitutupPada    *time.Time
}

// Kasbon adalah pinjaman pekerja yang dipotong dari gaji.
type Kasbon struct {
	ID         int64
	PekerjaID  int64
	Tanggal    time.Time
	Jumlah     Rupiah
	Sisa       Rupiah
	Keterangan string
}

// Status menurunkan label Berjalan atau Lunas dari kolom sisa.
func (k Kasbon) Status() string {
	if k.Sisa > 0 {
		return "Berjalan"
	}
	return "Lunas"
}

// KoreksiGaji mencatat selisih untuk periode yang sudah ditutup.
type KoreksiGaji struct {
	ID               int64
	PekerjaID        int64
	PeriodeAsalID    int64
	PeriodeTerapanID *int64
	Jumlah           Rupiah
	Keterangan       string
	DibuatPada       time.Time
}

// Status menurunkan label Menunggu atau Diterapkan dari periode terapan.
func (k KoreksiGaji) Status() string {
	if k.PeriodeTerapanID == nil {
		return "Menunggu"
	}
	return "Diterapkan"
}

// SlipGaji adalah snapshot hasil perhitungan satu pekerja pada satu periode.
type SlipGaji struct {
	ID               int64
	PeriodeID        int64
	PekerjaID        int64
	UpahPokok        Rupiah
	TotalLembur      Rupiah
	PotonganTelat    Rupiah
	TotalKoreksi     Rupiah
	GajiBruto        Rupiah
	PotonganKasbon   Rupiah
	GajiBersih       Rupiah
	HariKerjaDibayar int
	HariKerjaPeriode int
	DibuatPada       time.Time
	Pekerja          *Pekerja
	Periode          *PeriodeGaji
}

// SlipGajiDetail adalah rincian upah per hari kerja.
type SlipGajiDetail struct {
	ID            int64
	SlipID        int64
	AbsensiID     int64
	Tanggal       time.Time
	UpahPokokHari Rupiah
	UpahPerJam    Rupiah
	MenitTelat    int
	PotonganTelat Rupiah
	MenitLembur   int
	UpahLembur    Rupiah
}

// PotonganKasbon mencatat kasbon yang dipotong pada satu slip.
type PotonganKasbon struct {
	ID       int64
	SlipID   int64
	KasbonID int64
	Jumlah   Rupiah
}

// RekapGaji adalah ringkasan seluruh slip pada satu periode.
type RekapGaji struct {
	PeriodeID           int64
	TotalBruto          Rupiah
	TotalLembur         Rupiah
	TotalPotonganKasbon Rupiah
	TotalBersih         Rupiah
	JumlahSlip          int
}

// Pengguna adalah akun login aplikasi.
type Pengguna struct {
	ID           int64
	Username     string
	PasswordHash string
	Role         Peran
}

// NamaPeran menghasilkan label peran untuk tampilan.
func (p Pengguna) NamaPeran() string {
	if p.Role == PeranAdmin {
		return "Admin"
	}
	if p.Role == PeranPemilik {
		return "Pemilik"
	}
	return string(p.Role)
}
