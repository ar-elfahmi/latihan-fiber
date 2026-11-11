package service

import (
	"math"
	"strconv"

	"latihan-fiber/app/model"
)

// SummarizeReviews menghitung ringkasan ulasan satu kelas: jumlah ulasan,
// rata-rata rating (dibulatkan 2 desimal), dan distribusi rating 1..5.
//
// Aturan murni (tanpa Fiber/MongoDB) supaya bisa diuji dengan go test
// tanpa layanan. Rating di luar rentang 1..5 tidak masuk rata-rata tetapi
// tetap dihitung sebagai ulasan.
func SummarizeReviews(courseID int, reviews []model.Review) model.ReviewSummary {
	distribution := map[string]int{"1": 0, "2": 0, "3": 0, "4": 0, "5": 0}
	total := 0
	sum := 0
	for _, r := range reviews {
		total++
		if r.Rating >= 1 && r.Rating <= 5 {
			distribution[strconv.Itoa(r.Rating)]++
			sum += r.Rating
		}
	}

	average := 0.0
	if total > 0 {
		average = math.Round(float64(sum)/float64(total)*100) / 100
	}

	return model.ReviewSummary{
		CourseID:      courseID,
		TotalReviews:  total,
		AverageRating: average,
		Distribution:  distribution,
	}
}
