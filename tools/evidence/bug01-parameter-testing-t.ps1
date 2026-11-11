# BUG 01 (B1): parameter test "t testing.T" (nilai, bukan pointer).
# Gejala: paket app/service gagal dikompilasi -> seluruh test package merah.
# Perbaikan: "t *testing.T".
$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "bukti-helper.ps1")

Write-Host "===== BUG 01: parameter test t testing.T (bukan *testing.T) ====="
$tmp = New-TempRepo "b1-bug01"
try {
    Replace-Teks (Join-Path $tmp "app\service\review_rules_test.go") `
        "func TestReviewRatingRules(t *testing.T) {" "func TestReviewRatingRules(t testing.T) {"
    $res = Invoke-Go $tmp @("test", "-count=1", "./app/service/")
    Write-Host $res.Output
    Write-Host ("exit code = {0} (HARUS != 0: compile error / test merah)" -f $res.Exit)
} finally {
    Remove-TempRepo $tmp
}
Write-Host ""
Tampilkan-GitStatus