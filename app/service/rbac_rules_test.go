package service

import "testing"

func TestHasPermission(t *testing.T) {
	perms := map[string]bool{
		"student:list":       true,
		"student:read:any":   true,
		"student:update:any": true,
	}
	if !HasPermission(perms, "student:read:any") {
		t.Error("admin harusnya punya student:read:any")
	}
	if HasPermission(perms, "student:delete") {
		t.Error("admin seharusnya TIDAK punya student:delete di test ini")
	}
	if HasPermission(nil, "student:read:any") {
		t.Error("perms nil harusnya gagal (fail closed)")
	}
}

func TestHasAnyPermissionFromGroups(t *testing.T) {
	permsStudentOnly := map[string]bool{"student:read:self": true}
	groups := []string{"student:read:any", "student:read:self"}
	if !HasAnyPermissionFromGroups(permsStudentOnly, groups) {
		t.Error("pemilik harusnya tetap boleh lewat scope :self")
	}
	if HasAnyPermissionFromGroups(nil, groups) {
		t.Error("nil harusnya ditolak")
	}
	if HasAnyPermissionFromGroups(permsStudentOnly, []string{"student:delete"}) {
		t.Error("tidak ada kandidat yang cocok harusnya ditolak")
	}
}

func TestRequireOwnerOrAdmin(t *testing.T) {
	if !RequireOwnerOrAdmin(7, 7) {
		t.Error("pemilik sah harusnya boleh")
	}
	if RequireOwnerOrAdmin(7, 8) {
		t.Error("orang lain harusnya ditolak")
	}
	if RequireOwnerOrAdmin(0, 7) {
		t.Error("actor tanpa identitas harusnya ditolak (fail closed)")
	}
	if RequireOwnerOrAdmin(7, 0) {
		t.Error("record tanpa owner harusnya ditolak di service; admin bypass di middleware")
	}
}

func TestIsAdminOnlyPermission(t *testing.T) {
	staff := map[string]bool{"student:create": true}
	if !IsAdminOnlyPermission(staff, "student:create") {
		t.Error("staff dengan student:create harusnya lewat cek")
	}
	student := map[string]bool{"student:read:self": true}
	if IsAdminOnlyPermission(student, "student:delete") {
		t.Error("student tanpa student:delete harusnya ditolak")
	}
	if IsAdminOnlyPermission(nil, "student:create") {
		t.Error("nil harusnya false")
	}
}
