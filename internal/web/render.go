// Pemilik: Fadli | Modul 6 | Aturan: tidak ada aturan bisnis
package web

import (
	"bytes"
	"context"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"penggajian-konveksi/internal/domain"
)

// Data adalah data minimum yang diterima setiap template.
type Data struct {
	Judul    string
	Aktif    string
	Pengguna *domain.Pengguna
	Admin    bool
	Pemilik  bool
	Polos    bool
	Galat    string
	Info     string
	Isi      any
}

// Renderer dipakai handler untuk merender satu template halaman.
type Renderer interface {
	Render(w http.ResponseWriter, r *http.Request, namaTemplate string, data Data)
}

// Render memuat layout dan halaman dari direktori template.
type Render struct {
	Dir string
}

// Render merender halaman dengan layout.html.
//
// Masukan: response writer, request, nama file template, dan Data.
// Hasil: HTML 200, atau 500 bila template gagal dimuat.
// Aturan: handler menyiapkan data; template hanya menampilkan.
// Tabel: tidak ada.
// Memanggil: template.ParseFiles, template.ExecuteTemplate.
func (v *Render) Render(w http.ResponseWriter, r *http.Request, namaTemplate string, data Data) {
	data.Pengguna = PenggunaDariContext(r.Context())
	if data.Pengguna != nil {
		data.Admin = data.Pengguna.Role == domain.PeranAdmin
		data.Pemilik = data.Pengguna.Role == domain.PeranPemilik
	}

	layout := filepath.Join(v.Dir, "layout.html")
	halaman := filepath.Join(v.Dir, namaTemplate)
	tmpl, err := template.ParseFiles(layout, halaman)
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}

type kunciPengguna struct{}

// DenganPengguna menyimpan pengguna pada context request.
func DenganPengguna(ctx context.Context, pengguna *domain.Pengguna) context.Context {
	return context.WithValue(ctx, kunciPengguna{}, pengguna)
}

// PenggunaDariContext mengambil pengguna dari context request.
func PenggunaDariContext(ctx context.Context) *domain.Pengguna {
	pengguna, _ := ctx.Value(kunciPengguna{}).(*domain.Pengguna)
	return pengguna
}
