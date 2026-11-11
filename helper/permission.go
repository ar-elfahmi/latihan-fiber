package helper

// PermissionSet adalah himpunan permission milik satu user.
// Bentuknya tetap map[string]bool (lookup O(1)), dibungkus tipe sendiri
// supaya punya method Has dengan kebijakan fail closed.
//
// File ini sengaja tidak mengimpor fiber: dipakai oleh aturan murni
// di app/service/student_authz_rules.go agar tetap bisa diuji go test.
type PermissionSet map[string]bool

// Has melapor apakah permission tersebut dimiliki user.
// Receiver nil atau set nil mengembalikan false (fail closed):
// "hak tidak diketahui" dianggap "tidak boleh".
func (p *PermissionSet) Has(name string) bool {
	if p == nil || *p == nil {
		return false
	}
	return (*p)[name]
}

// PermissionActions adalah kumpulan permission yang dikelompokkan per aksi.
//
// Dipakai route/route.go supaya nama permission tidak ditulis inline di
// sana. Satu aksi = satu kelompok; kelompok dengan dua permission
// (":any" dan ":self") berarti salah satunya cukup.
//
// Catatan nama: PermissionSet (tipe di atas) sudah dipakai untuk himpunan
// permission milik satu user, jadi tipe untuk katalog aksi memakai nama
// PermissionActions.
type PermissionActions struct {
	CourseEnroll   []string
	CourseManage   []string
	QueueMonitor   []string
	PrestasiList   []string
	PrestasiRead   []string
	PrestasiCreate []string
	StudentList    []string
	StudentRead    []string
	StudentCreate  []string
	StudentUpdate  []string
	StudentDelete  []string
	UserList       []string
}

// NewPermissionSet menyusun katalog permission per aksi.
//
// Daftar permission di sini adalah kontrak: nilainya TIDAK boleh berubah
// tanpa keputusan di docs/DECISIONS.md, karena middleware membandingkan
// string ini dengan isi tabel permissions.
func NewPermissionSet() PermissionActions {
	return PermissionActions{
		CourseEnroll:   []string{"course:enroll"},
		CourseManage:   []string{"course:manage"},
		QueueMonitor:   []string{"queue:monitor"},
		PrestasiList:   []string{"prestasi:list"},
		PrestasiRead:   []string{"prestasi:read:any", "prestasi:read:self"},
		PrestasiCreate: []string{"prestasi:create"},
		StudentList:    []string{"student:list"},
		StudentRead:    []string{"student:read:any", "student:read:self"},
		StudentCreate:  []string{"student:create"},
		StudentUpdate:  []string{"student:update:any", "student:update:self"},
		StudentDelete:  []string{"student:delete"},
		UserList:       []string{"user:list"},
	}
}
