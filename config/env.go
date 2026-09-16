package config

import (
	"os"
	"strings"
)

// Get membaca variabel environment. Mengembalikan string kosong bila tidak ada.
func Get(key string) string {
	return os.Getenv(key)
}

// GetWithDefault membaca variabel environment; bila kosong pakai nilai default.
func GetWithDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// LoadEnv membaca file .env (baris "KEY=value") lalu menaruhnya ke
// environment variable, TANPA menimpa yang sudah ada di sistem.
// Dipanggil sekali di main.go sebelum config.Get dipakai.
func LoadEnv() {
	data, err := os.ReadFile(".env")
	if err != nil {
		return // tidak ada .env: anggap konfigurasi datang dari environment
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"`)
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}
