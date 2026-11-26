// Pemilik: Fadli | Modul 6 | Aturan: tidak ada aturan bisnis
package web

import (
	"bytes"
	"context"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"penggajian-konveksi/internal/domain"
)

// Data adalah data minimum yang diterima setiap template.
type Data struct {
	Judul     string
	Aktif     string
	Pengguna  *domain.Pengguna
	Admin     bool
	Pemilik   bool
	Polos     bool
	Galat     string
	Info      string
	Isi       any
	CSRFToken string
	Status    int
}

// Renderer dipakai handler untuk merender satu template halaman.
type Renderer interface {
	Render(w http.ResponseWriter, r *http.Request, namaTemplate string, data Data)
}

// Render memuat dan men-cache layout serta halaman dari direktori template.
type Render struct {
	Dir   string
	mu    sync.RWMutex
	cache map[string]*template.Template
}

var fungsiTemplate = template.FuncMap{
	"rupiah": FormatRupiah,
}

// FormatRupiah memformat nominal Rupiah dengan pemisah ribuan titik (contoh: "Rp 441.875").
func FormatRupiah(nominal domain.Rupiah) string {
	neg := nominal < 0
	nilai := int64(nominal)
	if neg {
		nilai = -nilai
	}
	angka := strconv.FormatInt(nilai, 10)
	var b strings.Builder
	for i := 0; i < len(angka); i++ {
		if i > 0 && (len(angka)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteByte(angka[i])
	}
	if neg {
		return "-Rp " + b.String()
	}
	return "Rp " + b.String()
}

func (v *Render) ambilTemplate(namaTemplate string) (*template.Template, error) {
	v.mu.RLock()
	if v.cache != nil {
		if tmpl, ada := v.cache[namaTemplate]; ada {
			v.mu.RUnlock()
			return tmpl, nil
		}
	}
	v.mu.RUnlock()

	layout := filepath.Join(v.Dir, "layout.html")
	halaman := filepath.Join(v.Dir, namaTemplate)
	tmpl, err := template.New("layout.html").Funcs(fungsiTemplate).ParseFiles(layout, halaman)
	if err != nil {
		return nil, err
	}

	v.mu.Lock()
	if v.cache == nil {
		v.cache = make(map[string]*template.Template)
	}
	v.cache[namaTemplate] = tmpl
	v.mu.Unlock()
	return tmpl, nil
}

// Render merender halaman dengan layout.html.
//
// Masukan: response writer, request, nama file template, dan Data.
// Hasil: HTML 200 (atau data.Status), atau 500 bila template gagal dimuat.
// Aturan: handler menyiapkan data; template hanya menampilkan.
// Tabel: tidak ada.
// Memanggil: template.ParseFiles (cached), template.ExecuteTemplate.
func (v *Render) Render(w http.ResponseWriter, r *http.Request, namaTemplate string, data Data) {
	data.Pengguna = PenggunaDariContext(r.Context())
	if data.Pengguna != nil {
		data.Admin = data.Pengguna.Role == domain.PeranAdmin
		data.Pemilik = data.Pengguna.Role == domain.PeranPemilik
	}
	// Ambil CSRF token dari context jika belum di-set oleh handler.
	if data.CSRFToken == "" {
		data.CSRFToken = CSRFTokenDariContext(r.Context())
	}

	tmpl, err := v.ambilTemplate(namaTemplate)
	if err != nil {
		log.Printf("GAGAL memuat template %s: %v", namaTemplate, err)
		http.Error(w, "Template belum siap: "+namaTemplate, http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout.html", data); err != nil {
		log.Printf("GAGAL merender template %s: %v", namaTemplate, err)
		http.Error(w, "Gagal merender halaman.", http.StatusInternalServerError)
		return
	}

	status := data.Status
	if status == 0 {
		status = http.StatusOK
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}

type kunciPengguna struct{}

type kunciCSRFToken struct{}

// DenganPengguna menyimpan pengguna pada context request.
func DenganPengguna(ctx context.Context, pengguna *domain.Pengguna) context.Context {
	return context.WithValue(ctx, kunciPengguna{}, pengguna)
}

// PenggunaDariContext mengambil pengguna dari context request.
func PenggunaDariContext(ctx context.Context) *domain.Pengguna {
	pengguna, _ := ctx.Value(kunciPengguna{}).(*domain.Pengguna)
	return pengguna
}

// DenganCSRFToken menyimpan CSRF token pada context request.
func DenganCSRFToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, kunciCSRFToken{}, token)
}

// CSRFTokenDariContext mengambil CSRF token dari context request.
func CSRFTokenDariContext(ctx context.Context) string {
	token, _ := ctx.Value(kunciCSRFToken{}).(string)
	return token
}
