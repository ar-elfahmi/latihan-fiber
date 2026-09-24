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
