# BUG 06 (B1): cfg.MaxConns = 1 -> uji konkurensi tidak benar-benar konkuren,
# sehingga mutasi "hapus FOR UPDATE" (T6) TIDAK tertangkap test.
# Perbaikan: cfg.MaxConns = 10.
#
# Skrip ini membuktikan gejalanya: mutasi T6 + MaxConns = 1 -> test HIJAU
# (mutan lolos), mutasi T6 + MaxConns = 10 -> test MERAH (mutan tertangkap).
$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "bukti-helper.ps1")
Set-TestDatabaseUrl

$argsGo = @("test", "-tags=integration", "-count=3", "-run", "Concurrent", "./tests/integration/")

Write-Host "===== BUG 06: MaxConns = 1 membuat uji konkurensi tidak konkuren ====="
Write-Host ""
Write-Host "--- BEFORE: MaxConns = 1 + mutasi T6 (hapus FOR UPDATE) ---"
$tmp = New-TempRepo "b1-bug06"
try {
    Replace-Teks (Join-Path $tmp "tests\integration\main_test.go") "cfg.MaxConns = 10" "cfg.MaxConns = 1"
    Replace-Teks (Join-Path $tmp "app\repository\course_repository.go") `
        'SELECT id FROM users WHERE id = $1 FOR UPDATE' 'SELECT id FROM users WHERE id = $1'
    $res = Invoke-Go $tmp $argsGo
    Write-Host $res.Output
    Write-Host ("exit code = {0}; test merah terdeteksi = {1}" -f $res.Exit, ($res.Output -match "FAIL"))
    Write-Host "GEJALA: mutan T6 LOLOS karena request tidak benar-benar konkuren (MaxConns = 1)."
} finally {
    Remove-TempRepo $tmp
}

Write-Host ""
Write-Host "--- SESUDAH: MaxConns = 10 + mutasi T6 (hapus FOR UPDATE) ---"
$tmp = New-TempRepo "b1-bug06-fix"
try {
    Replace-Teks (Join-Path $tmp "app\repository\course_repository.go") `
        'SELECT id FROM users WHERE id = $1 FOR UPDATE' 'SELECT id FROM users WHERE id = $1'
    $res = Invoke-Go $tmp $argsGo
    Write-Host $res.Output
    Write-Host ("exit code = {0}; test merah terdeteksi = {1} (HARUS: True)" -f $res.Exit, ($res.Output -match "FAIL"))
    Write-Host "PERBAIKAN: dengan MaxConns = 10, mutan T6 TERDAKAP test konkurensi."
} finally {
    Remove-TempRepo $tmp
}

Write-Host ""
Tampilkan-GitStatus