# BUG 07 (B1): seedCourse memakai defer di dalam helper -> data terhapus
# sebelum test memakainya (bocor antar test / data hilang).
# Perbaikan: t.Cleanup.
$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "bukti-helper.ps1")
Set-TestDatabaseUrl

Write-Host "===== BUG 07: seedCourse pakai defer, bukan t.Cleanup ====="
$tmp = New-TempRepo "b1-bug07"
try {
    $file = Join-Path $tmp "tests\integration\main_test.go"
    Replace-Teks $file "	t.Cleanup(func() { cleanupCourse(t, course.ID) })" "	defer cleanupCourse(t, course.ID)"
    Replace-Teks $file "	t.Cleanup(func() { cleanupStudent(t, user.ID) })" "	defer cleanupStudent(t, user.ID)"
    $res = Invoke-Go $tmp @("test", "-tags=integration", "-count=1", "-run", "TestIncrementEnrolled_StopsAtQuota", "./tests/integration/")
    Write-Host $res.Output
    Write-Host ("exit code = {0} (HARUS != 0: test merah)" -f $res.Exit)
} finally {
    Remove-TempRepo $tmp
}
Write-Host ""
Tampilkan-GitStatus