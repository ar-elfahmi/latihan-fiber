# BUG 04 (B1): fake repository disimpan di variabel GLOBAL (sharedReviews).
# State fake bocor antar test. Perbaikan: fake dibuat per test lewat
# newFakeReviewRepo().
$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "bukti-helper.ps1")

Write-Host "===== BUG 04: fake repository global bocor antar test ====="
$tmp = New-TempRepo "b1-bug04"
try {
    $file = Join-Path $tmp "app\service\review_service_test.go"
    $global = @"
// sharedReviews adalah fake GLOBAL: state-nya dipakai semua test.
var sharedReviews = func() *fakeReviewRepo {
    f := &fakeReviewRepo{}
    f.reviews = append(f.reviews, model.Review{ID: "sisa", CourseID: 1, UserID: 7, Rating: 5, Comment: "ulasan sisa test lain", Tags: []string{"sisa"}})
    return f
}()

"@
    Replace-Teks $file "func newFakeReviewRepo() *fakeReviewRepo {" ($global + "func newFakeReviewRepo() *fakeReviewRepo {")
    Replace-Teks $file "	return &fakeReviewRepo{}" "	return sharedReviews"
    $res = Invoke-Go $tmp @("test", "-count=1", "-run", "TestReviewServiceFakeIsolatedPerTest", "./app/service/")
    Write-Host $res.Output
    Write-Host ("exit code = {0} (HARUS != 0: test merah)" -f $res.Exit)
} finally {
    Remove-TempRepo $tmp
}
Write-Host ""
Tampilkan-GitStatus