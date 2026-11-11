# BUG 08 (B1): tidak ada skenario batas 24 SKS -> mutasi "<=" menjadi "<"
# lolos (test hijau padahal aturan rusak).
#
# CATATAN: skenario batas 24 SKS SUDAH ADA di repo sejak S-003
# (commit cadb1d3 "test: cover credit limit and tag normalization rules").
# Skrip ini membuktikannya: mutasi "<=" -> "<" tetap membuat test merah.
$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "bukti-helper.ps1")

Write-Host "===== BUG 08: skenario batas 24 SKS (sudah ada sebelum slice ini) ====="
Write-Host ""
Write-Host "--- riwayat skenario batas di app/service/course_rules_test.go ---"
Write-Host (& git -C $script:RepoRoot log --oneline -- app/service/course_rules_test.go | Out-String)
Write-Host "--- mutasi T5: <= menjadi < ---"
$tmp = New-TempRepo "b1-bug08"
try {
    Replace-Teks (Join-Path $tmp "app\service\course_rules.go") `
        "return current+delta <= MaxCreditsPerSemester" "return current+delta < MaxCreditsPerSemester"
    $res = Invoke-Go $tmp @("test", "-count=1", "-run", "TestWithinCreditLimit", "./app/service/")
    Write-Host $res.Output
    Write-Host ("exit code = {0}; test merah terdeteksi = {1} (HARUS: True -> skenario batas menangkap mutan)" -f $res.Exit, ($res.Output -match "FAIL"))
} finally {
    Remove-TempRepo $tmp
}
Write-Host ""
Tampilkan-GitStatus