package helper

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

// ErrInvalidCursor dikembalikan DecodeCursor bila string bukan cursor
// yang sah. Pemakai dapat menerjemahkan ke 400 Bad Request.
var ErrInvalidCursor = errors.New("cursor tidak valid")

// cursorPayload adalah isi cursor yang di-encode JSON lalu base64.
// time disimpan sebagai RFC3339Nano agar presisi sampai nanosecond.
type cursorPayload struct {
	T string `json:"t"` // created_at RFC3339Nano
	I int    `json:"i"` // id baris terakhir
}

// EncodeCursor menghasilkan string base64-RawURL dari (createdAt, id).
// Dipakai service List untuk menandai "setelah baris ini".
func EncodeCursor(createdAt time.Time, id int) string {
	payload, _ := json.Marshal(cursorPayload{
		T: createdAt.UTC().Format(time.RFC3339Nano),
		I: id,
	})
	return base64.RawURLEncoding.EncodeToString(payload)
}

// DecodeCursor mengambil (createdAt, id) dari cursor. Mengembalikan
// ErrInvalidCursor pada tiap jalur gagal: string kosong, base64 gagal,
// JSON gagal, field kosong, atau tipe salah.
func DecodeCursor(s string) (time.Time, int, error) {
	if s == "" {
		return time.Time{}, 0, ErrInvalidCursor
	}
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return time.Time{}, 0, ErrInvalidCursor
	}
	var p cursorPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return time.Time{}, 0, ErrInvalidCursor
	}
	if p.T == "" || p.I == 0 {
		return time.Time{}, 0, ErrInvalidCursor
	}
	t, err := time.Parse(time.RFC3339Nano, p.T)
	if err != nil {
		return time.Time{}, 0, ErrInvalidCursor
	}
	return t, p.I, nil
}
