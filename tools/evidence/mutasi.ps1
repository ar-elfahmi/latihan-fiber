# Uji mutasi T5, T6, T7 (Modul 12 C.1).
#
# Untuk tiap mutasi: salin repo ke %TEMP%, suntik mutasi ke kode production,
# jalankan perintah, tampilkan output (test WAJIB merah), lalu hapus salinan.
# Working tree asli tidak pernah berubah.
#
# Pakai (Windows PowerShell):
#   powershell -ExecutionPolicy Bypass -File tools\evidence\mutasi.ps1

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "bukti-helper.ps1")
Set-TestDatabaseUrl

function Jalankan-Mutasi($no, $nama, $file, $old, $new, $argsGo, $testHarusMerah) {
    Write-Host ""
    Write-Host ("===== MUTASI {0}: {1} =====" -f $no, $nama)
    Write-Host ("perintah: go {0}" -f ($argsGo -join " "))
    $tmp = New-TempRepo ("b1-mutasi-" + $no)
    try {
        Replace-Teks (Join-Path $tmp $file) $old $new
        Write-Host ("mutasi diterapkan: {0}" -f $file)
        $res = Invoke-Go $tmp $argsGo
        Write-Host $res.Output
        $merah = ($res.Output -match "FAIL")
        Write-Host ("exit code = {0}; test merah terdeteksi = {1}; test yang harus merah = {2}" -f $res.Exit, $merah, $testHarusMerah)
    } finally {
        Remove-TempRepo $tmp
    }
}

Jalankan-Mutasi "T5" "WithinCreditLimit: <= menjadi <" "app\service\course_rules.go" `
    "return current+delta <= MaxCreditsPerSemester" "return current+delta < MaxCreditsPerSemester" `
    @("test", "-count=1", "./app/service/") "TestWithinCreditLimit"

Jalankan-Mutasi "T6" "LockStudent (lockKRS): hapus FOR UPDATE pada baris user" "app\repository\course_repository.go" `
    'SELECT id FROM users WHERE id = $1 FOR UPDATE' 'SELECT id FROM users WHERE id = $1' `
    @("test", "-tags=integration", "-count=3", "-run", "Concurrent", "./tests/integration/") "TestEnrollEndpoint_ConcurrentCreditLimit (3x)"

Jalankan-Mutasi "T7" "Claim: hapus ON CONFLICT (outbox_id, user_id) DO NOTHING" "app\repository\notification_repository.go" `
    'VALUES ($1, $2, $3, $4, $5) ON CONFLICT (outbox_id, user_id) DO NOTHING' 'VALUES ($1, $2, $3, $4, $5)' `
    @("test", "-tags=integration", "-count=1", "-run", "Claim", "./tests/integration/") "TestClaim_OnlyOnce"

Write-Host ""
Tampilkan-GitStatus