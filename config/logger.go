package config

import "log"

// Logger adalah logger bawaan Go yang dipakai seluruh aplikasi.
// Pakai log.Println untuk pesan biasa, log.Fatalf untuk error fatal.
var Logger = log.Default()
