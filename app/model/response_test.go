package model

import (
	"encoding/json"
	"testing"
)

func TestWebResponseJSON(t *testing.T) {
	body, err := json.Marshal(WebResponse{
		Success: true,
		Data:    map[string]any{"id": 3},
		Message: "ok",
		Meta:    &CursorMeta{Limit: 20, HasMore: true, NextCursor: "abc"},
	})
	if err != nil {
		t.Fatalf("marshal gagal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("hasil bukan JSON valid: %v", err)
	}
	if got["success"] != true {
		t.Errorf("success harus true, dapat %v", got["success"])
	}
	if got["message"] != "ok" {
		t.Errorf("message harus ok, dapat %v", got["message"])
	}
	data, ok := got["data"].(map[string]any)
	if !ok || data["id"] != float64(3) {
		t.Errorf("data harus berisi id=3, dapat %v", got["data"])
	}
	meta, ok := got["meta"].(map[string]any)
	if !ok || meta["has_more"] != true || meta["next_cursor"] != "abc" {
		t.Errorf("meta harus berisi pagination, dapat %v", got["meta"])
	}
}

func TestWebResponseOmitsEmptyFields(t *testing.T) {
	body, err := json.Marshal(WebResponse{Success: true})
	if err != nil {
		t.Fatalf("marshal gagal: %v", err)
	}
	if string(body) != `{"success":true}` {
		t.Errorf("amplop sukses minimal harus {\"success\":true}, dapat %s", string(body))
	}
}

func TestErrorResponseJSON(t *testing.T) {
	body, err := json.Marshal(ErrorResponse{
		Success:   false,
		Code:      "VALIDATION_FAILED",
		Message:   "validasi gagal",
		Fields:    map[string]string{"nama": "wajib diisi"},
		RequestID: "req-1",
	})
	if err != nil {
		t.Fatalf("marshal gagal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("hasil bukan JSON valid: %v", err)
	}
	if got["success"] != false {
		t.Errorf("success harus false, dapat %v", got["success"])
	}
	if got["code"] != "VALIDATION_FAILED" {
		t.Errorf("code harus VALIDATION_FAILED, dapat %v", got["code"])
	}
	if got["message"] != "validasi gagal" {
		t.Errorf("message harus 'validasi gagal', dapat %v", got["message"])
	}
	fields, ok := got["fields"].(map[string]any)
	if !ok || fields["nama"] != "wajib diisi" {
		t.Errorf("fields harus berisi nama, dapat %v", got["fields"])
	}
	if got["request_id"] != "req-1" {
		t.Errorf("request_id harus req-1, dapat %v", got["request_id"])
	}
}

func TestErrorResponseOmitsEmptyFields(t *testing.T) {
	body, err := json.Marshal(ErrorResponse{
		Success:   false,
		Code:      "NOT_FOUND",
		Message:   "mahasiswa tidak ditemukan",
		RequestID: "req-2",
	})
	if err != nil {
		t.Fatalf("marshal gagal: %v", err)
	}
	if string(body) != `{"success":false,"code":"NOT_FOUND","message":"mahasiswa tidak ditemukan","request_id":"req-2"}` {
		t.Errorf("fields kosong harus tidak muncul, dapat %s", string(body))
	}
}
