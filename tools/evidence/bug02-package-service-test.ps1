# BUG 02 (B1): file test memakai package service_test -> tidak bisa mengakses
# translateReviewError yang tidak diekspor. Gejala: compile error.
# Perbaikan: package service.
$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "bukti-helper.ps1")

Write-Host "===== BUG 02: package service_test tak bisa akses translateReviewError ====="
$tmp = New-TempRepo "b1-bug02"
try {
    Replace-Teks (Join-Path $tmp "app\service\review_errors_test.go") `
        "package service" "package service_test" $true
    $res = Invoke-Go $tmp @("test", "-count=1", "./app/service/")
    Write-Host $res.Output
    Write-Host ("exit code = {0} (HARUS != 0: compile error / test merah)" -f $res.Exit)
} finally {
    Remove-TempRepo $tmp
}
Write-Host ""
Tampilkan-GitStatus