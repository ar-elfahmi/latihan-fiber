package service

// Services adalah struct yang dipakai main.go untuk meneruskan semua service
// ke route.Setup tanpa main.go harus tahu urutan param.
//
// Setiap service memegang repository lewat **interface** repository
// (lihat app/repository/interfaces.go), bukan struct konkret. Dengan begitu
// service bisa diuji memakai fake repository tanpa database, dan implementasi
// PostgreSQL bisa diganti tanpa mengubah handler.
type Services struct {
	Student      StudentService
	Prestasi     PrestasiService
	Auth         AuthService
	Course       CourseService
	User         UserService
	Portfolio    PortfolioService
	Notification NotificationService
}
