package service

import (
	"testing"

	"latihan-fiber/app/model"
)

// TestSummarizeReviewsEmpty memastikan kelas tanpa ulasan menghasilkan
// ringkasan kosong (bukan NaN dan bukan map nil).
func TestSummarizeReviewsEmpty(t *testing.T) {
	got := SummarizeReviews(7, nil)

	if got.CourseID != 7 {
		t.Errorf("course_id: dapat %d, harus 7", got.CourseID)
	}
	if got.TotalReviews != 0 {
		t.Errorf("total_reviews: dapat %d, harus 0", got.TotalReviews)
	}
	if got.AverageRating != 0 {
		t.Errorf("average_rating: dapat %v, harus 0", got.AverageRating)
	}
	for rating := 1; rating <= 5; rating++ {
		key := string(rune('0' + rating))
		if got.Distribution[key] != 0 {
			t.Errorf("distribusi[%s]: dapat %d, harus 0", key, got.Distribution[key])
		}
	}
}

// TestSummarizeReviewsAverage memastikan rata-rata rating dihitung dari
// seluruh ulasan dan dibulatkan dua desimal.
func TestSummarizeReviewsAverage(t *testing.T) {
	reviews := []model.Review{
		{CourseID: 1, Rating: 4},
		{CourseID: 1, Rating: 4},
		{CourseID: 1, Rating: 5},
	}

	got := SummarizeReviews(1, reviews)

	if got.TotalReviews != 3 {
		t.Errorf("total_reviews: dapat %d, harus 3", got.TotalReviews)
	}
	if got.AverageRating != 4.33 {
		t.Errorf("average_rating: dapat %v, harus 4.33", got.AverageRating)
	}
	if got.Distribution["4"] != 2 || got.Distribution["5"] != 1 {
		t.Errorf("distribusi: dapat %v, harus 4=>2 dan 5=>1", got.Distribution)
	}
	if got.Distribution["1"] != 0 {
		t.Errorf("distribusi[1]: dapat %d, harus 0", got.Distribution["1"])
	}
}

// TestSummarizeReviewsIgnoresInvalidRating memastikan rating di luar
// rentang 1..5 tidak ikut dihitung ke rata-rata, tetapi ulasan tetap dihitung.
func TestSummarizeReviewsIgnoresInvalidRating(t *testing.T) {
	reviews := []model.Review{
		{CourseID: 1, Rating: 0},
		{CourseID: 1, Rating: 5},
	}

	got := SummarizeReviews(1, reviews)

	if got.TotalReviews != 2 {
		t.Errorf("total_reviews: dapat %d, harus 2", got.TotalReviews)
	}
	if got.AverageRating != 2.5 {
		t.Errorf("average_rating: dapat %v, harus 2.5", got.AverageRating)
	}
}
