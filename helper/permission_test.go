package helper

import (
	"reflect"
	"testing"
)

// TestNewPermissionSetKeepsPermissionList mengunci daftar permission per
// aksi. Nilai string ini dibandingkan dengan isi tabel permissions, jadi
// mengubahnya = mengubah kontrak (D-002/AGENTS.md §5).
func TestNewPermissionSetKeepsPermissionList(t *testing.T) {
	perms := NewPermissionSet()

	cases := []struct {
		aksi string
		got  []string
		want []string
	}{
		{aksi: "CourseEnroll", got: perms.CourseEnroll, want: []string{"course:enroll"}},
		{aksi: "CourseManage", got: perms.CourseManage, want: []string{"course:manage"}},
		{aksi: "QueueMonitor", got: perms.QueueMonitor, want: []string{"queue:monitor"}},
		{aksi: "PrestasiList", got: perms.PrestasiList, want: []string{"prestasi:list"}},
		{aksi: "PrestasiRead", got: perms.PrestasiRead, want: []string{"prestasi:read:any", "prestasi:read:self"}},
		{aksi: "PrestasiCreate", got: perms.PrestasiCreate, want: []string{"prestasi:create"}},
		{aksi: "StudentList", got: perms.StudentList, want: []string{"student:list"}},
		{aksi: "StudentRead", got: perms.StudentRead, want: []string{"student:read:any", "student:read:self"}},
		{aksi: "StudentCreate", got: perms.StudentCreate, want: []string{"student:create"}},
		{aksi: "StudentUpdate", got: perms.StudentUpdate, want: []string{"student:update:any", "student:update:self"}},
		{aksi: "StudentDelete", got: perms.StudentDelete, want: []string{"student:delete"}},
		{aksi: "UserList", got: perms.UserList, want: []string{"user:list"}},
		{aksi: "UserRead", got: perms.UserRead, want: []string{"user:read"}},
		{aksi: "ReviewModerate", got: perms.ReviewModerate, want: []string{"review:moderate"}},
	}

	for _, tc := range cases {
		if !reflect.DeepEqual(tc.got, tc.want) {
			t.Errorf("permission %s harus %v, dapat %v", tc.aksi, tc.want, tc.got)
		}
	}
}

// TestNewPermissionSetHasNoDuplicates memastikan tidak ada permission yang
// terulang di dalam satu kelompok.
func TestNewPermissionSetHasNoDuplicates(t *testing.T) {
	perms := NewPermissionSet()
	groups := map[string][]string{
		"CourseEnroll":   perms.CourseEnroll,
		"CourseManage":   perms.CourseManage,
		"QueueMonitor":   perms.QueueMonitor,
		"PrestasiList":   perms.PrestasiList,
		"PrestasiRead":   perms.PrestasiRead,
		"PrestasiCreate": perms.PrestasiCreate,
		"StudentList":    perms.StudentList,
		"StudentRead":    perms.StudentRead,
		"StudentCreate":  perms.StudentCreate,
		"StudentUpdate":  perms.StudentUpdate,
		"StudentDelete":  perms.StudentDelete,
		"UserList":       perms.UserList,
		"UserRead":       perms.UserRead,
		"ReviewModerate": perms.ReviewModerate,
	}

	for aksi, group := range groups {
		if len(group) == 0 {
			t.Errorf("kelompok %s harus punya minimal satu permission", aksi)
		}
		seen := map[string]bool{}
		for _, name := range group {
			if seen[name] {
				t.Errorf("permission %s terulang di kelompok %s", name, aksi)
			}
			seen[name] = true
		}
	}
}
