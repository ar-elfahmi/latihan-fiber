package service

// Services adalah struct yang dipakai main.go untuk meneruskan semua service
// ke route.Setup tanpa main.go harus tahu urutan param.
type Services struct {
	Student   StudentService
	Prestasi  PrestasiService
	Auth      AuthService
	User      UserService
	Portfolio PortfolioService
}
