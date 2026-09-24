package model

// CursorMeta adalah metadata pagination cursor untuk ListResponse.
// Tanpa Total/TotalPages: cursor pagination tidak menjamin jumlah
// total baris karena OFFSET tidak dipakai. Client cukup mengikuti
// next_cursor hingga has_more=false.
type CursorMeta struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}
