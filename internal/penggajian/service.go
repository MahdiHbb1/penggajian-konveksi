// Pemilik: Farah | Modul 4 | Aturan: 5, 6, 7, 8, 9, 10, 12, 17, 18
package penggajian

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"penggajian-konveksi/internal/domain"
	pkgjadwal "penggajian-konveksi/internal/jadwal"
	"penggajian-konveksi/internal/kontrak"
)

// Pesan galat yang aman ditampilkan di halaman penggajian.
var (
	ErrPeriodeTidakValid = errors.New("Periode gaji tidak valid.")
	ErrPeriodeTertutup   = errors.New("Periode yang sudah ditutup tidak dapat dihitung.")
)

// GalatHitung adalah galat aturan bisnis dengan pesan untuk pengguna.
type GalatHitung struct {
	Pesan string
}

// Error mengembalikan pesan galat hitung.
func (g GalatHitung) Error() string {
	return g.Pesan
}

func salahHitung(pesan string) error {
	return GalatHitung{Pesan: pesan}
}

// PesanGalatHitung mengembalikan pesan bila err adalah galat yang boleh
// ditampilkan. Hasil kosong berarti galat sistem.
func PesanGalatHitung(err error) string {
	var galat GalatHitung
	if errors.As(err, &galat) {
		return galat.Pesan
	}
	switch {
	case errors.Is(err, ErrPeriodeTertutup):
		return ErrPeriodeTertutup.Error()
	case errors.Is(err, ErrPeriodeTidakValid):
		return ErrPeriodeTidakValid.Error()
	case errors.Is(err, domain.ErrTidakDitemukan):
		return "Data yang diminta tidak ditemukan."
	default:
		return ""
	}
}

// Repositori adalah kontrak penyimpanan yang dibutuhkan Service.
// Implementasi produksi: *Repository. Implementasi tes: fake in-memory.
type Repositori interface {
	SimpanSlipDanRincian(ctx context.Context, slip domain.SlipGaji, rincian []domain.SlipGajiDetail, potongan []domain.PotonganKasbon) error
	SimpanHasilPeriode(ctx context.Context, periodeID int64, hasil []HasilHitung) error
	AmbilSlip(ctx context.Context, periodeID, pekerjaID int64) (domain.SlipGaji, []domain.SlipGajiDetail, error)
	AmbilRincian(ctx context.Context, periodeID, pekerjaID int64) (domain.SlipGaji, []RincianHarian, error)
	DaftarSlip(ctx context.Context, periodeID int64) ([]domain.SlipGaji, error)
	DaftarPeriode(ctx context.Context) ([]domain.PeriodeGaji, error)
	AmbilPeriode(ctx context.Context, id int64) (domain.PeriodeGaji, error)
	PeriodeSebelumnya(ctx context.Context, periode domain.PeriodeGaji) (domain.PeriodeGaji, bool, error)
	PekerjaAktif(ctx context.Context, jenis domain.JenisPekerja, mulai, selesai time.Time) ([]domain.Pekerja, error)
	AbsensiPeriode(ctx context.Context, periodeID int64) ([]domain.AbsensiLengkap, error)
	AbsensiBelumLengkap(ctx context.Context, periodeID int64) (int, error)
	UpahSampai(ctx context.Context, pekerjaID int64, sampai time.Time) ([]domain.Upah, error)
	KasbonSampai(ctx context.Context, pekerjaID int64, sampai time.Time) ([]domain.Kasbon, error)
	SisaKasbon(ctx context.Context, pekerjaID int64) (domain.Rupiah, error)
	KoreksiUntukPeriode(ctx context.Context, periodeID, pekerjaID int64) ([]domain.KoreksiGaji, error)
	KoreksiDiterapkan(ctx context.Context, periodeID, pekerjaID int64) ([]domain.KoreksiGaji, error)
	TandaiKoreksiPeriode(ctx context.Context, periodeID int64) error
	TanggalTerkunci(ctx context.Context, tanggal time.Time) (bool, error)
	UbahStatusPeriode(ctx context.Context, periodeID int64, status domain.StatusPeriode) error
}

// Prasyarat adalah hasil pemeriksaan sebelum periode dihitung.
type Prasyarat struct {
	StatusPeriodeOK     bool
	PeriodeSebelumnyaOK bool
	AbsensiLengkap      bool
	JumlahAbsensiKurang int
	Pesan               string
}

// Lolos menandakan seluruh prasyarat terpenuhi.
func (p Prasyarat) Lolos() bool {
	return p.StatusPeriodeOK && p.PeriodeSebelumnyaOK && p.AbsensiLengkap
}

// Service menghitung gaji satu periode.
type Service struct {
	repo    Repositori
	pekerja kontrak.PekerjaProvider
	jadwal  kontrak.JadwalProvider
	kasbon  kontrak.KasbonProvider
	absensi kontrak.AbsensiProvider
	periode kontrak.PeriodeProvider
	koreksi kontrak.KoreksiProvider
}

// NewService membuat service penggajian dengan seluruh dependensi kontrak.
func NewService(
	repo Repositori,
	pekerja kontrak.PekerjaProvider,
	jadwal kontrak.JadwalProvider,
	kasbon kontrak.KasbonProvider,
	absensi kontrak.AbsensiProvider,
	periode kontrak.PeriodeProvider,
	koreksi kontrak.KoreksiProvider,
) *Service {
	return &Service{
		repo:    repo,
		pekerja: pekerja,
		jadwal:  jadwal,
		kasbon:  kasbon,
		absensi: absensi,
		periode: periode,
		koreksi: koreksi,
	}
}

// HitungGajiPeriode menghitung dan menyimpan gaji seluruh pekerja pada periode.
//
// Masukan: context dan ID periode.
// Hasil: error.
// Aturan: 5, 6, 7, 8, 9, 12, 17, 18.
// Tabel: periode_gaji, pekerja, riwayat_upah, jadwal_shift, absensi, kasbon, koreksi_gaji, slip_gaji, slip_gaji_detail, potongan_kasbon.
// Memanggil: PeriksaPrasyaratHitung, PekerjaProvider, AbsensiProvider, KasbonProvider, KoreksiProvider, Repository.SimpanHasilPeriode, PeriodeProvider.UbahStatusPeriode.
func (s *Service) HitungGajiPeriode(ctx context.Context, periodeID int64) error {
	if periodeID <= 0 {
		return ErrPeriodeTidakValid
	}
	if s.repo == nil {
		return errDatabaseTidakSiap
	}

	periode, err := s.ambilPeriode(ctx, periodeID)
	if err != nil {
		return err
	}
	if periode.Status == domain.PeriodeDitutup {
		return ErrPeriodeTertutup
	}

	prasyarat, err := s.PeriksaPrasyaratHitung(ctx, periodeID)
	if err != nil {
		return err
	}
	if !prasyarat.Lolos() {
		return salahHitung(prasyarat.Pesan)
	}

	jenis, err := jenisPekerjaUntukPeriode(periode.Tipe)
	if err != nil {
		return err
	}

	pekerja, err := s.daftarPekerjaAktif(ctx, jenis, periode.TanggalMulai, periode.TanggalSelesai)
	if err != nil {
		return err
	}
	absensi, err := s.absensiDalamPeriode(ctx, periodeID)
	if err != nil {
		return err
	}

	perPekerja := make(map[int64][]domain.AbsensiLengkap)
	for _, a := range absensi {
		pekerjaID := a.Jadwal.PekerjaID
		if pekerjaID == 0 {
			pekerjaID = a.Pekerja.ID
		}
		perPekerja[pekerjaID] = append(perPekerja[pekerjaID], a)
	}

	waktuHitung := time.Now()
	hasil := make([]HasilHitung, 0, len(pekerja))
	for _, p := range pekerja {
		h, ada, err := s.hitungPekerja(ctx, periode, p, perPekerja[p.ID], waktuHitung)
		if err != nil {
			return err
		}
		if ada {
			hasil = append(hasil, h)
		}
	}

	if err := s.repo.SimpanHasilPeriode(ctx, periodeID, hasil); err != nil {
		return err
	}
	if err := s.tandaiKoreksiDiterapkan(ctx, periodeID); err != nil {
		return err
	}
	return s.ubahStatusPeriode(ctx, periodeID, domain.PeriodeDihitung)
}

// PeriksaPrasyaratHitung memeriksa syarat hitung gaji.
//
// Masukan: context dan ID periode.
// Hasil: Prasyarat.
// Aturan: 2, 13.
// Tabel: periode_gaji, absensi, jadwal_shift.
// Memanggil: PeriodeProvider.AmbilPeriode, PeriodeProvider.PeriodeSebelumnyaDitutup, AbsensiProvider.AdaAbsensiBelumLengkap.
func (s *Service) PeriksaPrasyaratHitung(ctx context.Context, periodeID int64) (Prasyarat, error) {
	if periodeID <= 0 {
		return Prasyarat{}, ErrPeriodeTidakValid
	}
	periode, err := s.ambilPeriode(ctx, periodeID)
	if err != nil {
		return Prasyarat{}, err
	}

	hasil := Prasyarat{
		StatusPeriodeOK: periode.Status == domain.PeriodeTerbuka ||
			periode.Status == domain.PeriodeDihitung,
	}
	hasil.PeriodeSebelumnyaOK, err = s.periodeSebelumnyaDitutup(ctx, periode)
	if err != nil {
		return Prasyarat{}, err
	}
	adaKurang, jumlahKurang, err := s.adaAbsensiBelumLengkap(ctx, periodeID)
	if err != nil {
		return Prasyarat{}, err
	}
	hasil.AbsensiLengkap = !adaKurang
	hasil.JumlahAbsensiKurang = jumlahKurang

	var pesan []string
	if !hasil.StatusPeriodeOK {
		pesan = append(pesan, "Periode harus berstatus Terbuka atau Dihitung.")
	}
	if !hasil.PeriodeSebelumnyaOK {
		pesan = append(pesan, "Periode sebelumnya dengan tipe yang sama belum ditutup.")
	}
	if !hasil.AbsensiLengkap {
		pesan = append(pesan, fmt.Sprintf(
			"Masih ada %d absensi tanpa jam keluar.", jumlahKurang))
	}
	hasil.Pesan = strings.Join(pesan, " ")
	return hasil, nil
}

// SimpanSlipDanRincian menyimpan hasil hitung satu slip.
//
// Masukan: context, slip, rincian, potongan.
// Hasil: error.
// Aturan: 13, 18.
// Tabel: slip_gaji, slip_gaji_detail, potongan_kasbon.
// Memanggil: Repository.SimpanSlipDanRincian.
func (s *Service) SimpanSlipDanRincian(ctx context.Context, slip domain.SlipGaji, rincian []domain.SlipGajiDetail, potongan []domain.PotonganKasbon) error {
	if s.repo == nil {
		return errDatabaseTidakSiap
	}
	return s.repo.SimpanSlipDanRincian(ctx, slip, rincian, potongan)
}

// AmbilRincianGaji mengambil slip dan rincian satu pekerja.
//
// Masukan: context, ID periode, ID pekerja.
// Hasil: slip dan rincian harian.
// Aturan: 17.
// Tabel: slip_gaji, slip_gaji_detail.
// Memanggil: Repository.AmbilSlip.
func (s *Service) AmbilRincianGaji(ctx context.Context, periodeID, pekerjaID int64) (domain.SlipGaji, []domain.SlipGajiDetail, error) {
	if s.repo == nil {
		return domain.SlipGaji{}, nil, errDatabaseTidakSiap
	}
	if periodeID <= 0 || pekerjaID <= 0 {
		return domain.SlipGaji{}, nil, domain.ErrTidakDitemukan
	}
	return s.repo.AmbilSlip(ctx, periodeID, pekerjaID)
}

// AmbilRincianTampil mengambil slip dengan jam absensi dan nama shift.
//
// Masukan: context, ID periode, ID pekerja.
// Hasil: slip dan rincian harian untuk halaman rincian.
// Aturan: 17.
// Tabel: slip_gaji, slip_gaji_detail, absensi, jadwal_shift, shift.
// Memanggil: Repository.AmbilRincian.
func (s *Service) AmbilRincianTampil(ctx context.Context, periodeID, pekerjaID int64) (domain.SlipGaji, []RincianHarian, error) {
	if s.repo == nil {
		return domain.SlipGaji{}, nil, errDatabaseTidakSiap
	}
	if periodeID <= 0 || pekerjaID <= 0 {
		return domain.SlipGaji{}, nil, domain.ErrTidakDitemukan
	}
	return s.repo.AmbilRincian(ctx, periodeID, pekerjaID)
}

// DaftarSlip mengambil seluruh slip pada periode.
func (s *Service) DaftarSlip(ctx context.Context, periodeID int64) ([]domain.SlipGaji, error) {
	if s.repo == nil {
		return nil, errDatabaseTidakSiap
	}
	return s.repo.DaftarSlip(ctx, periodeID)
}

// DaftarPeriode mengambil periode untuk pilihan halaman penggajian.
func (s *Service) DaftarPeriode(ctx context.Context) ([]domain.PeriodeGaji, error) {
	if s.repo == nil {
		return nil, errDatabaseTidakSiap
	}
	return s.repo.DaftarPeriode(ctx)
}

// SisaKasbon menghasilkan total sisa kasbon pekerja.
func (s *Service) SisaKasbon(ctx context.Context, pekerjaID int64) (domain.Rupiah, error) {
	if s.kasbon != nil {
		total, err := s.kasbon.SisaKasbon(ctx, pekerjaID)
		if err == nil {
			return total, nil
		}
		if !providerBelumSiap(err) {
			return 0, err
		}
	}
	if s.repo == nil {
		return 0, errDatabaseTidakSiap
	}
	return s.repo.SisaKasbon(ctx, pekerjaID)
}

func (s *Service) hitungPekerja(
	ctx context.Context,
	periode domain.PeriodeGaji,
	pekerja domain.Pekerja,
	absensi []domain.AbsensiLengkap,
	dibuatPada time.Time,
) (HasilHitung, bool, error) {
	mulaiKerja := periode.TanggalMulai
	if pekerja.TanggalMulai.After(mulaiKerja) {
		mulaiKerja = pekerja.TanggalMulai
	}
	selesaiKerja := periode.TanggalSelesai
	if pekerja.TanggalBerhenti != nil && pekerja.TanggalBerhenti.Before(selesaiKerja) {
		selesaiKerja = *pekerja.TanggalBerhenti
	}
	mulaiKerja = tanggalSaja(mulaiKerja)
	selesaiKerja = tanggalSaja(selesaiKerja)
	if selesaiKerja.Before(mulaiKerja) {
		return HasilHitung{}, false, nil
	}

	riwayatUpah, err := s.repo.UpahSampai(ctx, pekerja.ID, selesaiKerja)
	if err != nil {
		return HasilHitung{}, false, err
	}
	if len(riwayatUpah) == 0 {
		return HasilHitung{}, false, salahHitung(fmt.Sprintf(
			"Upah pekerja %s belum diatur.", pekerja.Nama))
	}

	var rincian []domain.SlipGajiDetail
	for _, a := range absensi {
		if a.Absensi.JamKeluar == nil {
			continue
		}
		tanggal := tanggalSaja(a.Jadwal.Tanggal)
		if tanggal.IsZero() {
			tanggal = tanggalSaja(a.Absensi.JamMasuk)
		}
		if tanggal.Before(mulaiKerja) || tanggal.After(selesaiKerja) {
			continue
		}

		jadwal := a.Jadwal
		if jadwal.Shift.JamMulai == "" {
			jadwal, err = s.ambilJadwal(ctx, pekerja.ID, tanggal)
			if err != nil {
				return HasilHitung{}, false, err
			}
		}
		shiftMulai, err := pkgjadwal.WaktuShift(tanggal, jadwal.Shift.JamMulai)
		if err != nil {
			return HasilHitung{}, false, err
		}
		shiftSelesai, err := pkgjadwal.WaktuShift(tanggal, jadwal.Shift.JamSelesai)
		if err != nil {
			return HasilHitung{}, false, err
		}
		if !shiftSelesai.After(shiftMulai) {
			shiftSelesai = shiftSelesai.AddDate(0, 0, 1)
		}

		nominal := upahPadaTanggal(riwayatUpah, tanggal)
		upahPerJam := HitungUpahPerJam(nominal, pekerja.Jenis)
		menitTelat := hitungMenitTerlambat(a.Absensi.JamMasuk, shiftMulai)
		menitLembur := hitungMenitLembur(*a.Absensi.JamKeluar, shiftSelesai)

		detail := domain.SlipGajiDetail{
			AbsensiID:     a.Absensi.ID,
			Tanggal:       tanggal,
			UpahPerJam:    upahPerJam,
			MenitTelat:    menitTelat,
			PotonganTelat: HitungPotonganTerlambat(menitTelat, upahPerJam),
			MenitLembur:   menitLembur,
			UpahLembur:    HitungUpahLembur(menitLembur, upahPerJam),
		}
		if pekerja.Jenis == domain.PekerjaHarian {
			detail.UpahPokokHari = nominal
		}
		rincian = append(rincian, detail)
	}
	sort.Slice(rincian, func(i, j int) bool {
		return rincian[i].Tanggal.Before(rincian[j].Tanggal)
	})

	var (
		pokok            domain.Rupiah
		hariKerjaDibayar int
		hariKerjaPeriode int
	)
	switch pekerja.Jenis {
	case domain.PekerjaHarian:
		pokok = HitungUpahPokokHarian(rincian)
		hariKerjaDibayar = len(rincian)
	case domain.PekerjaBulanan:
		segmen := segmenUpahBulanan(periode, mulaiKerja, selesaiKerja, riwayatUpah)
		pokok = HitungUpahPokokBulanan(segmen)
		hariKerjaDibayar = HitungHariKerja(mulaiKerja, selesaiKerja)
		hariKerjaPeriode = HitungHariKerja(
			awalBulan(periode.TanggalMulai),
			akhirBulan(periode.TanggalMulai),
		)
	default:
		return HasilHitung{}, false, salahHitung("Jenis pekerja tidak dikenali.")
	}

	var totalLembur, potonganTelat domain.Rupiah
	for _, d := range rincian {
		totalLembur += d.UpahLembur
		potonganTelat += d.PotonganTelat
	}

	koreksi, err := s.koreksiUntukPekerja(ctx, periode.ID, pekerja.ID)
	if err != nil {
		return HasilHitung{}, false, err
	}
	var totalKoreksi domain.Rupiah
	for _, k := range koreksi {
		totalKoreksi += k.Jumlah
	}

	bruto := HitungGajiBruto(pokok, totalLembur, potonganTelat, totalKoreksi)
	daftarKasbon, err := s.kasbonSampai(ctx, pekerja.ID, periode.TanggalSelesai)
	if err != nil {
		return HasilHitung{}, false, err
	}
	potonganKasbon, totalPotongan := AlokasiPotonganKasbon(bruto, daftarKasbon)

	slip := domain.SlipGaji{
		PeriodeID:        periode.ID,
		PekerjaID:        pekerja.ID,
		UpahPokok:        pokok,
		TotalLembur:      totalLembur,
		PotonganTelat:    potonganTelat,
		TotalKoreksi:     totalKoreksi,
		GajiBruto:        bruto,
		PotonganKasbon:   totalPotongan,
		GajiBersih:       bruto - totalPotongan,
		HariKerjaDibayar: hariKerjaDibayar,
		HariKerjaPeriode: hariKerjaPeriode,
		DibuatPada:       dibuatPada,
		Pekerja:          &pekerja,
		Periode:          &periode,
	}
	return HasilHitung{Slip: slip, Rincian: rincian, Potongan: potonganKasbon}, true, nil
}

func segmenUpahBulanan(periode domain.PeriodeGaji, mulai, selesai time.Time, riwayat []domain.Upah) []SegmenUpah {
	hariKerjaBulan := HitungHariKerja(
		awalBulan(periode.TanggalMulai),
		akhirBulan(periode.TanggalMulai),
	)
	if hariKerjaBulan <= 0 {
		return nil
	}

	batas := []time.Time{mulai}
	for _, u := range riwayat {
		tanggal := tanggalSaja(u.BerlakuMulai)
		if tanggal.After(mulai) && !tanggal.After(selesai) {
			batas = append(batas, tanggal)
		}
	}
	batas = append(batas, selesai.AddDate(0, 0, 1))
	sort.Slice(batas, func(i, j int) bool { return batas[i].Before(batas[j]) })

	var segmen []SegmenUpah
	for i := 0; i < len(batas)-1; i++ {
		if !batas[i+1].After(batas[i]) {
			continue
		}
		segmenSelesai := batas[i+1].AddDate(0, 0, -1)
		nominal := upahPadaTanggal(riwayat, batas[i])
		if nominal <= 0 {
			continue
		}
		hari := HitungHariKerja(batas[i], segmenSelesai)
		if hari <= 0 {
			continue
		}
		segmen = append(segmen, SegmenUpah{
			Nominal:        nominal,
			HariDijalani:   hari,
			HariKerjaBulan: hariKerjaBulan,
		})
	}
	return segmen
}

func upahPadaTanggal(riwayat []domain.Upah, tanggal time.Time) domain.Rupiah {
	tanggal = tanggalSaja(tanggal)
	var nominal domain.Rupiah
	for _, u := range riwayat {
		if !tanggalSaja(u.BerlakuMulai).After(tanggal) {
			nominal = u.Nominal
		}
	}
	return nominal
}

func jenisPekerjaUntukPeriode(tipe domain.TipePeriode) (domain.JenisPekerja, error) {
	switch tipe {
	case domain.PeriodeMingguan:
		return domain.PekerjaHarian, nil
	case domain.PeriodeBulanan:
		return domain.PekerjaBulanan, nil
	default:
		return "", salahHitung("Tipe periode tidak dikenali.")
	}
}

func awalBulan(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func akhirBulan(t time.Time) time.Time {
	return awalBulan(t).AddDate(0, 1, -1)
}

// hitungMenitTerlambat dan hitungMenitLembur adalah salinan aturan Modul 3.
// AbsensiProvider sudah dipakai untuk data absensi, tetapi kedua metode menit
// di Modul 3 masih kerangka; helper ini menjaga hasil Modul 4 tetap benar.
func hitungMenitTerlambat(jamMasuk, jamMulaiShift time.Time) int {
	if !jamMasuk.After(jamMulaiShift) {
		return 0
	}
	return int(jamMasuk.Sub(jamMulaiShift) / time.Minute)
}

func hitungMenitLembur(jamKeluar, jamSelesaiShift time.Time) int {
	if !jamKeluar.After(jamSelesaiShift) {
		return 0
	}
	return int(jamKeluar.Sub(jamSelesaiShift) / time.Minute)
}

// ---- Pembungkus provider dengan cadangan repository ----

// providerBelumSiap menandai provider modul lain yang masih mengembalikan
// sentinel kerangka. Cadangan repository dipakai agar Modul 4 tetap berjalan
// sebelum Modul 3 dan 5 selesai; tidak ada fungsi penggajian yang mengembalikan
// sentinel ini.
func providerBelumSiap(err error) bool {
	return errors.Is(err, domain.ErrBelumDiimplementasi)
}

func (s *Service) ambilPeriode(ctx context.Context, periodeID int64) (domain.PeriodeGaji, error) {
	if s.periode != nil {
		periode, err := s.periode.AmbilPeriode(ctx, periodeID)
		if err == nil {
			return periode, nil
		}
		if !providerBelumSiap(err) {
			return domain.PeriodeGaji{}, err
		}
	}
	return s.repo.AmbilPeriode(ctx, periodeID)
}

func (s *Service) periodeSebelumnyaDitutup(ctx context.Context, periode domain.PeriodeGaji) (bool, error) {
	if s.periode != nil {
		ok, err := s.periode.PeriodeSebelumnyaDitutup(ctx, periode)
		if err == nil {
			return ok, nil
		}
		if !providerBelumSiap(err) {
			return false, err
		}
	}
	sebelumnya, ada, err := s.repo.PeriodeSebelumnya(ctx, periode)
	if err != nil {
		return false, err
	}
	if !ada {
		return true, nil
	}
	return sebelumnya.Status == domain.PeriodeDitutup, nil
}

func (s *Service) adaAbsensiBelumLengkap(ctx context.Context, periodeID int64) (bool, int, error) {
	if s.absensi != nil {
		ada, jumlah, err := s.absensi.AdaAbsensiBelumLengkap(ctx, periodeID)
		if err == nil {
			return ada, jumlah, nil
		}
		if !providerBelumSiap(err) {
			return false, 0, err
		}
	}
	jumlah, err := s.repo.AbsensiBelumLengkap(ctx, periodeID)
	if err != nil {
		return false, 0, err
	}
	return jumlah > 0, jumlah, nil
}

func (s *Service) daftarPekerjaAktif(ctx context.Context, jenis domain.JenisPekerja, mulai, selesai time.Time) ([]domain.Pekerja, error) {
	if s.pekerja != nil {
		daftar, err := s.pekerja.DaftarPekerjaAktif(ctx, jenis, mulai, selesai)
		if err == nil {
			return daftar, nil
		}
		if !providerBelumSiap(err) {
			return nil, err
		}
	}
	return s.repo.PekerjaAktif(ctx, jenis, mulai, selesai)
}

func (s *Service) absensiDalamPeriode(ctx context.Context, periodeID int64) ([]domain.AbsensiLengkap, error) {
	if s.absensi != nil {
		daftar, err := s.absensi.AbsensiDalamPeriode(ctx, periodeID)
		if err == nil {
			return daftar, nil
		}
		if !providerBelumSiap(err) {
			return nil, err
		}
	}
	return s.repo.AbsensiPeriode(ctx, periodeID)
}

func (s *Service) ambilJadwal(ctx context.Context, pekerjaID int64, tanggal time.Time) (domain.JadwalShift, error) {
	if s.jadwal == nil {
		return domain.JadwalShift{}, salahHitung("Jadwal shift pekerja tidak ditemukan.")
	}
	jadwal, err := s.jadwal.AmbilJadwal(ctx, pekerjaID, tanggal)
	if providerBelumSiap(err) {
		return domain.JadwalShift{}, salahHitung("Jadwal shift pekerja tidak ditemukan.")
	}
	return jadwal, err
}

func (s *Service) kasbonSampai(ctx context.Context, pekerjaID int64, tanggal time.Time) ([]domain.Kasbon, error) {
	if s.kasbon != nil {
		daftar, err := s.kasbon.KasbonDiprosesSampai(ctx, pekerjaID, tanggal)
		if err == nil {
			return daftar, nil
		}
		if !providerBelumSiap(err) {
			return nil, err
		}
	}
	return s.repo.KasbonSampai(ctx, pekerjaID, tanggal)
}

func (s *Service) koreksiUntukPekerja(ctx context.Context, periodeID, pekerjaID int64) ([]domain.KoreksiGaji, error) {
	hasil := make(map[int64]domain.KoreksiGaji)
	pakaiRepoLangsung := true

	if s.koreksi != nil {
		daftar, err := s.koreksi.DaftarKoreksiMenunggu(ctx, pekerjaID)
		switch {
		case err == nil:
			pakaiRepoLangsung = false
			for _, k := range daftar {
				hasil[k.ID] = k
			}
		case providerBelumSiap(err):
			pakaiRepoLangsung = true
		default:
			return nil, err
		}
	}

	var (
		daftar []domain.KoreksiGaji
		err    error
	)
	if pakaiRepoLangsung {
		daftar, err = s.repo.KoreksiUntukPeriode(ctx, periodeID, pekerjaID)
	} else {
		// Koreksi yang sudah menempel pada periode ini harus tetap ikut saat
		// hitung ulang, tetapi koreksi yang sudah menempel pada periode lain
		// tidak boleh dihitung lagi.
		daftar, err = s.repo.KoreksiDiterapkan(ctx, periodeID, pekerjaID)
	}
	if err != nil {
		return nil, err
	}
	for _, k := range daftar {
		hasil[k.ID] = k
	}

	koreksi := make([]domain.KoreksiGaji, 0, len(hasil))
	for _, k := range hasil {
		koreksi = append(koreksi, k)
	}
	sort.Slice(koreksi, func(i, j int) bool {
		if koreksi[i].DibuatPada.Equal(koreksi[j].DibuatPada) {
			return koreksi[i].ID < koreksi[j].ID
		}
		return koreksi[i].DibuatPada.Before(koreksi[j].DibuatPada)
	})
	return koreksi, nil
}

func (s *Service) tandaiKoreksiDiterapkan(ctx context.Context, periodeID int64) error {
	if s.koreksi != nil {
		err := s.koreksi.TandaiKoreksiDiterapkan(ctx, periodeID)
		if err == nil {
			return nil
		}
		if !providerBelumSiap(err) {
			return err
		}
	}
	return s.repo.TandaiKoreksiPeriode(ctx, periodeID)
}

func (s *Service) ubahStatusPeriode(ctx context.Context, periodeID int64, status domain.StatusPeriode) error {
	if s.periode != nil {
		err := s.periode.UbahStatusPeriode(ctx, periodeID, status)
		if err == nil {
			return nil
		}
		if !providerBelumSiap(err) {
			return err
		}
	}
	return s.repo.UbahStatusPeriode(ctx, periodeID, status)
}
