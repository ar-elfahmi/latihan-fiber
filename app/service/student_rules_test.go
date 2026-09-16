package service

import (
	"testing"

	"latihan-fiber/app/model"
)

// Pengujian di file ini TIDAK memerlukan server, database, atau fiber.Ctx.
// Cukup panggil fungsi, cocokkan hasilnya. Inilah gunanya file rules
// dipisah dari handler: bisa dites terisolasi.

func TestValidateCreate(t *testing.T) {
	cases := []struct {
		name string
		req  model.CreateStudentRequest
		want int
	}{
		{"semua kosong", model.CreateStudentRequest{}, 2},
		{"grade di luar batas", model.CreateStudentRequest{NIM: "434241047", Name: "Alfian", Grade: 4.5}, 1},
		{"data valid", model.CreateStudentRequest{NIM: "434241047", Name: "Alfian", Grade: 3.5}, 0},
	}
	for _, tc := range cases {
		if got := len(ValidateCreate(tc.req)); got != tc.want {
			t.Errorf("%s: harap %d error, dapat %d", tc.name, tc.want, got)
		}
	}
}

func TestValidateReplace(t *testing.T) {
	cases := []struct {
		name string
		req  model.ReplaceStudentRequest
		want int
	}{
		{"PUT wajib isi semua", model.ReplaceStudentRequest{}, 2},
		{"PUT valid", model.ReplaceStudentRequest{NIM: "434241047", Name: "Alfian", Grade: 3}, 0},
	}
	for _, tc := range cases {
		if got := len(ValidateReplace(tc.req)); got != tc.want {
			t.Errorf("%s: harap %d error, dapat %d", tc.name, tc.want, got)
		}
	}
}

func TestApplyPatch(t *testing.T) {
	initial := model.Student{ID: 1, NIM: "434241047", Name: "Alfian", Grade: 3.5, IsActive: true}
	inactive := false
	updated, errs := ApplyPatch(initial, model.PatchStudentRequest{IsActive: &inactive})
	if len(errs) != 0 {
		t.Fatalf("tidak seharusnya ada error: %v", errs)
	}
	if updated.IsActive {
		t.Error("is_active seharusnya berubah menjadi false")
	}
	if updated.Name != "Alfian" {
		t.Error("field yang tidak dikirim seharusnya tidak berubah")
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
