# BUG 05 (B1): "if dsn != """ terbalik -> database uji tidak pernah dipakai,
# integration test dilewati (0 baris --- PASS). Perbaikan: "if dsn == """ ->
# test gagal dengan pesan jelas kalau TEST_DATABASE_URL kosong.
$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "bukti-helper.ps1")
Set-TestDatabaseUrl

Write-Host "===== BUG 05: if dsn != (terbalik) -> integration test dilewati ====="
$tmp = New-TempRepo "b1-bug05"
try {
    $file = Join-Path $tmp "tests\integration\main_test.go"
    Replace-Teks $file '	if dsn == "" {' '	if dsn != "" {'
    Replace-Teks $file "TEST_DATABASE_URL tidak diisi: integration test tidak bisa dijalankan" "TEST_DATABASE_URL tidak diisi: integration test dilewati"
    $res = Invoke-Go $tmp @("test", "-tags=integration", "-count=1", "-v", "./tests/integration/")
    Write-Host $res.Output
    Write-Host ("exit code = {0}; jumlah baris '--- PASS' = {1} (HARUS > 0 untuk benar)" -f $res.Exit, ([regex]::Matches($res.Output, "(?m)^--- PASS")).Count)
} finally {
    Remove-TempRepo $tmp
}
Write-Host ""
Tampilkan-GitStatus