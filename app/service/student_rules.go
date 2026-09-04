package service

import (
	"latihan-fiber/app/model"
	"strings"
)

func ValidateCreate(r model.CreateStudentRequest) map[string]string {
	e := map[string]string{}
	r.NIM = strings.TrimSpace(r.NIM)
	r.Name = strings.TrimSpace(r.Name)
	if r.NIM == "" {
		e["nim"] = "wajib diisi"
	}
	if r.Name == "" {
		e["name"] = "wajib diisi"
	}
	if r.Grade < 0 || r.Grade > 4 {
		e["grade"] = "harus bernilai antara 0.0 - 4.0"
	}
	return e
}
func ValidateUpdate(r model.ReplaceStudentRequest) map[string]string {
	e := map[string]string{}
	if strings.TrimSpace(r.NIM) == "" {
		e["nim"] = "wajib diisi pada PUT"
	}
	if strings.TrimSpace(r.Name) == "" {
		e["name"] = "wajib diisi pada PUT"
	}
	if r.Grade < 0 || r.Grade > 4 {
		e["grade"] = "harus bernilai antara 0.0 - 4.0"
	}
	return e
}
func ApplyPatch(s model.Student, r model.PatchStudentRequest) (model.Student, map[string]string) {
	e := map[string]string{}
	if r.NIM != nil {
		v := strings.TrimSpace(*r.NIM)
		if v == "" {
			e["nim"] = "tidak boleh kosong"
		} else {
			s.NIM = v
		}
	}
	if r.Name != nil {
		v := strings.TrimSpace(*r.Name)
		if v == "" {
			e["name"] = "tidak boleh kosong"
		} else {
			s.Name = v
		}
	}
	if r.Grade != nil {
		if *r.Grade < 0 || *r.Grade > 4 {
			e["grade"] = "harus bernilai antara 0.0 - 4.0"
		} else {
			s.Grade = *r.Grade
		}
	}
	if r.IsActive != nil {
		s.IsActive = *r.IsActive
	}
	return s, e
}
func CountTotalPages(total, limit int) int {
	if total == 0 {
		return 0
	}
	return (total + limit - 1) / limit
}
