package helper

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"latihan-fiber/app/model"
	"strconv"
	"strings"
	"time"
)

func OK(c *fiber.Ctx, msg string, data any) error {
	return c.Status(200).JSON(model.WebResponse{Success: true, Message: msg, Data: data})
}
func OKList(c *fiber.Ctx, msg string, data any, meta *model.Meta) error {
	return c.Status(200).JSON(model.WebResponse{Success: true, Message: msg, Data: data, Meta: meta})
}
func Created(c *fiber.Ctx, msg string, data any, location string) error {
	c.Set("Location", location)
	return c.Status(201).JSON(model.WebResponse{Success: true, Message: msg, Data: data})
}
func NoContent(c *fiber.Ctx) error { return c.SendStatus(204) }
func Fail(c *fiber.Ctx, status int, msg string) error {
	return c.Status(status).JSON(model.WebResponse{Success: false, Message: msg})
}
func FailValidation(c *fiber.Ctx, e map[string]string) error {
	return c.Status(422).JSON(model.WebResponse{Success: false, Message: "validasi gagal", Errors: e})
}
func RequestContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}
func ParamID(c *fiber.Ctx) (int, bool) {
	id, e := strconv.Atoi(c.Params("id"))
	return id, e == nil && id > 0
}
func ParseListQuery(c *fiber.Ctx) model.ListQuery {
	q := model.ListQuery{Page: c.QueryInt("page", 1), Limit: c.QueryInt("limit", 10), Search: strings.TrimSpace(c.Query("search")), Sort: c.Query("sort", "id"), Order: strings.ToLower(c.Query("order", "asc"))}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 50 {
		q.Limit = 10
	}
	allowed := map[string]bool{"id": true, "nim": true, "name": true, "grade": true, "created_at": true}
	if !allowed[q.Sort] {
		q.Sort = "id"
	}
	if q.Order != "desc" {
		q.Order = "asc"
	}
	if raw := c.Query("is_active"); raw != "" {
		if v, e := strconv.ParseBool(raw); e == nil {
			q.IsActive = &v
		}
	}
	return q
}
