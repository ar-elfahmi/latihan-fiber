package service

import (
	"latihan-fiber/app/model"
	"testing"
)

func TestValidateCreate(t *testing.T) {
	if len(ValidateCreate(model.CreateStudentRequest{})) != 2 {
		t.Fatal("validasi create gagal")
	}
}
func TestValidateUpdate(t *testing.T) {
	if len(ValidateUpdate(model.ReplaceStudentRequest{NIM: "1", Name: "Sari", Grade: 3})) != 0 {
		t.Fatal("data PUT valid ditolak")
	}
}
func TestApplyPatch(t *testing.T) {
	s := model.Student{NIM: "1", Name: "Sari", Grade: 3, IsActive: true}
	v := false
	out, e := ApplyPatch(s, model.PatchStudentRequest{IsActive: &v})
	if len(e) > 0 || out.IsActive {
		t.Fatal("patch gagal")
	}
}
func TestCountTotalPages(t *testing.T) {
	if CountTotalPages(137, 20) != 7 {
		t.Fatal("paginasi gagal")
	}
}
