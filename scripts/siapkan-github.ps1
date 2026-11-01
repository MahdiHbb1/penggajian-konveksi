# Membuat label, issue, dan item GitHub Projects untuk enam modul.
# Jalankan dari root repo setelah gh login dan repo sudah di-push.
#
# Contoh:
#   .\scripts\siapkan-github.ps1 -ProjectNumber 1
#   .\scripts\siapkan-github.ps1 -BuatProject

[CmdletBinding()]
param(
    [string]$Owner = "@me",
    [string]$Repo = "",
    [string]$ProjectTitle = "Penggajian Konveksi",
    [int]$ProjectNumber = 0,
    [switch]$BuatProject,
    [string]$AnggotaFile = "docs/github-anggota.json"
)

$ErrorActionPreference = "Stop"

if (-not (Get-Command gh -ErrorAction SilentlyContinue)) {
    $kandidatGh = @(
        (Join-Path $env:ProgramFiles "GitHub CLI"),
        (Join-Path $env:LOCALAPPDATA "GitHubCLI")
    )
    foreach ($folder in $kandidatGh) {
        if (Test-Path (Join-Path $folder "gh.exe")) {
            $env:Path = "$folder;$env:Path"
            break
        }
    }
}

function Panggil-Gh {
    param([string[]]$Argumen)

    $hasil = & gh @Argumen 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "Perintah gagal: gh $($Argumen -join ' ')`n$($hasil -join "`n")"
    }
    return $hasil
}

if (-not (Get-Command gh -ErrorAction SilentlyContinue)) {
    throw "GitHub CLI belum terpasang. Install dari https://cli.github.com lalu buka terminal baru."
}

& gh auth status 2>&1 | Out-Null
if ($LASTEXITCODE -ne 0) {
    throw "Belum login GitHub. Jalankan: gh auth login"
}

$konfigurasi = $null
if (Test-Path $AnggotaFile) {
    $konfigurasi = Get-Content -Raw $AnggotaFile | ConvertFrom-Json
    if (-not $Repo -and $konfigurasi.repo) {
        $Repo = [string]$konfigurasi.repo
    }
    if ($ProjectNumber -eq 0 -and $konfigurasi.project_number) {
        $ProjectNumber = [int]$konfigurasi.project_number
    }
}

if (-not $Repo) {
    $Repo = (Panggil-Gh @("repo", "view", "--json", "nameWithOwner", "-q", ".nameWithOwner") | Select-Object -Last 1).ToString().Trim()
}
if ($Repo -notmatch "^[^/]+/[^/]+$") {
    throw "Repo tidak valid: '$Repo'. Isi docs/github-anggota.json atau pakai -Repo OWNER/NAMA."
}

Write-Host "Repo       : $Repo"
Write-Host "Pemilik    : $Owner"

if ($BuatProject) {
    $json = Panggil-Gh @("project", "create", "--owner", $Owner, "--title", $ProjectTitle, "--format", "json")
    $project = ($json -join "`n") | ConvertFrom-Json
    $ProjectNumber = [int]$project.number
    Write-Host "Project    : #$ProjectNumber ($ProjectTitle)"
} elseif ($ProjectNumber -gt 0) {
    Write-Host "Project    : #$ProjectNumber"
} else {
    Write-Host "Project    : tidak dipakai, issue dibuat tanpa project"
}

$label = @(
    @{ Nama = "modul-1"; Warna = "1D76DB"; Deskripsi = "Modul 1 pekerja dan riwayat upah" },
    @{ Nama = "modul-2"; Warna = "5319E7"; Deskripsi = "Modul 2 jadwal shift dan kasbon" },
    @{ Nama = "modul-3"; Warna = "0E8A16"; Deskripsi = "Modul 3 absensi" },
    @{ Nama = "modul-4"; Warna = "B60205"; Deskripsi = "Modul 4 perhitungan gaji" },
    @{ Nama = "modul-5"; Warna = "D93F0B"; Deskripsi = "Modul 5 periode, penutupan, koreksi" },
    @{ Nama = "modul-6"; Warna = "006B75"; Deskripsi = "Modul 6 slip, rekap, autentikasi, antarmuka" },
    @{ Nama = "implementasi"; Warna = "FBCA04"; Deskripsi = "Pekerjaan implementasi fitur" },
    @{ Nama = "tes"; Warna = "C2E0C6"; Deskripsi = "Pekerjaan pengujian" },
    @{ Nama = "dokumen"; Warna = "BFDADC"; Deskripsi = "Pekerjaan dokumentasi" }
)

foreach ($l in $label) {
    Panggil-Gh @("label", "create", $l.Nama, "--repo", $Repo, "--color", $l.Warna, "--description", $l.Deskripsi, "--force") | Out-Null
}
Write-Host "Label      : $($label.Count) siap"

if ($ProjectNumber -gt 0) {
    & gh project link $ProjectNumber --owner $Owner --repo $Repo 2>&1 | Out-Null

    $field = @(
        @{ Nama = "Modul"; Tipe = "SINGLE_SELECT"; Opsi = "1 Pekerja,2 Jadwal dan Kasbon,3 Absensi,4 Penggajian,5 Periode dan Koreksi,6 Laporan dan Antarmuka" },
        @{ Nama = "PIC"; Tipe = "TEXT"; Opsi = $null },
        @{ Nama = "Branch"; Tipe = "TEXT"; Opsi = $null }
    )

    foreach ($f in $field) {
        $argumen = @("project", "field-create", "$ProjectNumber", "--owner", $Owner, "--name", $f.Nama, "--data-type", $f.Tipe)
        if ($f.Opsi) {
            $argumen += @("--single-select-options", $f.Opsi)
        }
        & gh @argumen 2>&1 | Out-Null
    }
    Write-Host "Field      : Modul, PIC, Branch siap atau sudah ada"
}

$modul = @(
    @{ Nomor = 1; PIC = "Mahdi"; Judul = "[Modul 1] Pekerja dan riwayat upah"; File = "docs/issues/modul-1.md" },
    @{ Nomor = 2; PIC = "Edgar"; Judul = "[Modul 2] Jadwal shift dan kasbon"; File = "docs/issues/modul-2.md" },
    @{ Nomor = 3; PIC = "Zahra"; Judul = "[Modul 3] Absensi"; File = "docs/issues/modul-3.md" },
    @{ Nomor = 4; PIC = "Farah"; Judul = "[Modul 4] Perhitungan gaji"; File = "docs/issues/modul-4.md" },
    @{ Nomor = 5; PIC = "Anya"; Judul = "[Modul 5] Periode, penutupan, dan koreksi"; File = "docs/issues/modul-5.md" },
    @{ Nomor = 6; PIC = "Fadli"; Judul = "[Modul 6] Slip, rekap, autentikasi, dan antarmuka"; File = "docs/issues/modul-6.md" }
)

$semuaIssue = Panggil-Gh @("issue", "list", "--repo", $Repo, "--state", "all", "--limit", "100", "--json", "number,title,url")
$issueLama = @()
if ($semuaIssue) {
    $issueLama = (($semuaIssue -join "`n") | ConvertFrom-Json)
}

foreach ($m in $modul) {
    if (-not (Test-Path $m.File)) {
        throw "File issue tidak ditemukan: $($m.File)"
    }

    $cocok = $issueLama | Where-Object { $_.title -eq $m.Judul } | Select-Object -First 1
    if ($cocok) {
        $url = $cocok.url
        $nomor = $cocok.number
        Write-Host "Issue #$nomor sudah ada: $($m.Judul)"
    } else {
        $keluaran = Panggil-Gh @("issue", "create", "--repo", $Repo, "--title", $m.Judul, "--body-file", $m.File, "--label", "modul-$($m.Nomor)", "--label", "implementasi")
        $url = ($keluaran | Where-Object { $_ -match "^https://github.com/" } | Select-Object -Last 1)
        if (-not $url) {
            throw "Gagal membaca URL issue baru untuk $($m.Judul)"
        }
        $url = $url.ToString().Trim()
        $nomor = [int](($url -split "/")[-1])
        Write-Host "Issue #$nomor dibuat: $($m.Judul)"
    }

    if ($ProjectNumber -gt 0) {
        Panggil-Gh @("project", "item-add", "$ProjectNumber", "--owner", $Owner, "--url", $url) | Out-Null
    }

    if ($konfigurasi -and $konfigurasi.anggota) {
        $properti = $konfigurasi.anggota.PSObject.Properties[$m.PIC]
        if ($properti -and $properti.Value) {
            Panggil-Gh @("issue", "edit", "$nomor", "--repo", $Repo, "--add-assignee", [string]$properti.Value) | Out-Null
            Write-Host "  assign ke $($properti.Value)"
        }
    }
}

Write-Host ""
Write-Host "Selesai. Buka project dan atur view Table, Board status, dan Board per PIC."
Write-Host "Jangan jalankan ulang jika tidak perlu, issue yang sudah ada akan dilewati tetapi item project bisa ditambahkan lagi."
