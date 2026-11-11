package config

import (
	"strings"
	"testing"
)

func TestNewAppReadsEnvironment(t *testing.T) {
	t.Setenv("APP_PORT", "8080")
	t.Setenv("JWT_SECRET", "secret-uji")
	t.Setenv("DB_USER", "postgres")
	t.Setenv("DB_PASSWORD", "rahasia")
	t.Setenv("DB_HOST", "127.0.0.1")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_NAME", "praktikum_backend_test")
	t.Setenv("DB_SSLMODE", "disable")
	t.Setenv("REDIS_ADDR", "127.0.0.1:6380")
	t.Setenv("MONGO_URI", "mongodb://127.0.0.1:27018")
	t.Setenv("MONGO_DATABASE", "db_uji")
	t.Setenv("SWAGGER_ENABLED", "true")

	app := NewApp()

	if app.Port != "8080" {
		t.Errorf("port harus 8080, dapat %q", app.Port)
	}
	if string(app.JWTSecret) != "secret-uji" {
		t.Errorf("JWT secret harus dari env, dapat %q", string(app.JWTSecret))
	}
	wantDSN := "postgres://postgres:rahasia@127.0.0.1:5433/praktikum_backend_test?sslmode=disable"
	if app.DatabaseDSN != wantDSN {
		t.Errorf("DSN harus %q, dapat %q", wantDSN, app.DatabaseDSN)
	}
	if app.Redis.Addr != "127.0.0.1:6380" {
		t.Errorf("alamat Redis harus dari env, dapat %q", app.Redis.Addr)
	}
	if app.MongoURI != "mongodb://127.0.0.1:27018" {
		t.Errorf("URI Mongo harus dari env, dapat %q", app.MongoURI)
	}
	if app.MongoDatabase != "db_uji" {
		t.Errorf("database Mongo harus dari env, dapat %q", app.MongoDatabase)
	}
	if !app.EnableSwagger {
		t.Error("SWAGGER_ENABLED=true harus mengaktifkan EnableSwagger")
	}
}

func TestNewAppUsesDefaults(t *testing.T) {
	for _, key := range []string{
		"APP_PORT", "JWT_SECRET", "REDIS_ADDR", "MONGO_URI", "MONGO_DATABASE", "SWAGGER_ENABLED",
	} {
		t.Setenv(key, "")
	}

	app := NewApp()

	if app.Port != DefaultPort {
		t.Errorf("port default harus %q, dapat %q", DefaultPort, app.Port)
	}
	if string(app.JWTSecret) != DefaultJWTSecret {
		t.Errorf("JWT secret default harus %q, dapat %q", DefaultJWTSecret, string(app.JWTSecret))
	}
	if app.MongoURI != DefaultMongoURI {
		t.Errorf("URI Mongo default harus %q, dapat %q", DefaultMongoURI, app.MongoURI)
	}
	if app.MongoDatabase != DefaultMongoDatabase {
		t.Errorf("database Mongo default harus %q, dapat %q", DefaultMongoDatabase, app.MongoDatabase)
	}
	if app.EnableSwagger {
		t.Error("Swagger harus mati secara default")
	}
	if !strings.HasPrefix(app.DatabaseDSN, "postgres://") {
		t.Errorf("DSN harus diawali postgres://, dapat %q", app.DatabaseDSN)
	}
}

func TestGetBool(t *testing.T) {
	cases := []struct {
		value string
		want  bool
	}{
		{value: "true", want: true},
		{value: "TRUE", want: true},
		{value: "1", want: true},
		{value: "false", want: false},
		{value: "0", want: false},
		{value: "", want: true}, // fallback dipakai
		{value: "bukan-bool", want: true},
	}
	for _, tc := range cases {
		t.Setenv("UJI_BOOL", tc.value)
		got := GetBool("UJI_BOOL", true)
		if got != tc.want {
			t.Errorf("GetBool(%q) harus %v, dapat %v", tc.value, tc.want, got)
		}
	}

	t.Setenv("UJI_BOOL", "")
	if GetBool("UJI_BOOL", false) {
		t.Error("nilai kosong harus memakai fallback false")
	}
}
