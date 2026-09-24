package service

import (
	"github.com/gofiber/fiber/v2"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

// UserService adalah struct konkret untuk entity users.
// Saat ini baru berisi List; method lain (Create/Replace/Patch/Delete)
// ditambahkan saat register/CRUD dibutuhkan.
type UserService struct {
	Repo repository.UserRepository
}

// List menampilkan daftar user dengan cursor pagination.
//
// Alur:
//  1. Parse limit dari query (default 10, max 50).
//  2. Bila tidak ada cursor, ambil halaman pertama dari Repo.List.
//  3. Bila ada cursor, decode via helper.DecodeCursor; gagal
//     decode -> helper.InvalidCursor (400).
//  4. Potong limit+1: bila kelebihan 1 baris, has_more = true dan
//     next_cursor diisi dari baris terakhir.
func (s UserService) List(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	cursorStr := c.Query("cursor")

	var (
		users []model.User
		err   error
	)
	if cursorStr == "" {
		users, err = s.Repo.List(c.Context(), limit)
	} else {
		cursorTime, cursorID, derr := helper.DecodeCursor(cursorStr)
		if derr != nil {
			return helper.InvalidCursor()
		}
		users, err = s.Repo.FindAfterCursor(c.Context(), cursorTime, cursorID, limit)
	}
	if err != nil {
		return translateError(err, "user")
	}

	hasMore := len(users) > limit
	if hasMore {
		users = users[:limit]
	}

	meta := model.CursorMeta{
		Limit:   limit,
		HasMore: hasMore,
	}
	if hasMore && len(users) > 0 {
		last := users[len(users)-1]
		meta.NextCursor = helper.EncodeCursor(last.CreatedAt, last.ID)
	}

	return c.JSON(fiber.Map{
		"data": users,
		"meta": meta,
	})
}
