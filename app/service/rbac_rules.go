package service

// RBAC rules murni: tidak menyentuh fiber.Ctx, tidak menyentuh database,
// tidak mengetahui HTTP. Dipakai oleh handler/middleware/service lain,
// dan bisa diuji via `go test` tanpa server.

// HasPermission memeriksa apakah himpunan perms menyertakan permission yang
// diminta. Policy fail closed: perms kosong/nil -> tolak.
func HasPermission(perms map[string]bool, required string) bool {
	if perms == nil {
		return false
	}
	return perms[required]
}

// HasAnyPermissionFromGroups menerima daftar kandidat permission (biasanya
// ["resource:read:any", "resource:read:self"]). True bila user memiliki
// SALAH SATU. Bentuk paling umum: middleware pasang dua permission, satu
// global, satu self; handler/service jatuh ke ownership check bila hanya
// punya :self.
func HasAnyPermissionFromGroups(perms map[string]bool, candidates []string) bool {
	if perms == nil {
		return false
	}
	for _, c := range candidates {
		if perms[c] {
			return true
		}
	}
	return false
}

// RequireOwnerOrAdmin memutuskan apakah aktor boleh menyentuh record tertentu.
// - ownerID <= 0                       -> tolak (record tanpa pemilik; admin
//                                         bypass sudah disediakan di :any).
// - actorID <= 0                       -> tolak (fail closed).
// - actorID == ownerID                 -> boleh.
// Catatan: middleware sudah membatasi domain; fungsi ini hanya menjawab
// "kalau boleh domain, siapa pemiliknya?".
func RequireOwnerOrAdmin(actorID, ownerID int) bool {
	if actorID <= 0 || ownerID <= 0 {
		return false
	}
	return actorID == ownerID
}

// IsAdminOnlyPermission digunakan service yang ingin menolak user siapapun
// yang bukan admin/staff (mis. tambah prestasi untuk student_id lain tanpa
// lewat path :any). Pemeriksaan dilakukan terhadap permission spesifik,
// BUKAN gabungan longgar.
//
// Helper lain dapat ditambah dengan nama yang jelas (CanCreatePrestasiFor,
// dsb.) bila ada kebutuhan serupa.
func IsAdminOnlyPermission(perms map[string]bool, perm string) bool {
	if perms == nil {
		return false
	}
	return perms[perm]
}
