# Helper bersama untuk skrip bukti temuan B1 (slice S-005).
#
# Fungsi: menyalin repo ke folder temp, menyuntikkan versi rusak, menjalankan
# perintah bukti, lalu memulihkan (menghapus folder temp). Working tree asli
# tidak pernah disentuh.
#
# Pakai (Windows PowerShell):
#   powershell -ExecutionPolicy Bypass -File tools\evidence\bug01-*.ps1

$script:RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path

# New-TempRepo menyalin repo ke %TEMP%\<nama> (tanpa .git).
function New-TempRepo([string]$nama) {
    $dest = Join-Path $env:TEMP $nama
    if (Test-Path $dest) { Remove-Item $dest -Recurse -Force }
    robocopy $script:RepoRoot $dest /E /XD .git logs /NFL /NDL /NJH /NJS /NP | Out-Null
    if ($LASTEXITCODE -ge 8) { throw "robocopy gagal: $LASTEXITCODE" }
    return $dest
}

# Remove-TempRepo menghapus salinan temp.
function Remove-TempRepo([string]$dest) {
    if (Test-Path $dest) { Remove-Item $dest -Recurse -Force }
}

# Replace-Teks mengganti potongan teks pada file (literal, bukan regex).
function Replace-Teks([string]$file, [string]$old, [string]$new, [bool]$sekali = $false) {
    $c = [IO.File]::ReadAllText($file)
    if ($sekali) {
        $idx = $c.IndexOf($old)
        if ($idx -lt 0) { throw "pola tidak ditemukan di $file : $old" }
        $c = $c.Remove($idx, $old.Length).Insert($idx, $new)
    } else {
        if (-not $c.Contains($old)) { throw "pola tidak ditemukan di $file : $old" }
        $c = $c.Replace($old, $new)
    }
    [IO.File]::WriteAllText($file, $c)
}

# Invoke-Go menjalankan perintah go di folder tertentu.
function Invoke-Go([string]$workdir, [string[]]$argsGo) {
    # Output go (termasuk stderr) ditangkap sebagai teks; error native tidak
    # dimatikan supaya pesan compile error ikut terbaca.
    $oldPref = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    Push-Location $workdir
    try {
        $lines = @(& go @argsGo 2>&1 | ForEach-Object { "$_" })
        $code = $LASTEXITCODE
    } finally {
        Pop-Location
        $ErrorActionPreference = $oldPref
    }
    return @{ Output = (($lines -join "`n") + "`n"); Exit = $code }
}

# Tampilkan-GitStatus membuktikan working tree asli tidak berubah.
function Tampilkan-GitStatus {
    $s = (& git -C $script:RepoRoot status --porcelain 2>&1 | Out-String)
    if ([string]::IsNullOrWhiteSpace($s)) {
        Write-Host "git status: BERSIH (working tree asli tidak berubah oleh skrip ini)"
    } else {
        Write-Host "git status:"
        Write-Host $s
    }
}

# Set-TestDatabaseUrl mengisi DSN database uji bila belum ada.
function Set-TestDatabaseUrl {
    if (-not $env:TEST_DATABASE_URL) {
        $env:TEST_DATABASE_URL = "postgres://postgres:@localhost:5432/praktikum_backend_test?sslmode=disable"
    }
}