package service

import (
	"testing"

	"latihan-fiber/app/model"
)

// Perhatikan pengujian ini tidak menyalakan server, tidak menyentuh
// database, dan tidak membuat fiber.Ctx.

func TestValidateCreate(t *testing.T) {
	cases := []struct {
		name string
		req  model.CreateStudentRequest
		want int // jumlah field yang bermasalah
	}{
		{"semua kosong", model.CreateStudentRequest{}, 2},
		{"grade di luar batas", model.CreateStudentRequest{NIM: "434241047", Name: "Alfian", Grade: 4.5}, 1},
		{"data lengkap dan valid", model.CreateStudentRequest{NIM: "434241047", Name: "Alfian", Grade: 3.5}, 0},
	}

	for _, tc := range cases {
		if got := len(ValidateCreate(tc.req)); got != tc.want {
			t.Errorf("%s: harap %d error, dapat %d", tc.name, tc.want, got)
		}
	}
}

func TestValidateUpdate(t *testing.T) {
	cases := []struct {
		name string
		req  model.ReplaceStudentRequest
		want int
	}{
		{"PUT wajib mengisi semua", model.ReplaceStudentRequest{}, 2},
		{"data PUT valid", model.ReplaceStudentRequest{NIM: "434241047", Name: "Alfian", Grade: 3}, 0},
	}

	for _, tc := range cases {
		if got := len(ValidateUpdate(tc.req)); got != tc.want {
			t.Errorf("%s: harap %d error, dapat %d", tc.name, tc.want, got)
		}
	}
}

func TestApplyPatch(t *testing.T) {
	initial := model.Student{ID: 1, NIM: "434241047", Name: "Alfian", Grade: 3.5, IsActive: true}
	inactive := false

	result, errs := ApplyPatch(initial, model.PatchStudentRequest{IsActive: &inactive})

	if len(errs) != 0 {
		t.Fatalf("tidak seharusnya ada error: %v", errs)
	}

	if result.IsActive {
		t.Error("is_active seharusnya berubah menjadi false")
	}

	if result.Name != "Alfian" {
		t.Error("field yang tidak dikirim seharusnya tidak berubah")
	}
}

func TestIsEmptyPatch(t *testing.T) {
	if !IsEmptyPatch(model.PatchStudentRequest{}) {
		t.Error("request tanpa field seharusnya dianggap kosong")
	}

	grade := 4.0
	if IsEmptyPatch(model.PatchStudentRequest{Grade: &grade}) {
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
