package helper

import (
	"encoding/csv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Format adalah format response yang didukung content negotiation.
type Format string

const (
	FormatJSON Format = "application/json"
	FormatCSV  Format = "text/csv"
)

// Negotiate menentukan format response berdasarkan header Accept.
// Accept kosong atau "*/*" menghasilkan default (format pertama
// dari offered). Bila tidak ada yang cocok, mengembalikan
// *AppError NotAcceptable.
//
// Pemakaian:
//
//	if err := helper.Negotiate(c, helper.FormatJSON, helper.FormatCSV); err != nil {
//	    return err
//	}
//
// Harus dipanggil SEBELUM query DB, supaya format yang tidak
// didukung tidak membuang-buang resource database.
func Negotiate(c *fiber.Ctx, offered ...Format) error {
	if len(offered) == 0 {
		offered = []Format{FormatJSON}
	}
	accept := c.Get("Accept")
	if accept == "" || accept == "*/*" {
		c.Locals("format", offered[0])
		return nil
	}
	for _, want := range strings.Split(accept, ",") {
		want = strings.TrimSpace(want)
		for _, f := range offered {
			if mimeMatches(want, string(f)) {
				c.Locals("format", f)
				return nil
			}
		}
	}
	return NotAcceptable("format tidak didukung: " + accept)
}

// ChosenFormat membaca format yang disimpan oleh Negotiate.
// Default JSON bila tidak ada format tersimpan (Negotiate tak dipanggil).
func ChosenFormat(c *fiber.Ctx) Format {
	f, _ := c.Locals("format").(Format)
	if f == "" {
		return FormatJSON
	}
	return f
}

// mimeMatches mencocokkan satu kandidat dari Accept dengan satu offer.
// Aturan:
//   - persis sama        -> cocok
//   - "*/*"              -> cocok dengan semua
//   - "type/*"           -> cocok dengan semua subtype dari type itu
func mimeMatches(want, offer string) bool {
	if want == offer || want == "*/*" {
		return true
	}
	if strings.HasSuffix(want, "/*") {
		wType := strings.TrimSuffix(want, "/*")
		oParts := strings.SplitN(offer, "/", 2)
		if len(oParts) == 2 && oParts[0] == wType {
			return true
		}
	}
	return false
}

// WriteCSV menulis response CSV dengan header dan baris yang diberikan.
// Wajib Flush() SEBELUM cek Error() (Modul 7 bug #9): tanpa Flush(),
// csv.Writer.Error() tidak mendeteksi kegagalan tulis dan data
// terpotong sebagian sampai ke klien.
func WriteCSV(c *fiber.Ctx, headers []string, rows [][]string) error {
	c.Set("Content-Type", "text/csv")
	w := csv.NewWriter(c.Response().BodyWriter())

	if err := w.Write(headers); err != nil {
		return Internal(err)
	}
	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return Internal(err)
		}
	}

	w.Flush() // WAJIB sebelum writer.Error().
	if err := w.Error(); err != nil {
		return Internal(err)
	}
	return nil
}
