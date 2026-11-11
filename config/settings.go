package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hibiken/asynq"
)

// Nilai default konfigurasi. Semua default hidup di config supaya
// tidak ada lagi pembacaan environment variable di luar package ini.
const (
	// DefaultPort adalah port HTTP bila APP_PORT tidak diisi.
	DefaultPort = "3000"
	// DefaultJWTSecret adalah kunci tanda tangan token bila JWT_SECRET
	// tidak diisi. Password bawaan ini hanya untuk pengembangan lokal;
	// produksi wajib menyetel JWT_SECRET sendiri.
	DefaultJWTSecret = "rahasia-latihan"
	// DefaultMongoURI adalah URI MongoDB bila MONGO_URI tidak diisi.
	DefaultMongoURI = "mongodb://localhost:27017"
	// DefaultMongoDatabase adalah database MongoDB bila MONGO_DATABASE
	// tidak diisi.
	DefaultMongoDatabase = "praktikum_backend"
)

// App adalah kumpulan konfigurasi aplikasi. Dibaca satu kali lewat
// NewApp lalu diteruskan ke lapisan bawah (route, service, worker)
// sehingga hanya ada satu sumber konfigurasi.
//
// Swagger: EnableSwagger dibaca dari SWAGGER_ENABLED dan dipakai saat
// UI Swagger dipasang (slice S-007).
type App struct {
	Port          string
	JWTSecret     []byte
	DatabaseDSN   string
	Redis         asynq.RedisClientOpt
	MongoURI      string
	MongoDatabase string
	EnableSwagger bool
}

// NewApp menyusun konfigurasi aplikasi dari environment variable.
// Nilai yang tidak diisi memakai default pada file ini.
func NewApp() App {
	return App{
		Port:          GetWithDefault("APP_PORT", DefaultPort),
		JWTSecret:     []byte(GetWithDefault("JWT_SECRET", DefaultJWTSecret)),
		DatabaseDSN:   DatabaseDSN(),
		Redis:         RedisOpt(),
		MongoURI:      GetWithDefault("MONGO_URI", DefaultMongoURI),
		MongoDatabase: GetWithDefault("MONGO_DATABASE", DefaultMongoDatabase),
		EnableSwagger: GetBool("SWAGGER_ENABLED", false),
	}
}

// DatabaseDSN menyusun DSN PostgreSQL dari environment variable.
// Nilai kosong tidak dipatching: koneksi gagal dengan pesan jelas
// bila kredensial belum diisi.
func DatabaseDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		Get("DB_USER"),
		Get("DB_PASSWORD"),
		Get("DB_HOST"),
		Get("DB_PORT"),
		Get("DB_NAME"),
		Get("DB_SSLMODE"),
	)
}

// GetBool membaca variabel environment sebagai boolean. Nilai yang
// dianggap benar: "1", "t", "true" (huruf besar/kecil). Bila variabel
// kosong atau tidak bisa dibaca, kembalikan fallback.
func GetBool(key string, fallback bool) bool {
	value := strings.TrimSpace(Get(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
