# BUG 03 (B1): deteksi error SkipRetry memakai "err != asynq.SkipRetry".
# Error dibungkus fmt.Errorf("%w: ..."), jadi perbandingan itu salah dan test
# merah. Perbaikan: !errors.Is(err, asynq.SkipRetry).
$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "bukti-helper.ps1")

Write-Host "===== BUG 03: err != asynq.SkipRetry salah deteksi error terbungkus ====="
$tmp = New-TempRepo "b1-bug03"
try {
    Replace-Teks (Join-Path $tmp "app\service\notification_worker_test.go") `
        "if !errors.Is(err, asynq.SkipRetry) {" "if err != asynq.SkipRetry {" $true
    $res = Invoke-Go $tmp @("test", "-count=1", "-run", "TestNotificationWorkerSkipsRetryOnBrokenTaskPayload", "./app/service/")
    Write-Host $res.Output
    Write-Host ("exit code = {0} (HARUS != 0: test merah)" -f $res.Exit)
} finally {
    Remove-TempRepo $tmp
}
Write-Host ""
Tampilkan-GitStatus