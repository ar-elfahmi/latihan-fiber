package service

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

// UserService adalah service untuk entity users.
// Repository disimpan lewat interface repository.UserRepository
// (lihat app/repository/interfaces.go) sehingga bisa diganti fake di test.
// Saat ini baru berisi List; method lain (Create/Replace/Patch/Delete)
// ditambahkan saat register/CRUD dibutuhkan.
type UserService struct {
	Repo repository.UserRepository
}

// userCSVHeaders adalah urutan kolom pada response CSV.
var userCSVHeaders = []string{"id", "username", "email", "is_active", "created_at"}

// List menampilkan daftar user dengan cursor pagination dan
// content negotiation (JSON default atau CSV bila Accept meminta).
//
// Alur:
//  1. Negotiate format SEBELUM query DB; gagal -> 406.
//  2. Decode cursor; gagal -> 400 INVALID_CURSOR.
//  3. Query repo (List atau FindAfterCursor).
//  4. Potong limit+1, isi meta.
//  5. Render sesuai format yang diminta.
// @Summary		Daftar user
// @Description	Menampilkan daftar user dengan cursor pagination. Format JSON secara
// @Description	default; CSV bila header Accept meminta.
// @Tags			Users
// @Produce		json
// @Param			limit	query		int	false	"Jumlah data per halaman (maks 50)"
// @Param			cursor	query		string	false	"Kursor halaman berikutnya"
// @Success		200	{object}	model.WebResponse
// @Failure		400	{object}	model.ErrorResponse
// @Failure		401	{object}	model.ErrorResponse
// @Failure		403	{object}	model.ErrorResponse
// @Router			/users [get]
func (s UserService) List(c *fiber.Ctx) error {
	if err := helper.Negotiate(c, helper.FormatJSON, helper.FormatCSV); err != nil {
		return err
	}
	format := helper.ChosenFormat(c)

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

	if format == helper.FormatCSV {
		rows := make([][]string, len(users))
		for i, u := range users {
			rows[i] = []string{
				strconv.Itoa(u.ID),
				u.Username,
				u.Email,
				strconv.FormatBool(u.IsActive),
				u.CreatedAt.Format(time.RFC3339),
			}
		}
		return helper.WriteCSV(c, userCSVHeaders, rows)
	}
	return c.JSON(fiber.Map{
		"data": users,
		"meta": meta,
	})
}

// GetByID mengambil detail satu user (GET /users/:id, permission user:read).
// @Summary		Detail satu user
// @Description	Mengambil detail user berdasarkan ID.
// @Tags			Users
// @Produce		json
// @Param			id	path		int	true	"ID user"
// @Success		200	{object}	model.WebResponse
// @Failure		401	{object}	model.ErrorResponse
// @Failure		403	{object}	model.ErrorResponse
// @Failure		404	{object}	model.ErrorResponse
// @Router			/users/{id} [get]
func (s UserService) GetByID(c *fiber.Ctx) error {
	id, err := userID(c)
	if err != nil {
		return err
	}
	u, err := s.Repo.FindByID(c.UserContext(), id)
	if err != nil {
		return translateError(err, "user")
	}
	return c.JSON(model.WebResponse{
		Success: true,
		Message: "detail user berhasil diambil",
		Data:    u,
	})
}

func userID(c *fiber.Ctx) (int, error) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, helper.BadRequest("id user harus berupa angka positif")
	}
	return id, nil
}
