package model

// Bentuk respons tunggal aplikasi.
//
// - WebResponse  : amplop untuk respons sukses.
// - ErrorResponse: amplop untuk respons error (dipakai config.ErrorHandler).
//
// Kedua amplop ini adalah SATU-SATUNYA bentuk JSON yang boleh dipakai
// handler baru. Endpoint lama (yang sudah ada sebelum S-001) tetap memakai
// bentuk lamanya supaya kontrak tidak berubah — lihat docs/API.md.
//
// Catatan:omitempty dipakai supaya amplop error identik dengan JSON
// sebelumnya: "fields" hanya muncul bila ada isinya.

// WebResponse adalah amplop respons sukses.
//
// Success selalu diisi true; Data dipakai untuk payload; Message untuk
// keterangan singkat; Meta untuk metadata pagination (CursorMeta).
type WebResponse struct {
	Success bool        `json:"success"`
	Data    any         `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Meta    *CursorMeta `json:"meta,omitempty"`
}

// ErrorResponse adalah amplop respons error.
//
// Bentuknya kontrak publik (D-002):
//
//	{
//	  "success": false,
//	  "code": "VALIDATION_FAILED",
//	  "message": "...",
//	  "fields": {"name": "wajib diisi"},
//	  "request_id": "..."
//	}
type ErrorResponse struct {
	Success   bool              `json:"success"`
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
	RequestID string            `json:"request_id"`
}
