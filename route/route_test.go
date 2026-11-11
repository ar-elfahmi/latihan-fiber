package route

import (
	"sort"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// wantRoutes adalah kontrak endpoint yang tidak boleh berubah selama
// refactor wiring (lihat docs/API.md). Daftar ini mengunci method + path.
var wantRoutes = []string{
	"GET /api/v1/health",
	"POST /api/v1/auth/login",
	"POST /api/v1/auth/register",
	"POST /api/v1/auth/refresh",
	"POST /api/v1/auth/logout",
	"GET /api/v1/auth/me",
	"GET /api/v1/courses",
	"GET /api/v1/courses/mine",
	"GET /api/v1/courses/:id",
	"POST /api/v1/courses/:id/enroll",
	"DELETE /api/v1/courses/:id/enroll",
	"GET /api/v1/courses/:id/reviews",
	"POST /api/v1/courses/:id/reviews",
	"POST /api/v1/enrollments/swap",
	"PATCH /api/v1/courses/:id",
	"GET /api/v1/notifications/mine",
	"GET /api/v1/admin/queue",
	"GET /api/v1/prestasi",
	"GET /api/v1/students/",
	"POST /api/v1/students/",
	"GET /api/v1/students/:id",
	"PUT /api/v1/students/:id",
	"PATCH /api/v1/students/:id",
	"DELETE /api/v1/students/:id",
	"GET /api/v1/students/:id/prestasi",
	"POST /api/v1/students/:id/prestasi",
	"GET /api/v1/users",
	"GET /api/v1/users/:id",
	"GET /api/v1/students/:id/portfolio",
	"PUT /api/v1/students/:id/portfolio",
	"POST /api/v1/students/:id/portfolio/projects",
	"DELETE /api/v1/students/:id/portfolio/projects/:projectId",
	"GET /api/v1/portfolios/stats/technologies",
}

// TestSetupRegistersContractRoutes memastikan daftar endpoint tetap sama
// setelah refactor wiring (S-001). Dependency dikosongkan karena yang
// diuji hanya peta rute, bukan perilakunya.
func TestSetupRegistersContractRoutes(t *testing.T) {
	app := fiber.New()
	Setup(app, Dependencies{})

	var got []string
	for _, r := range app.GetRoutes() {
		if r.Method == "" || r.Path == "" {
			continue
		}
		// "/api/v1" sendiri adalah entri group (middleware), bukan endpoint.
		if r.Path == "/api/v1" {
			continue
		}
		switch r.Method {
		case fiber.MethodGet, fiber.MethodPost, fiber.MethodPut, fiber.MethodPatch, fiber.MethodDelete:
			got = append(got, r.Method+" "+r.Path)
		}
	}
	sort.Strings(got)
	sort.Strings(wantRoutes)

	if len(got) != len(wantRoutes) {
		t.Fatalf("jumlah endpoint berubah: dapat %d, harus %d\ndidapat:\n%s", len(got), len(wantRoutes), strings.Join(got, "\n"))
	}
	for i := range got {
		if got[i] != wantRoutes[i] {
			t.Errorf("endpoint ke-%d berbeda: dapat %q, harus %q", i, got[i], wantRoutes[i])
		}
	}
}
