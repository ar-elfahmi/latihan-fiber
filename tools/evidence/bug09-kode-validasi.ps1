# BUG 09 (B1): kode error validasi ditulis "VALIDATION_ERROR" padahal kode
# stabil repo adalah "VALIDATION_FAILED" (helper.CodeValidationFailed, D-002).
# Perbaikan: ekspektasi test memakai VALIDATION_FAILED.
#
# CATATAN: helper/errors.go sudah memakai VALIDATION_FAILED sebelum slice ini
# (commit 4f5ca65); temuan ini soal ekspektasi test, bukan kode produksi.
$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "bukti-helper.ps1")

Write-Host "===== BUG 09: kode error validasi VALIDATION_ERROR != VALIDATION_FAILED ====="
$tmp = New-TempRepo "b1-bug09"
try {
    Replace-Teks (Join-Path $tmp "app\service\review_errors_test.go") `
        "VALIDATION_FAILED" "VALIDATION_ERROR"
    $res = Invoke-Go $tmp @("test", "-count=1", "-run", "TestReviewValidationCodeIsStable", "./app/service/")
    Write-Host $res.Output
    Write-Host ("exit code = {0} (HARUS != 0: test merah)" -f $res.Exit)
} finally {
    Remove-TempRepo $tmp
}
Write-Host ""
Tampilkan-GitStatus