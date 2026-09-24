package service

import (
	"testing"

	"latihan-fiber/app/model"
)

// Pengujian di file ini TIDAK memerlukan server, database, atau fiber.Ctx.
// Cukup panggil fungsi, cocokkan hasilnya. Inilah gunanya file rules
// dipisah dari handler: bisa dites terisolasi.

func TestApplyPatch(t *testing.T) {
	initial := model.Student{ID: 1, NIM: "434241047", Name: "Alfian", Grade: 3.5, IsActive: true}
	inactive := false
	updated := ApplyPatch(initial, model.PatchStudentRequest{IsActive: &inactive})
	if updated.IsActive {
		t.Error("is_active seharusnya berubah menjadi false")
	}
	if updated.Name != "Alfian" {
		t.Error("field yang tidak dikirim seharusnya tidak berubah")
	}
	if updated.NIM != "434241047" {
		t.Error("nim yang tidak dikirim seharusnya tidak berubah")
	}
}

func TestIsEmptyPatch(t *testing.T) {
	if !IsEmptyPatch(model.PatchStudentRequest{}) {
		t.Error("request tanpa field seharusnya dianggap kosong")
	}
	g := 4.0
	if IsEmptyPatch(model.PatchStudentRequest{Grade: &g}) {
		t.Error("request berisi grade seharusnya tidak dianggap kosong")
	}
}

func TestCountTotalPages(t *testing.T) {
	cases := []struct{ total, limit, want int }{
		{0, 10, 0},
		{1, 10, 1},
		{10, 10, 1},
		{11, 10, 2},
		{137, 20, 7},
	}
	for _, tc := range cases {
		if got := CountTotalPages(tc.total, tc.limit); got != tc.want {
			t.Errorf("total=%d limit=%d: harap %d, dapat %d",
				tc.total, tc.limit, tc.want, got)
		}
	}
}
