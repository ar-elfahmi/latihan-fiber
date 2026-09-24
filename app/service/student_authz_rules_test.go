package service

import (
	"testing"

	"latihan-fiber/app/model"
	"latihan-fiber/helper"
)

// Uji murni CanAccessStudent: tanpa server, tanpa database, tanpa fiber.
func TestCanAccessStudent(t *testing.T) {
	admin := &helper.PermissionSet{
		"student:read:any":   true,
		"student:update:any": true,
	}
	userSelfOnly := &helper.PermissionSet{
		"student:read:self":   true,
		"student:update:self": true,
	}

	// Pemilik data selalu boleh meski hanya punya permission :self.
	if !CanAccessStudent(model.AuthUser{ID: 3}, 3, userSelfOnly, "student:read:any") {
		t.Error("pemilik data (id 3 = owner 3) seharusnya boleh mengakses")
	}

	// Pemegang :any boleh meski bukan pemilik.
	if !CanAccessStudent(model.AuthUser{ID: 1}, 3, admin, "student:read:any") {
		t.Error("pemegang student:read:any seharusnya boleh mengakses data siapa pun")
	}

	// Bukan pemilik dan tidak punya :any -> tolak.
	if CanAccessStudent(model.AuthUser{ID: 3}, 4, userSelfOnly, "student:read:any") {
		t.Error("user id 3 tidak boleh mengakses data milik id 4")
	}

	// Record tanpa pemilik (owner_id NULL -> 0): hanya :any yang boleh.
	if CanAccessStudent(model.AuthUser{ID: 3}, 0, userSelfOnly, "student:read:any") {
		t.Error("record tanpa pemilik seharusnya tidak bisa diakses pemilik :self manapun")
	}
	if !CanAccessStudent(model.AuthUser{ID: 1}, 0, admin, "student:read:any") {
		t.Error("record tanpa pemilih tetap boleh diakses pemegang :any")
	}

	// perms nil: jalur :any tertutup (fail closed), tapi pemilik tetap
	// boleh karena aturannya "pemilik data selalu boleh".
	if !CanAccessStudent(model.AuthUser{ID: 3}, 3, nil, "student:read:any") {
		t.Error("pemilik dengan perms nil seharusnya tetap boleh (pemilik selalu boleh)")
	}
	if CanAccessStudent(model.AuthUser{ID: 3}, 4, nil, "student:read:any") {
		t.Error("bukan pemilik dengan perms nil harus ditolak (fail closed)")
	}

	// Fail closed: identitas pemanggil tidak valid.
	if CanAccessStudent(model.AuthUser{ID: 0}, 0, userSelfOnly, "student:read:any") {
		t.Error("identitas pemanggil tidak valid harus ditolak")
	}
}
